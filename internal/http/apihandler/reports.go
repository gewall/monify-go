package apihandler

import (
	"net/http"
	"time"

	"github.com/alginugraha/monify/internal/service"
	"github.com/go-chi/chi/v5"
)

type Reports struct {
	svc *service.ReportService
}

func NewReports(svc *service.ReportService) *Reports {
	return &Reports{svc: svc}
}

func (h *Reports) Routes(r chi.Router) {
	r.Get("/dashboard", h.Dashboard)
}

// dashboardMonth reads ?month=YYYY-MM, defaulting to the current month.
func dashboardMonth(v string) time.Time {
	if v != "" {
		if t, err := time.Parse("2006-01", v); err == nil {
			return t
		}
	}
	return time.Now()
}

func (h *Reports) Dashboard(w http.ResponseWriter, r *http.Request) {
	month := dashboardMonth(r.URL.Query().Get("month"))
	sum, err := h.svc.Dashboard(r.Context(), month)
	if err != nil {
		writeError(w, err)
		return
	}
	surplus, err := h.svc.AvgMonthlySurplus(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"summary":                   sum,
		"avg_monthly_surplus_minor": surplus,
	})
}
