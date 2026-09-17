package apihandler

import (
	"net/http"
	"time"

	"github.com/alginugraha/monify/internal/domain"
	"github.com/alginugraha/monify/internal/service"
	"github.com/go-chi/chi/v5"
)

type Wishlist struct {
	svc *service.WishlistService
}

func NewWishlist(svc *service.WishlistService) *Wishlist {
	return &Wishlist{svc: svc}
}

func (h *Wishlist) Routes(r chi.Router) {
	r.Get("/wishlist", h.List)
	r.Post("/wishlist", h.Create)
	r.Get("/wishlist/{id}", h.Get)
	r.Put("/wishlist/{id}", h.Update)
	r.Post("/wishlist/{id}/save", h.Save)
	r.Post("/wishlist/{id}/status", h.Status)
	r.Delete("/wishlist/{id}", h.Delete)
}

type wishBody struct {
	Name        string  `json:"name"`
	TargetMinor int64   `json:"target_minor"`
	SavedMinor  int64   `json:"saved_minor"`
	Priority    int     `json:"priority"`
	URL         string  `json:"url,omitempty"`
	Status      string  `json:"status,omitempty"`
	TargetDate  *string `json:"target_date,omitempty"` // "YYYY-MM-DD"
}

func (b wishBody) toInput() (service.WishInput, error) {
	in := service.WishInput{
		Name: b.Name, TargetMinor: b.TargetMinor, SavedMinor: b.SavedMinor,
		Priority: b.Priority, URL: b.URL, Status: b.Status,
	}
	if b.TargetDate != nil && *b.TargetDate != "" {
		t, err := time.Parse("2006-01-02", *b.TargetDate)
		if err != nil {
			return service.WishInput{}, errJoinInvalid(err)
		}
		in.TargetDate = &t
	}
	return in, nil
}

// sequential controls whether ?mode=sequential estimates cumulative funding
// across items in priority order, vs. each item independently (parallel,
// the default).
func sequential(r *http.Request) bool {
	return r.URL.Query().Get("mode") == "sequential"
}

func (h *Wishlist) List(w http.ResponseWriter, r *http.Request) {
	view, err := h.svc.Overview(r.Context(), sequential(r))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, view)
}

func (h *Wishlist) Get(w http.ResponseWriter, r *http.Request) {
	item, err := h.svc.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *Wishlist) Create(w http.ResponseWriter, r *http.Request) {
	var body wishBody
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, err)
		return
	}
	in, err := body.toInput()
	if err != nil {
		writeError(w, err)
		return
	}
	item, err := h.svc.Create(r.Context(), in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *Wishlist) Update(w http.ResponseWriter, r *http.Request) {
	var body wishBody
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, err)
		return
	}
	in, err := body.toInput()
	if err != nil {
		writeError(w, err)
		return
	}
	item, err := h.svc.Update(r.Context(), chi.URLParam(r, "id"), in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

type saveBody struct {
	AmountMinor int64 `json:"amount_minor"`
}

func (h *Wishlist) Save(w http.ResponseWriter, r *http.Request) {
	var body saveBody
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, err)
		return
	}
	item, err := h.svc.AddSaving(r.Context(), chi.URLParam(r, "id"), body.AmountMinor)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

type statusBody struct {
	Status string `json:"status"`
}

func (h *Wishlist) Status(w http.ResponseWriter, r *http.Request) {
	var body statusBody
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, err)
		return
	}
	item, err := h.svc.SetStatus(r.Context(), chi.URLParam(r, "id"), domain.WishStatus(body.Status))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *Wishlist) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
