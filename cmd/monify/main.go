package main

import (
	"bufio"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/alginugraha/monify/internal/config"
	httpx "github.com/alginugraha/monify/internal/http"
	"github.com/alginugraha/monify/internal/http/render"
	"github.com/alginugraha/monify/internal/platform/session"
	"github.com/alginugraha/monify/internal/postgres"
	"github.com/alginugraha/monify/internal/service"
	"github.com/alginugraha/monify/migrations"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"golang.org/x/term"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "seed-user":
			if err := seedUser(logger); err != nil {
				logger.Error("seed-user failed", "err", err)
				os.Exit(1)
			}
			return
		case "migrate":
			if err := migrateDB(logger); err != nil {
				logger.Error("migrate failed", "err", err)
				os.Exit(1)
			}
			return
		case "apikey":
			if err := apiKeyCmd(logger, os.Args[2:]); err != nil {
				logger.Error("apikey failed", "err", err)
				os.Exit(1)
			}
			return
		}
	}

	if err := run(logger); err != nil {
		logger.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	logger.Info("connected to database")

	rnd, err := render.New()
	if err != nil {
		return err
	}

	deps := httpx.Deps{
		Pool:    pool,
		Session: session.NewManager(cfg.SessionSecret, cfg.SessionSecure),
		Render:  rnd,
	}

	// Background: generate due recurring transactions hourly.
	recurring := service.NewRecurringService(postgres.NewRecurringRepo(pool))
	go runRecurring(ctx, logger, recurring)

	srv := &http.Server{
		Addr:         cfg.Addr,
		Handler:      httpx.NewRouter(deps),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("http server listening", "addr", cfg.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

// runRecurring generates due recurring transactions at startup and every hour
// until the context is cancelled. The repo's advisory lock makes concurrent
// runs safe.
func runRecurring(ctx context.Context, logger *slog.Logger, svc *service.RecurringService) {
	tick := func() {
		runCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		if n, err := svc.Run(runCtx); err != nil {
			logger.Error("recurring run failed", "err", err)
		} else if n > 0 {
			logger.Info("recurring transactions generated", "count", n)
		}
	}

	tick()
	t := time.NewTicker(time.Hour)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			tick()
		}
	}
}

// migrateDB applies all embedded goose migrations against DATABASE_URL.
func migrateDB(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	goose.SetBaseFS(migrations.FS)
	goose.SetLogger(goose.NopLogger())
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	if err := goose.Up(db, "."); err != nil {
		return err
	}
	v, _ := goose.GetDBVersion(db)
	logger.Info("migrations applied", "version", v)
	return nil
}

// seedUser creates the first application user interactively.
func seedUser(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx := context.Background()
	pool, err := postgres.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Email: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	fmt.Print("Password (min 8 chars): ")
	pwBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return err
	}

	auth := service.NewAuthService(postgres.NewUserRepo(pool))
	u, err := auth.CreateUser(ctx, email, string(pwBytes))
	if err != nil {
		return err
	}
	logger.Info("user created", "id", u.ID, "email", u.Email)
	return nil
}

// apiKeyCmd handles `monify apikey create <email> <name>`, minting a new
// external-API bearer token for an existing user. The plaintext token is
// printed once and cannot be recovered afterwards.
func apiKeyCmd(logger *slog.Logger, args []string) error {
	if len(args) < 1 || args[0] != "create" {
		return errors.New(`usage: monify apikey create <email> <name>`)
	}
	if len(args) != 3 {
		return errors.New(`usage: monify apikey create <email> <name>`)
	}
	email, name := args[1], args[2]

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx := context.Background()
	pool, err := postgres.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	users := postgres.NewUserRepo(pool)
	u, err := users.ByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("lookup user %q: %w", email, err)
	}

	keys := service.NewAPIKeyService(postgres.NewAPIKeyRepo(pool), users)
	_, token, err := keys.Create(ctx, u.ID, name)
	if err != nil {
		return err
	}
	logger.Info("api key created", "user", u.Email, "name", name)
	fmt.Println("Token (save it now, it will not be shown again):")
	fmt.Println(token)
	return nil
}
