package handler

import (
	"context"
	"net/http"

	"github.com/alginugraha/monify/internal/domain"
	"github.com/alginugraha/monify/internal/http/render"
	"github.com/alginugraha/monify/internal/service"
	"github.com/go-chi/chi/v5"
)

type Categories struct {
	rnd *render.Renderer
	svc *service.CategoryService
}

func NewCategories(rnd *render.Renderer, svc *service.CategoryService) *Categories {
	return &Categories{rnd: rnd, svc: svc}
}

func (h *Categories) Routes(r chi.Router) {
	r.Get("/categories", h.List)
	r.Post("/categories", h.Create)
	r.Get("/categories/{id}/edit", h.EditRow)
	r.Get("/categories/{id}/row", h.Row)
	r.Put("/categories/{id}", h.Update)
	r.Delete("/categories/{id}", h.Delete)
}

// catView pairs a category with the list of possible parents for its <select>.
type catView struct {
	domain.Category
	Parents []domain.Category
}

// SelectedParent returns the parent id as a plain string ("" when top-level),
// for comparison in the edit template.
func (v catView) SelectedParent() string {
	if v.ParentID == nil {
		return ""
	}
	return *v.ParentID
}

func (h *Categories) List(w http.ResponseWriter, r *http.Request) {
	cats, err := h.svc.List(r.Context())
	if err != nil {
		httpError(w, err)
		return
	}
	h.rnd.Page(w, http.StatusOK, "categories.html", map[string]any{
		"Auth": true, "Categories": cats, "Parents": topLevel(cats),
	})
}

func (h *Categories) Create(w http.ResponseWriter, r *http.Request) {
	in := parseCategoryForm(r)
	cat, err := h.svc.Create(r.Context(), in)
	if err != nil {
		httpError(w, err)
		return
	}
	h.rnd.Partial(w, http.StatusOK, "category-row", cat)
}

func (h *Categories) EditRow(w http.ResponseWriter, r *http.Request) {
	one, err := findCategory(r.Context(), h.svc, chi.URLParam(r, "id"))
	if err != nil {
		httpError(w, err)
		return
	}
	all, _ := h.svc.List(r.Context())
	h.rnd.Partial(w, http.StatusOK, "category-row-edit", catView{Category: one, Parents: topLevel(all)})
}

func (h *Categories) Row(w http.ResponseWriter, r *http.Request) {
	one, err := findCategory(r.Context(), h.svc, chi.URLParam(r, "id"))
	if err != nil {
		httpError(w, err)
		return
	}
	h.rnd.Partial(w, http.StatusOK, "category-row", one)
}

func (h *Categories) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Update(r.Context(), id, parseCategoryForm(r)); err != nil {
		httpError(w, err)
		return
	}
	one, err := findCategory(r.Context(), h.svc, id)
	if err != nil {
		httpError(w, err)
		return
	}
	h.rnd.Partial(w, http.StatusOK, "category-row", one)
}

func (h *Categories) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		httpError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func parseCategoryForm(r *http.Request) service.CategoryInput {
	_ = r.ParseForm()
	return service.CategoryInput{
		Name:     r.PostFormValue("name"),
		Kind:     r.PostFormValue("kind"),
		ParentID: r.PostFormValue("parent_id"),
	}
}

func topLevel(cats []domain.Category) []domain.Category {
	var out []domain.Category
	for _, c := range cats {
		if c.ParentID == nil {
			out = append(out, c)
		}
	}
	return out
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
