// Package middleware holds cross-cutting HTTP middleware.
package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/alginugraha/monify/internal/domain"
)

type ctxKey int

const userKey ctxKey = 0

// UserProvider resolves a user id (from the session) to a domain user.
type UserProvider interface {
	ByID(ctx context.Context, id string) (domain.User, error)
}

// SessionReader extracts the authenticated user id from a request.
type SessionReader interface {
	UserID(r *http.Request) (string, bool)
}

// RequireAuth redirects unauthenticated requests to /login and stores the
// resolved user in the request context for downstream handlers.
func RequireAuth(sess SessionReader, users UserProvider) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id, ok := sess.UserID(r)
			if !ok {
				redirectToLogin(w, r)
				return
			}
			u, err := users.ByID(r.Context(), id)
			if err != nil {
				redirectToLogin(w, r)
				return
			}
			ctx := context.WithValue(r.Context(), userKey, u)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// CurrentUser returns the user placed in context by RequireAuth.
func CurrentUser(ctx context.Context) (domain.User, bool) {
	u, ok := ctx.Value(userKey).(domain.User)
	return u, ok
}

// APIKeyAuthenticator resolves a bearer token to its owning user.
type APIKeyAuthenticator interface {
	Authenticate(ctx context.Context, token string) (domain.User, error)
}

// RequireAPIKey authenticates requests via "Authorization: Bearer <token>",
// for external clients that can't hold a browser session cookie. On failure
// it responds with a JSON error rather than redirecting to /login.
func RequireAPIKey(keys APIKeyAuthenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := bearerToken(r)
			if token == "" {
				writeAPIAuthError(w, "missing bearer token")
				return
			}
			u, err := keys.Authenticate(r.Context(), token)
			if err != nil {
				writeAPIAuthError(w, "invalid or revoked api key")
				return
			}
			ctx := context.WithValue(r.Context(), userKey, u)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func bearerToken(r *http.Request) string {
	const p = "Bearer "
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, p) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(h, p))
}

func writeAPIAuthError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// CORS allows cross-origin calls to the token-authenticated API. This is
// safe to leave permissive: bearer tokens are sent explicitly by the caller
// (never carried implicitly like cookies), so opening CORS here cannot be
// used to ride a browser's existing session on the cookie-based app.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func redirectToLogin(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/login")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
