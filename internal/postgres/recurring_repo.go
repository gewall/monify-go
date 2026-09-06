package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/alginugraha/monify/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RecurringRepo struct {
	pool *pgxpool.Pool
}

func NewRecurringRepo(pool *pgxpool.Pool) *RecurringRepo {
	return &RecurringRepo{pool: pool}
}

const ruleColumns = `id, name, kind, amount_minor, account_id, to_account_id, category_id,
	note, freq, interval, day_of_month, next_run_at, end_at, active, created_at, updated_at`

func scanRule(row scannable) (domain.RecurringRule, error) {
	var r domain.RecurringRule
	err := row.Scan(&r.ID, &r.Name, &r.Kind, &r.AmountMinor, &r.AccountID, &r.ToAccountID,
		&r.CategoryID, &r.Note, &r.Freq, &r.Interval, &r.DayOfMonth, &r.NextRunAt, &r.EndAt,
		&r.Active, &r.CreatedAt, &r.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.RecurringRule{}, domain.ErrNotFound
	}
	return r, err
}

func (r *RecurringRepo) List(ctx context.Context) ([]domain.RecurringRule, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+ruleColumns+` FROM recurring_rules WHERE deleted_at IS NULL ORDER BY active DESC, next_run_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.RecurringRule
	for rows.Next() {
		rule, err := scanRule(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, rule)
	}
	return out, rows.Err()
}

func (r *RecurringRepo) ByID(ctx context.Context, id string) (domain.RecurringRule, error) {
	return scanRule(r.pool.QueryRow(ctx,
		`SELECT `+ruleColumns+` FROM recurring_rules WHERE id=$1 AND deleted_at IS NULL`, id))
}

func (r *RecurringRepo) Create(ctx context.Context, rule domain.RecurringRule) (domain.RecurringRule, error) {
	return scanRule(r.pool.QueryRow(ctx, `
INSERT INTO recurring_rules
  (name, kind, amount_minor, account_id, to_account_id, category_id, note,
   freq, interval, day_of_month, next_run_at, end_at, active)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
RETURNING `+ruleColumns,
		rule.Name, rule.Kind, rule.AmountMinor, rule.AccountID, rule.ToAccountID, rule.CategoryID,
		rule.Note, rule.Freq, rule.Interval, rule.DayOfMonth, rule.NextRunAt, rule.EndAt, rule.Active))
}

func (r *RecurringRepo) Update(ctx context.Context, rule domain.RecurringRule) error {
	ct, err := r.pool.Exec(ctx, `
UPDATE recurring_rules SET
  name=$2, kind=$3, amount_minor=$4, account_id=$5, to_account_id=$6, category_id=$7,
  note=$8, freq=$9, interval=$10, day_of_month=$11, next_run_at=$12, end_at=$13, active=$14,
  updated_at=now()
WHERE id=$1 AND deleted_at IS NULL`,
		rule.ID, rule.Name, rule.Kind, rule.AmountMinor, rule.AccountID, rule.ToAccountID,
		rule.CategoryID, rule.Note, rule.Freq, rule.Interval, rule.DayOfMonth, rule.NextRunAt,
		rule.EndAt, rule.Active)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *RecurringRepo) SoftDelete(ctx context.Context, id string) error {
	ct, err := r.pool.Exec(ctx,
		`UPDATE recurring_rules SET deleted_at=now(), updated_at=now() WHERE id=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// RunDue posts transactions for every rule whose next_run_at has arrived, then
// advances the schedule. It holds a session advisory lock so concurrent runners
// (e.g. the ticker and a manual trigger) can't double-post.
func (r *RecurringRepo) RunDue(ctx context.Context, asOf time.Time) (int, error) {
	const lockKey = int64(748923001) // arbitrary app-wide key for "recurring runner"
	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()

	var got bool
	if err := conn.QueryRow(ctx, `SELECT pg_try_advisory_lock($1)`, lockKey).Scan(&got); err != nil {
		return 0, err
	}
	if !got {
		return 0, nil // another runner is working
	}
	defer conn.Exec(context.Background(), `SELECT pg_advisory_unlock($1)`, lockKey)

	tx, err := conn.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx,
		`SELECT `+ruleColumns+` FROM recurring_rules
		 WHERE deleted_at IS NULL AND active AND next_run_at <= $1 FOR UPDATE`, asOf)
	if err != nil {
		return 0, err
	}
	var due []domain.RecurringRule
	for rows.Next() {
		rule, err := scanRule(rows)
		if err != nil {
			rows.Close()
			return 0, err
		}
		due = append(due, rule)
	}
	rows.Close()

	generated := 0
	for _, rule := range due {
		next := rule.NextRunAt
		for i := 0; i < 120 && !next.After(asOf); i++ {
			if rule.EndAt != nil && next.After(*rule.EndAt) {
				break
			}
			if _, err := tx.Exec(ctx, `
INSERT INTO transactions (kind, amount_minor, account_id, to_account_id, category_id, occurred_at, note)
VALUES ($1,$2,$3,$4,$5,$6,$7)`,
				rule.Kind, rule.AmountMinor, rule.AccountID, rule.ToAccountID, rule.CategoryID,
				next, rule.Note); err != nil {
				return 0, err
			}
			generated++
			next = rule.Advance(next)
		}
		active := rule.Active
		if rule.EndAt != nil && next.After(*rule.EndAt) {
			active = false
		}
		if _, err := tx.Exec(ctx,
			`UPDATE recurring_rules SET next_run_at=$2, active=$3, updated_at=now() WHERE id=$1`,
			rule.ID, next, active); err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return generated, nil
}
