package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/alginugraha/monify/internal/http/render"
	"github.com/alginugraha/monify/internal/platform/money"
	"github.com/alginugraha/monify/internal/service"
	"github.com/go-chi/chi/v5"
)

type Recurring struct {
	rnd  *render.Renderer
	svc  *service.RecurringService
	accs *service.AccountService
	cats *service.CategoryService
}

func NewRecurring(rnd *render.Renderer, svc *service.RecurringService,
	accs *service.AccountService, cats *service.CategoryService) *Recurring {
	return &Recurring{rnd: rnd, svc: svc, accs: accs, cats: cats}
}

func (h *Recurring) Routes(r chi.Router) {
	r.Get("/recurring", h.Index)
	r.Post("/recurring", h.Create)
	r.Post("/recurring/run", h.Run)
	r.Get("/recurring/{id}/edit", h.EditRow)
	r.Get("/recurring/{id}/row", h.Row)
	r.Put("/recurring/{id}", h.Update)
	r.Delete("/recurring/{id}", h.Delete)
}

func (h *Recurring) formData(r *http.Request) map[string]any {
	accs, _ := h.accs.List(r.Context())
	cats, _ := h.cats.List(r.Context())
	return map[string]any{"Accounts": accs, "Categories": cats}
}

func (h *Recurring) Index(w http.ResponseWriter, r *http.Request) {
	rules, err := h.svc.List(r.Context())
	if err != nil {
		httpError(w, err)
		return
	}
	data := h.formData(r)
	data["Auth"] = true
	data["Rules"] = rules
	data["Today"] = time.Now().Format("2006-01-02")
	h.rnd.Page(w, http.StatusOK, "recurring.html", data)
}

func (h *Recurring) Create(w http.ResponseWriter, r *http.Request) {
	in, err := h.parseForm(r)
	if err != nil {
		httpError(w, err)
		return
	}
	rule, err := h.svc.Create(r.Context(), in)
	if err != nil {
		httpError(w, err)
		return
	}
	h.rnd.Partial(w, http.StatusOK, "rule-row", rule)
}

func (h *Recurring) Update(w http.ResponseWriter, r *http.Request) {
	in, err := h.parseForm(r)
	if err != nil {
		httpError(w, err)
		return
	}
	rule, err := h.svc.Update(r.Context(), chi.URLParam(r, "id"), in)
	if err != nil {
		httpError(w, err)
		return
	}
	h.rnd.Partial(w, http.StatusOK, "rule-row", rule)
}

func (h *Recurring) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		httpError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Recurring) Row(w http.ResponseWriter, r *http.Request) {
	rule, err := h.svc.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httpError(w, err)
		return
	}
	h.rnd.Partial(w, http.StatusOK, "rule-row", rule)
}

func (h *Recurring) EditRow(w http.ResponseWriter, r *http.Request) {
	rule, err := h.svc.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httpError(w, err)
		return
	}
	data := h.formData(r)
	data["Rule"] = rule
	h.rnd.Partial(w, http.StatusOK, "rule-row-edit", data)
}

func (h *Recurring) Run(w http.ResponseWriter, r *http.Request) {
	n, err := h.svc.Run(r.Context())
	if err != nil {
		httpError(w, err)
		return
	}
	rules, err := h.svc.List(r.Context())
	if err != nil {
		httpError(w, err)
		return
	}
	h.rnd.Partial(w, http.StatusOK, "rule-table", map[string]any{"Rules": rules, "Generated": n})
}

func (h *Recurring) parseForm(r *http.Request) (service.RuleInput, error) {
	if err := r.ParseForm(); err != nil {
		return service.RuleInput{}, err
	}
	amount, err := money.Parse(r.PostFormValue("amount"))
	if err != nil {
		return service.RuleInput{}, errJoinInvalid(err)
	}
	next, err := time.Parse("2006-01-02", r.PostFormValue("next_run_at"))
	if err != nil {
		return service.RuleInput{}, errJoinInvalid(err)
	}
	interval, _ := strconv.Atoi(r.PostFormValue("interval"))
	dom, _ := strconv.Atoi(r.PostFormValue("day_of_month"))

	in := service.RuleInput{
		Name:        r.PostFormValue("name"),
		Kind:        r.PostFormValue("kind"),
		AmountMinor: amount,
		AccountID:   r.PostFormValue("account_id"),
		ToAccountID: r.PostFormValue("to_account_id"),
		CategoryID:  r.PostFormValue("category_id"),
		Note:        r.PostFormValue("note"),
		Freq:        r.PostFormValue("freq"),
		Interval:    interval,
		DayOfMonth:  dom,
		NextRunAt:   next,
		Active:      r.PostFormValue("active") == "on",
	}
	if v := r.PostFormValue("end_at"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			in.EndAt = &t
		}
	}
	return in, nil
}
