// Package http wires HTTP routing for the Monify application.
package http

import (
	"context"
	"encoding/json"
	"io/fs"
	"net/http"
	"time"

	"github.com/alginugraha/monify/internal/http/handler"
	appmw "github.com/alginugraha/monify/internal/http/middleware"
	"github.com/alginugraha/monify/internal/http/render"
	"github.com/alginugraha/monify/internal/platform/session"
	"github.com/alginugraha/monify/internal/postgres"
	"github.com/alginugraha/monify/internal/service"
	"github.com/alginugraha/monify/web"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Deps carries the primitives the router needs; services and handlers are wired
// from these here so main stays thin.
type Deps struct {
	Pool    *pgxpool.Pool
	Session *session.Manager
	Render  *render.Renderer
}

// NewRouter builds the application's root HTTP handler.
func NewRouter(d Deps) http.Handler {
	accountRepo := postgres.NewAccountRepo(d.Pool)
	categoryRepo := postgres.NewCategoryRepo(d.Pool)
	txRepo := postgres.NewTransactionRepo(d.Pool)

	authSvc := service.NewAuthService(postgres.NewUserRepo(d.Pool))
	accountSvc := service.NewAccountService(accountRepo)
	categorySvc := service.NewCategoryService(categoryRepo)
	txSvc := service.NewTransactionService(txRepo)
	importSvc := service.NewImportService(accountRepo, categoryRepo, txRepo)
	budgetRepo := postgres.NewBudgetRepo(d.Pool)
	budgetSvc := service.NewBudgetService(budgetRepo)
	reportSvc := service.NewReportService(postgres.NewReportRepo(d.Pool), budgetRepo)
	recurringSvc := service.NewRecurringService(postgres.NewRecurringRepo(d.Pool))
	wishlistSvc := service.NewWishlistService(
		postgres.NewWishlistRepo(d.Pool), postgres.NewReportRepo(d.Pool))

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Get("/healthz", healthz(d.Pool))

	staticFS, _ := fs.Sub(web.FS, "static")
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	authH := handler.NewAuth(d.Render, authSvc, d.Session)
	r.Get("/login", authH.LoginForm)
	r.With(appmw.SameOrigin).Post("/login", authH.Login)
	r.With(appmw.SameOrigin).Post("/logout", authH.Logout)

	r.Group(func(pr chi.Router) {
		pr.Use(appmw.RequireAuth(d.Session, authSvc))
		pr.Use(appmw.SameOrigin)

		handler.NewDashboard(d.Render, reportSvc, recurringSvc).Routes(pr)
		handler.NewAccounts(d.Render, accountSvc).Routes(pr)
		handler.NewCategories(d.Render, categorySvc).Routes(pr)
		handler.NewTransactions(d.Render, txSvc, accountSvc, categorySvc).Routes(pr)
		handler.NewBudgets(d.Render, budgetSvc).Routes(pr)
		handler.NewRecurring(d.Render, recurringSvc, accountSvc, categorySvc).Routes(pr)
		handler.NewWishlist(d.Render, wishlistSvc).Routes(pr)
		handler.NewImportExport(d.Render, importSvc, txSvc).Routes(pr)
		handler.NewReports(d.Render, reportSvc).Routes(pr)
	})

	return r
}

func healthz(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx, cancel := context.WithTimeout(req.Context(), 2*time.Second)
		defer cancel()

		status := "ok"
		code := http.StatusOK
		if err := pool.Ping(ctx); err != nil {
			status = "unavailable"
			code = http.StatusServiceUnavailable
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": status})
	}
}
