package postgres

import (
	"context"
	"time"

	"github.com/alginugraha/monify/internal/domain"
)

type ReportRepo struct {
	db DBTX
}

func NewReportRepo(db DBTX) *ReportRepo {
	return &ReportRepo{db: db}
}

// TotalBalance sums every non-deleted account's current balance.
func (r *ReportRepo) TotalBalance(ctx context.Context) (int64, error) {
	const q = `
SELECT COALESCE(SUM(bal), 0) FROM (
  SELECT a.initial_balance
       + COALESCE(SUM(CASE
           WHEN t.kind='income'  THEN t.amount_minor
           WHEN t.kind='expense' THEN -t.amount_minor
           WHEN t.kind='transfer' AND t.account_id    = a.id THEN -t.amount_minor
           WHEN t.kind='transfer' AND t.to_account_id = a.id THEN  t.amount_minor
           ELSE 0 END), 0) AS bal
  FROM accounts a
  LEFT JOIN transactions t ON t.deleted_at IS NULL
        AND (t.account_id = a.id OR t.to_account_id = a.id)
  WHERE a.deleted_at IS NULL
  GROUP BY a.id
) s`
	var total int64
	err := r.db.QueryRow(ctx, q).Scan(&total)
	return total, err
}

// Cashflow returns total income and expense in [from, to).
func (r *ReportRepo) Cashflow(ctx context.Context, from, to time.Time) (income, expense int64, err error) {
	const q = `
SELECT
  COALESCE(SUM(amount_minor) FILTER (WHERE kind='income'), 0),
  COALESCE(SUM(amount_minor) FILTER (WHERE kind='expense'), 0)
FROM transactions
WHERE deleted_at IS NULL AND occurred_at >= $1 AND occurred_at < $2`
	err = r.db.QueryRow(ctx, q, from, to).Scan(&income, &expense)
	return
}

// TopExpenseCategories lists the biggest expense categories in [from, to).
func (r *ReportRepo) TopExpenseCategories(ctx context.Context, from, to time.Time, limit int) ([]domain.CategorySpend, error) {
	const q = `
SELECT COALESCE(c.id::text, ''), COALESCE(c.name, 'Tanpa kategori'), SUM(t.amount_minor)
FROM transactions t
LEFT JOIN categories c ON c.id = t.category_id
WHERE t.deleted_at IS NULL AND t.kind='expense'
  AND t.occurred_at >= $1 AND t.occurred_at < $2
GROUP BY c.id, c.name
ORDER BY SUM(t.amount_minor) DESC
LIMIT $3`
	rows, err := r.db.Query(ctx, q, from, to, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.CategorySpend
	for rows.Next() {
		var cs domain.CategorySpend
		if err := rows.Scan(&cs.CategoryID, &cs.CategoryName, &cs.AmountMinor); err != nil {
			return nil, err
		}
		out = append(out, cs)
	}
	return out, rows.Err()
}

// AvgMonthlySurplus averages (income - expense) over the last n whole months
// ending before the current month.
func (r *ReportRepo) AvgMonthlySurplus(ctx context.Context, months int) (int64, error) {
	const q = `
WITH m AS (
  SELECT date_trunc('month', occurred_at)::date AS mon,
         SUM(CASE WHEN kind='income' THEN amount_minor
                  WHEN kind='expense' THEN -amount_minor ELSE 0 END) AS net
  FROM transactions
  WHERE deleted_at IS NULL
    AND occurred_at >= (date_trunc('month', now()) - ($1 || ' months')::interval)::date
    AND occurred_at <  date_trunc('month', now())::date
  GROUP BY 1
)
SELECT COALESCE(AVG(net), 0)::bigint FROM m`
	var avg int64
	err := r.db.QueryRow(ctx, q, months).Scan(&avg)
	return avg, err
}
