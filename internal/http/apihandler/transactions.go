package apihandler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/alginugraha/monify/internal/domain"
	"github.com/alginugraha/monify/internal/service"
	"github.com/go-chi/chi/v5"
)

const (
	txDefaultLimit = 25
	txMaxLimit     = 200
)

type Transactions struct {
	svc *service.TransactionService
}

func NewTransactions(svc *service.TransactionService) *Transactions {
	return &Transactions{svc: svc}
}

func (h *Transactions) Routes(r chi.Router) {
	r.Get("/transactions", h.List)
	r.Post("/transactions", h.Create)
	r.Get("/transactions/{id}", h.Get)
	r.Put("/transactions/{id}", h.Update)
	r.Delete("/transactions/{id}", h.Delete)
}

// txBody is the JSON create/update payload. AmountMinor is always positive;
// OccurredAt is a "YYYY-MM-DD" date. ToAccountID is required for transfers,
// CategoryID is ignored for transfers.
type txBody struct {
	Kind        string `json:"kind"`
	AmountMinor int64  `json:"amount_minor"`
	AccountID   string `json:"account_id"`
	ToAccountID string `json:"to_account_id,omitempty"`
	CategoryID  string `json:"category_id,omitempty"`
	OccurredAt  string `json:"occurred_at"`
	Note        string `json:"note,omitempty"`
}

type txListResponse struct {
	Transactions []domain.Transaction `json:"transactions"`
	Total        int                  `json:"total"`
	Limit        int                  `json:"limit"`
	Offset       int                  `json:"offset"`
}

func (h *Transactions) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := domain.TxFilter{
		AccountID:  q.Get("account_id"),
		CategoryID: q.Get("category_id"),
		Kind:       q.Get("kind"),
		Limit:      queryInt(q, "limit", txDefaultLimit, 1, txMaxLimit),
		Offset:     queryInt(q, "offset", 0, 0, 1<<31-1),
	}
	if v := q.Get("from"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			f.From = &t
		}
	}
	if v := q.Get("to"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			f.To = &t
		}
	}

	items, total, err := h.svc.List(r.Context(), f)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, txListResponse{
		Transactions: items, Total: total, Limit: f.Limit, Offset: f.Offset,
	})
}

func (h *Transactions) Get(w http.ResponseWriter, r *http.Request) {
	one, err := h.svc.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, one)
}

func (h *Transactions) Create(w http.ResponseWriter, r *http.Request) {
	in, err := parseTxBody(r)
	if err != nil {
		writeError(w, err)
		return
	}
	tx, err := h.svc.Create(r.Context(), in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, tx)
}

func (h *Transactions) Update(w http.ResponseWriter, r *http.Request) {
	in, err := parseTxBody(r)
	if err != nil {
		writeError(w, err)
		return
	}
	tx, err := h.svc.Update(r.Context(), chi.URLParam(r, "id"), in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, tx)
}

func (h *Transactions) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func parseTxBody(r *http.Request) (service.TxInput, error) {
	var body txBody
	if err := decodeJSON(r, &body); err != nil {
		return service.TxInput{}, err
	}
	occ, err := time.Parse("2006-01-02", body.OccurredAt)
	if err != nil {
		return service.TxInput{}, errJoinInvalid(err)
	}
	return service.TxInput{
		Kind:        body.Kind,
		AmountMinor: body.AmountMinor,
		AccountID:   body.AccountID,
		ToAccountID: body.ToAccountID,
		CategoryID:  body.CategoryID,
		OccurredAt:  occ,
		Note:        body.Note,
	}, nil
}

func queryInt(q map[string][]string, key string, def, min, max int) int {
	v, ok := q[key]
	if !ok || len(v) == 0 || v[0] == "" {
		return def
	}
	n, err := strconv.Atoi(v[0])
	if err != nil {
		return def
	}
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}
