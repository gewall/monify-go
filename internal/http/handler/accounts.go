package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/alginugraha/monify/internal/domain"
	"github.com/alginugraha/monify/internal/http/render"
	"github.com/alginugraha/monify/internal/platform/money"
	"github.com/alginugraha/monify/internal/service"
	"github.com/go-chi/chi/v5"
)

type Accounts struct {
	rnd *render.Renderer
	svc *service.AccountService
}

func NewAccounts(rnd *render.Renderer, svc *service.AccountService) *Accounts {
	return &Accounts{rnd: rnd, svc: svc}
}

func (h *Accounts) Routes(r chi.Router) {
	r.Get("/accounts", h.List)
	r.Post("/accounts", h.Create)
	r.Get("/accounts/{id}/edit", h.EditRow)
	r.Get("/accounts/{id}/row", h.Row)
	r.Put("/accounts/{id}", h.Update)
	r.Delete("/accounts/{id}", h.Delete)
}

func (h *Accounts) List(w http.ResponseWriter, r *http.Request) {
	accounts, err := h.svc.List(r.Context())
	if err != nil {
		httpError(w, err)
		return
	}
	h.rnd.Page(w, http.StatusOK, "accounts.html", map[string]any{
		"Auth": true, "Accounts": accounts,
	})
}

func (h *Accounts) Create(w http.ResponseWriter, r *http.Request) {
	in, err := parseAccountForm(r)
	if err != nil {
		httpError(w, err)
		return
	}
	acc, err := h.svc.Create(r.Context(), in)
	if err != nil {
		httpError(w, err)
		return
	}
	h.rnd.Partial(w, http.StatusOK, "account-row", acc)
}

func (h *Accounts) EditRow(w http.ResponseWriter, r *http.Request) {
	one, err := findAccount(r.Context(), h.svc, chi.URLParam(r, "id"))
	if err != nil {
		httpError(w, err)
		return
	}
	h.rnd.Partial(w, http.StatusOK, "account-row-edit", one)
}

func (h *Accounts) Row(w http.ResponseWriter, r *http.Request) {
	one, err := findAccount(r.Context(), h.svc, chi.URLParam(r, "id"))
	if err != nil {
		httpError(w, err)
		return
	}
	h.rnd.Partial(w, http.StatusOK, "account-row", one)
}

func (h *Accounts) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	in, err := parseAccountForm(r)
	if err != nil {
		httpError(w, err)
		return
	}
	if err := h.svc.Update(r.Context(), id, in); err != nil {
		httpError(w, err)
		return
	}
	one, err := findAccount(r.Context(), h.svc, id)
	if err != nil {
		httpError(w, err)
		return
	}
	h.rnd.Partial(w, http.StatusOK, "account-row", one)
}

func (h *Accounts) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		httpError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK) // empty body -> htmx removes the row
}

func parseAccountForm(r *http.Request) (service.AccountInput, error) {
	if err := r.ParseForm(); err != nil {
		return service.AccountInput{}, err
	}
	var bal int64
	if v := r.PostFormValue("initial_balance"); v != "" {
		parsed, err := money.Parse(v)
		if err != nil {
			return service.AccountInput{}, errors.Join(domain.ErrInvalid, err)
		}
		bal = parsed
	}
	return service.AccountInput{
		Name:           r.PostFormValue("name"),
		Type:           r.PostFormValue("type"),
		InitialBalance: bal,
		Archived:       r.PostFormValue("archived") == "on",
	}, nil
}

// findAccount reloads a single account (with balance) from the list.
func findAccount(ctx context.Context, svc *service.AccountService, id string) (domain.Account, error) {
	list, err := svc.List(ctx)
	if err != nil {
		return domain.Account{}, err
	}
	for _, a := range list {
		if a.ID == id {
			return a, nil
		}
	}
	return domain.Account{}, domain.ErrNotFound
}
