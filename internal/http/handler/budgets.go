package handler

import (
	"net/http"
	"time"

	"github.com/alginugraha/monify/internal/domain"
	"github.com/alginugraha/monify/internal/http/render"
	"github.com/alginugraha/monify/internal/platform/money"
	"github.com/alginugraha/monify/internal/service"
	"github.com/go-chi/chi/v5"
)

// budgetRowVM carries the month alongside a line so the row form can post it back.
type budgetRowVM struct {
	domain.BudgetLine
	Month string
}

type Budgets struct {
	rnd *render.Renderer
	svc *service.BudgetService
}

func NewBudgets(rnd *render.Renderer, svc *service.BudgetService) *Budgets {
	return &Budgets{rnd: rnd, svc: svc}
}

func (h *Budgets) Routes(r chi.Router) {
	r.Get("/budgets", h.Index)
	r.Post("/budgets", h.Set)
}

// monthParam reads ?month=YYYY-MM, defaulting to the current month.
func monthParam(r *http.Request) time.Time {
	if v := r.FormValue("month"); v != "" {
		if t, err := time.Parse("2006-01", v); err == nil {
			return t
		}
	}
	return time.Now()
}

func (h *Budgets) Index(w http.ResponseWriter, r *http.Request) {
	month := monthParam(r)
	lines, err := h.svc.Lines(r.Context(), month)
	if err != nil {
		httpError(w, err)
		return
	}
	mStr := service.MonthStart(month).Format("2006-01")
	rows := make([]budgetRowVM, len(lines))
	for i, l := range lines {
		rows[i] = budgetRowVM{BudgetLine: l, Month: mStr}
	}
	h.rnd.Page(w, http.StatusOK, "budgets.html", map[string]any{
		"Auth":       true,
		"Rows":       rows,
		"Month":      mStr,
		"MonthLabel": service.MonthStart(month).Format("January 2006"),
	})
}

func (h *Budgets) Set(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		httpError(w, err)
		return
	}
	month := monthParam(r)
	var amount int64
	if v := r.PostFormValue("amount"); v != "" {
		parsed, err := money.Parse(v)
		if err != nil {
			httpError(w, errJoinInvalid(err))
			return
		}
		amount = parsed
	}
	line, err := h.svc.Set(r.Context(), r.PostFormValue("category_id"), month, amount)
	if err != nil {
		httpError(w, err)
		return
	}
	h.rnd.Partial(w, http.StatusOK, "budget-row", budgetRowVM{
		BudgetLine: line, Month: service.MonthStart(month).Format("2006-01"),
	})
}
