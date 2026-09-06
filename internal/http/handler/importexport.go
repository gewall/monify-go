package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/alginugraha/monify/internal/domain"
	"github.com/alginugraha/monify/internal/http/render"
	"github.com/alginugraha/monify/internal/platform/money"
	"github.com/alginugraha/monify/internal/service"
	"github.com/go-chi/chi/v5"
)

type ImportExport struct {
	rnd *render.Renderer
	imp *service.ImportService
	txs *service.TransactionService
}

func NewImportExport(rnd *render.Renderer, imp *service.ImportService, txs *service.TransactionService) *ImportExport {
	return &ImportExport{rnd: rnd, imp: imp, txs: txs}
}

func (h *ImportExport) Routes(r chi.Router) {
	r.Get("/import", h.Form)
	r.Post("/import/preview", h.Preview)
	r.Post("/import/commit", h.Commit)
	r.Get("/export/transactions.csv", h.Export)
}

func (h *ImportExport) Form(w http.ResponseWriter, r *http.Request) {
	h.rnd.Page(w, http.StatusOK, "import.html", map[string]any{"Auth": true})
}

func (h *ImportExport) Preview(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		httpError(w, err)
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		httpError(w, fmt.Errorf("%w: file wajib diunggah", domain.ErrInvalid))
		return
	}
	defer file.Close()

	rows, err := service.ParseTransactionCSV(file)
	if err != nil {
		httpError(w, err)
		return
	}
	preview, err := h.imp.Preview(r.Context(), rows)
	if err != nil {
		httpError(w, err)
		return
	}

	var importable int
	for _, p := range preview {
		if p.OK() && !p.Duplicate {
			importable++
		}
	}
	h.rnd.Partial(w, http.StatusOK, "import-preview", map[string]any{
		"Rows": preview, "Importable": importable,
	})
}

func (h *ImportExport) Commit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		httpError(w, err)
		return
	}
	created, failed := 0, 0
	for _, idx := range r.Form["include"] {
		amt, err := money.Parse(r.PostFormValue("amount_" + idx))
		if err != nil {
			failed++
			continue
		}
		occ, err := time.Parse("2006-01-02", r.PostFormValue("date_"+idx))
		if err != nil {
			failed++
			continue
		}
		in := service.TxInput{
			Kind:        r.PostFormValue("kind_" + idx),
			AmountMinor: amt,
			AccountID:   r.PostFormValue("account_" + idx),
			ToAccountID: r.PostFormValue("to_account_" + idx),
			CategoryID:  r.PostFormValue("category_" + idx),
			OccurredAt:  occ,
			Note:        r.PostFormValue("note_" + idx),
		}
		if _, err := h.txs.Create(r.Context(), in); err != nil {
			failed++
			continue
		}
		created++
	}
	h.rnd.Partial(w, http.StatusOK, "import-result", map[string]any{
		"Created": created, "Failed": failed,
	})
}

func (h *ImportExport) Export(w http.ResponseWriter, r *http.Request) {
	f := domain.TxFilter{} // no limit -> everything
	if v := r.URL.Query().Get("from"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			f.From = &t
		}
	}
	if v := r.URL.Query().Get("to"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			f.To = &t
		}
	}
	txns, _, err := h.txs.List(r.Context(), f)
	if err != nil {
		httpError(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition",
		fmt.Sprintf(`attachment; filename="monify-transaksi-%s.csv"`, time.Now().Format("20060102")))
	if err := service.WriteTransactionCSV(w, txns); err != nil {
		httpError(w, err)
	}
}
