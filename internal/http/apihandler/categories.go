package apihandler

import (
	"context"
	"net/http"

	"github.com/alginugraha/monify/internal/domain"
	"github.com/alginugraha/monify/internal/service"
	"github.com/go-chi/chi/v5"
)

type Categories struct {
	svc *service.CategoryService
}

func NewCategories(svc *service.CategoryService) *Categories {
	return &Categories{svc: svc}
}

func (h *Categories) Routes(r chi.Router) {
	r.Get("/categories", h.List)
	r.Post("/categories", h.Create)
	r.Get("/categories/{id}", h.Get)
	r.Put("/categories/{id}", h.Update)
	r.Delete("/categories/{id}", h.Delete)
}

type categoryBody struct {
	Name     string `json:"name"`
	Kind     string `json:"kind"`
	ParentID string `json:"parent_id"`
}

func (h *Categories) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.List(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *Categories) Get(w http.ResponseWriter, r *http.Request) {
	one, err := findCategory(r.Context(), h.svc, chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, one)
}

func (h *Categories) Create(w http.ResponseWriter, r *http.Request) {
	var body categoryBody
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, err)
		return
	}
	cat, err := h.svc.Create(r.Context(), service.CategoryInput{
		Name: body.Name, Kind: body.Kind, ParentID: body.ParentID,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, cat)
}

func (h *Categories) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body categoryBody
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, err)
		return
	}
	if err := h.svc.Update(r.Context(), id, service.CategoryInput{
		Name: body.Name, Kind: body.Kind, ParentID: body.ParentID,
	}); err != nil {
		writeError(w, err)
		return
	}
	one, err := findCategory(r.Context(), h.svc, id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, one)
}

func (h *Categories) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func findCategory(ctx context.Context, svc *service.CategoryService, id string) (domain.Category, error) {
	list, err := svc.List(ctx)
	if err != nil {
		return domain.Category{}, err
	}
	for _, c := range list {
		if c.ID == id {
			return c, nil
		}
	}
	return domain.Category{}, domain.ErrNotFound
}
