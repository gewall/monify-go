package handler

import (
	"net/http"
	"time"

	"github.com/alginugraha/monify/internal/http/middleware"
	"github.com/alginugraha/monify/internal/http/render"
	"github.com/alginugraha/monify/internal/service"
	"github.com/go-chi/chi/v5"
)

type Dashboard struct {
	rnd       *render.Renderer
	report    *service.ReportService
	recurring *service.RecurringService
}

func NewDashboard(rnd *render.Renderer, report *service.ReportService, recurring *service.RecurringService) *Dashboard {
	return &Dashboard{rnd: rnd, report: report, recurring: recurring}
}

func (h *Dashboard) Routes(r chi.Router) {
	r.Get("/", h.Index)
}

func (h *Dashboard) Index(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.CurrentUser(r.Context())

	sum, err := h.report.Dashboard(r.Context(), time.Now())
	if err != nil {
		httpError(w, err)
		return
	}
	surplus, err := h.report.AvgMonthlySurplus(r.Context())
	if err != nil {
		httpError(w, err)
		return
	}
	upcoming, err := h.recurring.Upcoming(r.Context(), 7)
	if err != nil {
		httpError(w, err)
		return
	}

	// Largest top-expense amount, for bar scaling.
	var maxExp int64
	for _, e := range sum.TopExpenses {
		if e.AmountMinor > maxExp {
			maxExp = e.AmountMinor
		}
	}

	h.rnd.Page(w, http.StatusOK, "dashboard.html", map[string]any{
		"Auth":       true,
		"User":       user,
		"Sum":        sum,
		"AvgSurplus": surplus,
		"Upcoming":   upcoming,
		"MaxExpense": maxExp,
		"MonthLabel": sum.Month.Format("January 2006"),
	})
}
