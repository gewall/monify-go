package service

import (
	"context"
	"time"

	"github.com/alginugraha/monify/internal/domain"
)

// ReportRepo aggregates figures for the dashboard.
type ReportRepo interface {
	TotalBalance(ctx context.Context) (int64, error)
	Cashflow(ctx context.Context, from, to time.Time) (income, expense int64, err error)
	TopExpenseCategories(ctx context.Context, from, to time.Time, limit int) ([]domain.CategorySpend, error)
	AvgMonthlySurplus(ctx context.Context, months int) (int64, error)
}

// BudgetLineRepo reads budget-vs-actual lines.
type BudgetLineRepo interface {
	Lines(ctx context.Context, period time.Time) ([]domain.BudgetLine, error)
}

type ReportService struct {
	rep     ReportRepo
	budgets BudgetLineRepo
}

func NewReportService(rep ReportRepo, budgets BudgetLineRepo) *ReportService {
	return &ReportService{rep: rep, budgets: budgets}
}

// MonthStart returns the first day (local midnight) of t's month.
func MonthStart(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

func (s *ReportService) Dashboard(ctx context.Context, month time.Time) (domain.DashboardSummary, error) {
	from := MonthStart(month)
	to := from.AddDate(0, 1, 0)

	sum := domain.DashboardSummary{Month: from}

	var err error
	if sum.TotalBalance, err = s.rep.TotalBalance(ctx); err != nil {
		return sum, err
	}
	if sum.MonthIncome, sum.MonthExpense, err = s.rep.Cashflow(ctx, from, to); err != nil {
		return sum, err
	}
	if sum.TopExpenses, err = s.rep.TopExpenseCategories(ctx, from, to, 5); err != nil {
		return sum, err
	}

	lines, err := s.budgets.Lines(ctx, from)
	if err != nil {
		return sum, err
	}
	for _, l := range lines {
		if l.AmountMinor > 0 {
			sum.Budgets = append(sum.Budgets, l)
		}
	}
	return sum, nil
}

// AvgMonthlySurplus exposes the rolling surplus average (last 3 months).
func (s *ReportService) AvgMonthlySurplus(ctx context.Context) (int64, error) {
	return s.rep.AvgMonthlySurplus(ctx, 3)
}
