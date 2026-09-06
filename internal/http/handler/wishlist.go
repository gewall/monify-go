package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/alginugraha/monify/internal/domain"
	"github.com/alginugraha/monify/internal/http/render"
	"github.com/alginugraha/monify/internal/platform/money"
	"github.com/alginugraha/monify/internal/service"
	"github.com/go-chi/chi/v5"
)

type Wishlist struct {
	rnd *render.Renderer
	svc *service.WishlistService
}

func NewWishlist(rnd *render.Renderer, svc *service.WishlistService) *Wishlist {
	return &Wishlist{rnd: rnd, svc: svc}
}

func (h *Wishlist) Routes(r chi.Router) {
	r.Get("/wishlist", h.Index)
	r.Get("/wishlist/list", h.ListPartial)
	r.Post("/wishlist", h.Create)
	r.Get("/wishlist/{id}/edit", h.EditCard)
	r.Put("/wishlist/{id}", h.Update)
	r.Post("/wishlist/{id}/save", h.Save)
	r.Post("/wishlist/{id}/status", h.Status)
	r.Delete("/wishlist/{id}", h.Delete)
}

func sequentialMode(r *http.Request) bool {
	return r.FormValue("mode") == "sequential"
}

func (h *Wishlist) Index(w http.ResponseWriter, r *http.Request) {
	seq := sequentialMode(r)
	view, err := h.svc.Overview(r.Context(), seq)
	if err != nil {
		httpError(w, err)
		return
	}
	h.rnd.Page(w, http.StatusOK, "wishlist.html", map[string]any{
		"Auth": true,
		"View": view,
		"Mode": modeStr(seq),
	})
}

func (h *Wishlist) ListPartial(w http.ResponseWriter, r *http.Request) {
	h.renderList(w, r)
}

func (h *Wishlist) renderList(w http.ResponseWriter, r *http.Request) {
	seq := sequentialMode(r)
	view, err := h.svc.Overview(r.Context(), seq)
	if err != nil {
		httpError(w, err)
		return
	}
	h.rnd.Partial(w, http.StatusOK, "wish-list", map[string]any{"View": view, "Mode": modeStr(seq)})
}

func (h *Wishlist) Create(w http.ResponseWriter, r *http.Request) {
	in, err := parseWishForm(r)
	if err != nil {
		httpError(w, err)
		return
	}
	if _, err := h.svc.Create(r.Context(), in); err != nil {
		httpError(w, err)
		return
	}
	h.renderList(w, r)
}

func (h *Wishlist) Update(w http.ResponseWriter, r *http.Request) {
	in, err := parseWishForm(r)
	if err != nil {
		httpError(w, err)
		return
	}
	if _, err := h.svc.Update(r.Context(), chi.URLParam(r, "id"), in); err != nil {
		httpError(w, err)
		return
	}
	h.renderList(w, r)
}

func (h *Wishlist) Save(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		httpError(w, err)
		return
	}
	delta, err := money.Parse(r.PostFormValue("amount"))
	if err != nil {
		httpError(w, errJoinInvalid(err))
		return
	}
	if _, err := h.svc.AddSaving(r.Context(), chi.URLParam(r, "id"), delta); err != nil {
		httpError(w, err)
		return
	}
	h.renderList(w, r)
}

func (h *Wishlist) Status(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		httpError(w, err)
		return
	}
	if _, err := h.svc.SetStatus(r.Context(), chi.URLParam(r, "id"),
		domain.WishStatus(r.PostFormValue("status"))); err != nil {
		httpError(w, err)
		return
	}
	h.renderList(w, r)
}

func (h *Wishlist) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		httpError(w, err)
		return
	}
	h.renderList(w, r)
}

func (h *Wishlist) EditCard(w http.ResponseWriter, r *http.Request) {
	item, err := h.svc.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httpError(w, err)
		return
	}
	h.rnd.Partial(w, http.StatusOK, "wish-card-edit", map[string]any{
		"Item": item, "Mode": modeStr(sequentialMode(r)),
	})
}

func parseWishForm(r *http.Request) (service.WishInput, error) {
	if err := r.ParseForm(); err != nil {
		return service.WishInput{}, err
	}
	target, err := money.Parse(r.PostFormValue("target"))
	if err != nil {
		return service.WishInput{}, errJoinInvalid(err)
	}
	var saved int64
	if v := r.PostFormValue("saved"); v != "" {
		saved, err = money.Parse(v)
		if err != nil {
			return service.WishInput{}, errJoinInvalid(err)
		}
	}
	prio, _ := strconv.Atoi(r.PostFormValue("priority"))
	in := service.WishInput{
		Name:        r.PostFormValue("name"),
		TargetMinor: target,
		SavedMinor:  saved,
		Priority:    prio,
		URL:         r.PostFormValue("url"),
		Status:      r.PostFormValue("status"),
	}
	if v := r.PostFormValue("target_date"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			in.TargetDate = &t
		}
	}
	return in, nil
}

func modeStr(seq bool) string {
	if seq {
		return "sequential"
	}
	return "parallel"
}
