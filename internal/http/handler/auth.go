// Package handler contains the HTTP handlers for each resource.
package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/alginugraha/monify/internal/domain"
	"github.com/alginugraha/monify/internal/http/render"
)

// Authenticator is the auth use case this handler depends on.
type Authenticator interface {
	Authenticate(ctx context.Context, email, password string) (domain.User, error)
}

// SessionWriter manages the session cookie.
type SessionWriter interface {
	Set(w http.ResponseWriter, userID string)
	Clear(w http.ResponseWriter)
	UserID(r *http.Request) (string, bool)
}

type Auth struct {
	rnd  *render.Renderer
	auth Authenticator
	sess SessionWriter
}

func NewAuth(rnd *render.Renderer, auth Authenticator, sess SessionWriter) *Auth {
	return &Auth{rnd: rnd, auth: auth, sess: sess}
}

func (h *Auth) LoginForm(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.sess.UserID(r); ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	h.rnd.Page(w, http.StatusOK, "login.html", map[string]any{"Auth": false})
}

func (h *Auth) Login(w http.ResponseWriter, r *http.Request) {
	email := r.PostFormValue("email")
	password := r.PostFormValue("password")

	u, err := h.auth.Authenticate(r.Context(), email, password)
	if err != nil {
		status := http.StatusUnauthorized
		msg := "Email atau password salah."
		if !errors.Is(err, domain.ErrUnauthorized) {
			status = http.StatusInternalServerError
			msg = "Terjadi kesalahan, coba lagi."
		}
		h.rnd.Page(w, status, "login.html", map[string]any{
			"Auth": false, "Email": email, "Error": msg,
		})
		return
	}

	h.sess.Set(w, u.ID)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Auth) Logout(w http.ResponseWriter, r *http.Request) {
	h.sess.Clear(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
