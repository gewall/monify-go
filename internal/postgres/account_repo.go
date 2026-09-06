package postgres

import (
	"context"
	"errors"

	"github.com/alginugraha/monify/internal/domain"
	"github.com/jackc/pgx/v5"
)

type AccountRepo struct {
	db DBTX
}

func NewAccountRepo(db DBTX) *AccountRepo {
	return &AccountRepo{db: db}
}

// List returns all non-deleted accounts with their current balance
// (initial_balance + net of posted transactions).
func (r *AccountRepo) List(ctx context.Context) ([]domain.Account, error) {
	const q = `
SELECT a.id, a.name, a.type, a.initial_balance, a.currency, a.archived,
       a.created_at, a.updated_at,
       a.initial_balance
         + COALESCE(SUM(CASE
             WHEN t.kind = 'income'   THEN t.amount_minor
             WHEN t.kind = 'expense'  THEN -t.amount_minor
             WHEN t.kind = 'transfer' AND t.account_id    = a.id THEN -t.amount_minor
             WHEN t.kind = 'transfer' AND t.to_account_id = a.id THEN  t.amount_minor
             ELSE 0 END), 0) AS balance
FROM accounts a
LEFT JOIN transactions t
       ON t.deleted_at IS NULL
      AND (t.account_id = a.id OR t.to_account_id = a.id)
WHERE a.deleted_at IS NULL
GROUP BY a.id
ORDER BY a.archived, a.name`
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Account
	for rows.Next() {
		var a domain.Account
		if err := rows.Scan(&a.ID, &a.Name, &a.Type, &a.InitialBalance, &a.Currency,
			&a.Archived, &a.CreatedAt, &a.UpdatedAt, &a.Balance); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *AccountRepo) ByID(ctx context.Context, id string) (domain.Account, error) {
	const q = `SELECT id, name, type, initial_balance, currency, archived, created_at, updated_at
	           FROM accounts WHERE id = $1 AND deleted_at IS NULL`
	var a domain.Account
	err := r.db.QueryRow(ctx, q, id).Scan(&a.ID, &a.Name, &a.Type, &a.InitialBalance,
		&a.Currency, &a.Archived, &a.CreatedAt, &a.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Account{}, domain.ErrNotFound
	}
	return a, err
}

func (r *AccountRepo) Create(ctx context.Context, a domain.Account) (domain.Account, error) {
	const q = `INSERT INTO accounts (name, type, initial_balance, currency)
	           VALUES ($1,$2,$3,$4)
	           RETURNING id, name, type, initial_balance, currency, archived, created_at, updated_at`
	var out domain.Account
	err := r.db.QueryRow(ctx, q, a.Name, a.Type, a.InitialBalance, a.Currency).
		Scan(&out.ID, &out.Name, &out.Type, &out.InitialBalance, &out.Currency,
			&out.Archived, &out.CreatedAt, &out.UpdatedAt)
	return out, err
}

func (r *AccountRepo) Update(ctx context.Context, a domain.Account) error {
	const q = `UPDATE accounts
	           SET name=$2, type=$3, initial_balance=$4, currency=$5, archived=$6, updated_at=now()
	           WHERE id=$1 AND deleted_at IS NULL`
	ct, err := r.db.Exec(ctx, q, a.ID, a.Name, a.Type, a.InitialBalance, a.Currency, a.Archived)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *AccountRepo) SoftDelete(ctx context.Context, id string) error {
	ct, err := r.db.Exec(ctx,
		`UPDATE accounts SET deleted_at=now(), updated_at=now() WHERE id=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
