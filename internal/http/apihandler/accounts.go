package apihandler

import (
	"context"
	"net/http"

	"github.com/alginugraha/monify/internal/domain"
	"github.com/alginugraha/monify/internal/service"
	"github.com/go-chi/chi/v5"
)

type Accounts struct {
	svc *service.AccountService
}

func NewAccounts(svc *service.AccountService) *Accounts {
	return &Accounts{svc: svc}
}

func (h *Accounts) Routes(r chi.Router) {
	r.Get("/accounts", h.List)
	r.Post("/accounts", h.Create)
	r.Get("/accounts/{id}", h.Get)
	r.Put("/accounts/{id}", h.Update)
	r.Delete("/accounts/{id}", h.Delete)
}

// accountBody is the JSON create/update payload. InitialBalanceMinor and all
// amounts across this API are integer minor units (cents) — never floats.
type accountBody struct {
	Name                string `json:"name"`
	Type                string `json:"type"`
	InitialBalanceMinor int64  `json:"initial_balance_minor"`
	Archived            bool   `json:"archived"`
}

func (h *Accounts) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.List(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *Accounts) Get(w http.ResponseWriter, r *http.Request) {
	one, err := findAccount(r.Context(), h.svc, chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, one)
}

func (h *Accounts) Create(w http.ResponseWriter, r *http.Request) {
	var body accountBody
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, err)
		return
	}
	acc, err := h.svc.Create(r.Context(), service.AccountInput{
		Name: body.Name, Type: body.Type,
		InitialBalance: body.InitialBalanceMinor, Archived: body.Archived,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, acc)
}

func (h *Accounts) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body accountBody
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, err)
		return
	}
	if err := h.svc.Update(r.Context(), id, service.AccountInput{
		Name: body.Name, Type: body.Type,
		InitialBalance: body.InitialBalanceMinor, Archived: body.Archived,
	}); err != nil {
		writeError(w, err)
		return
	}
	one, err := findAccount(r.Context(), h.svc, id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, one)
}

func (h *Accounts) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

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
