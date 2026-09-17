package apihandler

import (
	"net/http"
	"time"

	"github.com/alginugraha/monify/internal/service"
	"github.com/go-chi/chi/v5"
)

type Budgets struct {
	svc *service.BudgetService
}

func NewBudgets(svc *service.BudgetService) *Budgets {
	return &Budgets{svc: svc}
}

func (h *Budgets) Routes(r chi.Router) {
	r.Get("/budgets", h.List)
	r.Post("/budgets", h.Set)
}

type budgetSetBody struct {
	CategoryID  string `json:"category_id"`
	Month       string `json:"month"` // "YYYY-MM"
	AmountMinor int64  `json:"amount_minor"`
}

// budgetMonth reads ?month=YYYY-MM, defaulting to the current month.
func budgetMonth(v string) time.Time {
	if v != "" {
		if t, err := time.Parse("2006-01", v); err == nil {
			return t
		}
	}
	return time.Now()
}

func (h *Budgets) List(w http.ResponseWriter, r *http.Request) {
	month := budgetMonth(r.URL.Query().Get("month"))
	lines, err := h.svc.Lines(r.Context(), month)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"month": service.MonthStart(month).Format("2006-01"),
		"lines": lines,
	})
}

func (h *Budgets) Set(w http.ResponseWriter, r *http.Request) {
	var body budgetSetBody
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, err)
		return
	}
	line, err := h.svc.Set(r.Context(), body.CategoryID, budgetMonth(body.Month), body.AmountMinor)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, line)
}
