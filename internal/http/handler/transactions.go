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

const txPageSize = 25

type Transactions struct {
	rnd  *render.Renderer
	txs  *service.TransactionService
	accs *service.AccountService
	cats *service.CategoryService
}

func NewTransactions(rnd *render.Renderer, txs *service.TransactionService,
	accs *service.AccountService, cats *service.CategoryService) *Transactions {
	return &Transactions{rnd: rnd, txs: txs, accs: accs, cats: cats}
}

func (h *Transactions) Routes(r chi.Router) {
	r.Get("/transactions", h.Index)
	r.Get("/transactions/list", h.List)
	r.Post("/transactions", h.Create)
	r.Get("/transactions/{id}/edit", h.EditRow)
	r.Get("/transactions/{id}/row", h.Row)
	r.Put("/transactions/{id}", h.Update)
	r.Delete("/transactions/{id}", h.Delete)
}

type txPanel struct {
	Txns   []domain.Transaction
	Page   int
	Pages  int
	Total  int
	Filter domain.TxFilter
	FForm  filterForm // raw string values to re-fill the filter form
}

type filterForm struct {
	From, To, AccountID, CategoryID, Kind string
}

func (h *Transactions) Index(w http.ResponseWriter, r *http.Request) {
	panel, err := h.buildPanel(r)
	if err != nil {
		httpError(w, err)
		return
	}
	accs, _ := h.accs.List(r.Context())
	cats, _ := h.cats.List(r.Context())
	h.rnd.Page(w, http.StatusOK, "transactions.html", map[string]any{
		"Auth":       true,
		"Panel":      panel,
		"Accounts":   accs,
		"Categories": cats,
		"Today":      time.Now().Format("2006-01-02"),
	})
}

func (h *Transactions) List(w http.ResponseWriter, r *http.Request) {
	panel, err := h.buildPanel(r)
	if err != nil {
		httpError(w, err)
		return
	}
	h.rnd.Partial(w, http.StatusOK, "tx-panel", panel)
}

func (h *Transactions) Create(w http.ResponseWriter, r *http.Request) {
	in, err := h.parseForm(r)
	if err != nil {
		httpError(w, err)
		return
	}
	if _, err := h.txs.Create(r.Context(), in); err != nil {
		httpError(w, err)
		return
	}
	h.refreshPanel(w, r)
}

func (h *Transactions) Update(w http.ResponseWriter, r *http.Request) {
	in, err := h.parseForm(r)
	if err != nil {
		httpError(w, err)
		return
	}
	one, err := h.txs.Update(r.Context(), chi.URLParam(r, "id"), in)
	if err != nil {
		httpError(w, err)
		return
	}
	h.rnd.Partial(w, http.StatusOK, "tx-row", one)
}

func (h *Transactions) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.txs.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		httpError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Transactions) Row(w http.ResponseWriter, r *http.Request) {
	one, err := h.txs.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httpError(w, err)
		return
	}
	h.rnd.Partial(w, http.StatusOK, "tx-row", one)
}

func (h *Transactions) EditRow(w http.ResponseWriter, r *http.Request) {
	one, err := h.txs.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httpError(w, err)
		return
	}
	accs, _ := h.accs.List(r.Context())
	cats, _ := h.cats.List(r.Context())
	h.rnd.Partial(w, http.StatusOK, "tx-row-edit", map[string]any{
		"Tx": one, "Accounts": accs, "Categories": cats,
	})
}

// refreshPanel re-renders the unfiltered first page after a mutation.
func (h *Transactions) refreshPanel(w http.ResponseWriter, r *http.Request) {
	f := domain.TxFilter{Limit: txPageSize, Offset: 0}
	items, total, err := h.txs.List(r.Context(), f)
	if err != nil {
		httpError(w, err)
		return
	}
	h.rnd.Partial(w, http.StatusOK, "tx-panel", txPanel{
		Txns: items, Page: 1, Pages: pages(total), Total: total,
	})
}

func (h *Transactions) buildPanel(r *http.Request) (txPanel, error) {
	q := r.URL.Query()
	ff := filterForm{
		From:       q.Get("from"),
		To:         q.Get("to"),
		AccountID:  q.Get("account_id"),
		CategoryID: q.Get("category_id"),
		Kind:       q.Get("kind"),
	}
	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}

	f := domain.TxFilter{
		AccountID:  ff.AccountID,
		CategoryID: ff.CategoryID,
		Kind:       ff.Kind,
		Limit:      txPageSize,
		Offset:     (page - 1) * txPageSize,
	}
	if ff.From != "" {
		if t, err := time.Parse("2006-01-02", ff.From); err == nil {
			f.From = &t
		}
	}
	if ff.To != "" {
		if t, err := time.Parse("2006-01-02", ff.To); err == nil {
			f.To = &t
		}
	}

	items, total, err := h.txs.List(r.Context(), f)
	if err != nil {
		return txPanel{}, err
	}
	return txPanel{
		Txns: items, Page: page, Pages: pages(total), Total: total, Filter: f, FForm: ff,
	}, nil
}

func (h *Transactions) parseForm(r *http.Request) (service.TxInput, error) {
	if err := r.ParseForm(); err != nil {
		return service.TxInput{}, err
	}
	amount, err := money.Parse(r.PostFormValue("amount"))
	if err != nil {
		return service.TxInput{}, errJoinInvalid(err)
	}
	occ, err := time.Parse("2006-01-02", r.PostFormValue("occurred_at"))
	if err != nil {
		return service.TxInput{}, errJoinInvalid(err)
	}
	return service.TxInput{
		Kind:        r.PostFormValue("kind"),
		AmountMinor: amount,
		AccountID:   r.PostFormValue("account_id"),
		ToAccountID: r.PostFormValue("to_account_id"),
		CategoryID:  r.PostFormValue("category_id"),
		OccurredAt:  occ,
		Note:        r.PostFormValue("note"),
	}, nil
}

func pages(total int) int {
	if total <= 0 {
		return 1
	}
	return (total + txPageSize - 1) / txPageSize
}
