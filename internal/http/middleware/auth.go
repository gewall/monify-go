// Package middleware holds cross-cutting HTTP middleware.
package middleware

import (
	"context"
	"net/http"

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

func redirectToLogin(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/login")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
