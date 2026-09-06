package handler

import (
	"net/http"
	"time"

	"github.com/alginugraha/monify/internal/http/render"
	"github.com/alginugraha/monify/internal/service"
	"github.com/go-chi/chi/v5"
)

type Reports struct {
	rnd    *render.Renderer
	report *service.ReportService
}

func NewReports(rnd *render.Renderer, report *service.ReportService) *Reports {
	return &Reports{rnd: rnd, report: report}
}

func (h *Reports) Routes(r chi.Router) {
	r.Get("/reports", h.Index)
}

func (h *Reports) Index(w http.ResponseWriter, r *http.Request) {
	month := time.Now()
	if v := r.FormValue("month"); v != "" {
		if t, err := time.Parse("2006-01", v); err == nil {
			month = t
		}
	}
	sum, err := h.report.Dashboard(r.Context(), month)
	if err != nil {
		httpError(w, err)
		return
	}
	h.rnd.Page(w, http.StatusOK, "reports.html", map[string]any{
		"Auth":       true,
		"Sum":        sum,
		"Month":      service.MonthStart(month).Format("2006-01"),
		"MonthLabel": service.MonthStart(month).Format("January 2006"),
		"Generated":  time.Now().Format("02 Jan 2006 15:04"),
	})
}
