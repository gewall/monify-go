package postgres

import (
	"context"
	"time"

	"github.com/alginugraha/monify/internal/domain"
)

type BudgetRepo struct {
	db DBTX
}

func NewBudgetRepo(db DBTX) *BudgetRepo {
	return &BudgetRepo{db: db}
}

// Lines returns every expense category with its budget (0 if unset) and the
// actual spend for the given month.
func (r *BudgetRepo) Lines(ctx context.Context, period time.Time) ([]domain.BudgetLine, error) {
	next := period.AddDate(0, 1, 0)
	const q = `
SELECT c.id, c.name,
       COALESCE(b.amount_minor, 0) AS budget,
       COALESCE(SUM(t.amount_minor) FILTER (
         WHERE t.deleted_at IS NULL AND t.kind='expense'
           AND t.occurred_at >= $1 AND t.occurred_at < $2), 0) AS spent
FROM categories c
LEFT JOIN budgets b ON b.category_id = c.id AND b.period = $1 AND b.deleted_at IS NULL
LEFT JOIN transactions t ON t.category_id = c.id
WHERE c.deleted_at IS NULL AND c.kind = 'expense'
GROUP BY c.id, c.name, b.amount_minor
ORDER BY c.name`
	rows, err := r.db.Query(ctx, q, period, next)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.BudgetLine
	for rows.Next() {
		var l domain.BudgetLine
		if err := rows.Scan(&l.CategoryID, &l.CategoryName, &l.AmountMinor, &l.SpentMinor); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

// Set upserts a budget amount; amount 0 removes the row.
func (r *BudgetRepo) Set(ctx context.Context, categoryID string, period time.Time, amount int64) error {
	if amount == 0 {
		_, err := r.db.Exec(ctx,
			`UPDATE budgets SET deleted_at=now(), updated_at=now()
			 WHERE category_id=$1 AND period=$2 AND deleted_at IS NULL`, categoryID, period)
		return err
	}
	const q = `
INSERT INTO budgets (category_id, period, amount_minor)
VALUES ($1, $2, $3)
ON CONFLICT (category_id, period) DO UPDATE
  SET amount_minor = EXCLUDED.amount_minor, deleted_at = NULL, updated_at = now()`
	_, err := r.db.Exec(ctx, q, categoryID, period, amount)
	return err
}

// LineFor returns a single category's budget line for the month.
func (r *BudgetRepo) LineFor(ctx context.Context, categoryID string, period time.Time) (domain.BudgetLine, error) {
	lines, err := r.Lines(ctx, period)
	if err != nil {
		return domain.BudgetLine{}, err
	}
	for _, l := range lines {
		if l.CategoryID == categoryID {
			return l, nil
		}
	}
	return domain.BudgetLine{}, domain.ErrNotFound
}
