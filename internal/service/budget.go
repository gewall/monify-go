package service

import (
	"context"
	"fmt"
	"time"

	"github.com/alginugraha/monify/internal/domain"
)

// BudgetRepo is the persistence port for budgets.
type BudgetRepo interface {
	Lines(ctx context.Context, period time.Time) ([]domain.BudgetLine, error)
	LineFor(ctx context.Context, categoryID string, period time.Time) (domain.BudgetLine, error)
	Set(ctx context.Context, categoryID string, period time.Time, amount int64) error
}

type BudgetService struct {
	repo BudgetRepo
}

func NewBudgetService(repo BudgetRepo) *BudgetService {
	return &BudgetService{repo: repo}
}

func (s *BudgetService) Lines(ctx context.Context, month time.Time) ([]domain.BudgetLine, error) {
	return s.repo.Lines(ctx, MonthStart(month))
}

func (s *BudgetService) Set(ctx context.Context, categoryID string, month time.Time, amount int64) (domain.BudgetLine, error) {
	if categoryID == "" {
		return domain.BudgetLine{}, fmt.Errorf("%w: kategori wajib", domain.ErrInvalid)
	}
	if amount < 0 {
		return domain.BudgetLine{}, fmt.Errorf("%w: nominal tidak boleh negatif", domain.ErrInvalid)
	}
	period := MonthStart(month)
	if err := s.repo.Set(ctx, categoryID, period, amount); err != nil {
		return domain.BudgetLine{}, err
	}
	return s.repo.LineFor(ctx, categoryID, period)
}
