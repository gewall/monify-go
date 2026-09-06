package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/alginugraha/monify/internal/domain"
	"github.com/jackc/pgx/v5"
)

type TransactionRepo struct {
	db DBTX
}

func NewTransactionRepo(db DBTX) *TransactionRepo {
	return &TransactionRepo{db: db}
}

const txSelect = `
SELECT t.id, t.kind, t.amount_minor, t.account_id, t.to_account_id, t.category_id,
       t.occurred_at, t.note, t.created_at, t.updated_at,
       a.name AS account_name,
       COALESCE(ta.name, '') AS to_account_name,
       COALESCE(c.name, '')  AS category_name
FROM transactions t
JOIN accounts a          ON a.id = t.account_id
LEFT JOIN accounts ta    ON ta.id = t.to_account_id
LEFT JOIN categories c   ON c.id = t.category_id
WHERE t.deleted_at IS NULL`

func txWhere(f domain.TxFilter, args *[]any) string {
	var b strings.Builder
	add := func(cond string, val any) {
		*args = append(*args, val)
		fmt.Fprintf(&b, " AND %s $%d", cond, len(*args))
	}
	if f.From != nil {
		add("t.occurred_at >=", *f.From)
	}
	if f.To != nil {
		add("t.occurred_at <=", *f.To)
	}
	if f.AccountID != "" {
		*args = append(*args, f.AccountID)
		n := len(*args)
		fmt.Fprintf(&b, " AND (t.account_id = $%d OR t.to_account_id = $%d)", n, n)
	}
	if f.CategoryID != "" {
		add("t.category_id =", f.CategoryID)
	}
	if f.Kind != "" {
		add("t.kind =", f.Kind)
	}
	return b.String()
}

func (r *TransactionRepo) List(ctx context.Context, f domain.TxFilter) ([]domain.Transaction, error) {
	args := make([]any, 0, 8)
	q := txSelect + txWhere(f, &args) + " ORDER BY t.occurred_at DESC, t.created_at DESC"
	if f.Limit > 0 {
		args = append(args, f.Limit)
		q += fmt.Sprintf(" LIMIT $%d", len(args))
		args = append(args, f.Offset)
		q += fmt.Sprintf(" OFFSET $%d", len(args))
	}

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Transaction
	for rows.Next() {
		t, err := scanTx(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *TransactionRepo) Count(ctx context.Context, f domain.TxFilter) (int, error) {
	args := make([]any, 0, 8)
	q := `SELECT COUNT(*) FROM transactions t WHERE t.deleted_at IS NULL` + txWhere(f, &args)
	var n int
	err := r.db.QueryRow(ctx, q, args...).Scan(&n)
	return n, err
}

func (r *TransactionRepo) ByID(ctx context.Context, id string) (domain.Transaction, error) {
	args := []any{id}
	row := r.db.QueryRow(ctx, txSelect+" AND t.id = $1", args...)
	t, err := scanTx(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Transaction{}, domain.ErrNotFound
	}
	return t, err
}

func (r *TransactionRepo) Create(ctx context.Context, t domain.Transaction) (string, error) {
	const q = `INSERT INTO transactions
	  (kind, amount_minor, account_id, to_account_id, category_id, occurred_at, note)
	  VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`
	var id string
	err := r.db.QueryRow(ctx, q, t.Kind, t.AmountMinor, t.AccountID, t.ToAccountID,
		t.CategoryID, t.OccurredAt, t.Note).Scan(&id)
	return id, err
}

func (r *TransactionRepo) Update(ctx context.Context, t domain.Transaction) error {
	const q = `UPDATE transactions
	  SET kind=$2, amount_minor=$3, account_id=$4, to_account_id=$5, category_id=$6,
	      occurred_at=$7, note=$8, updated_at=now()
	  WHERE id=$1 AND deleted_at IS NULL`
	ct, err := r.db.Exec(ctx, q, t.ID, t.Kind, t.AmountMinor, t.AccountID, t.ToAccountID,
		t.CategoryID, t.OccurredAt, t.Note)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *TransactionRepo) SoftDelete(ctx context.Context, id string) error {
	ct, err := r.db.Exec(ctx,
		`UPDATE transactions SET deleted_at=now(), updated_at=now() WHERE id=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func scanTx(row scannable) (domain.Transaction, error) {
	var t domain.Transaction
	err := row.Scan(&t.ID, &t.Kind, &t.AmountMinor, &t.AccountID, &t.ToAccountID, &t.CategoryID,
		&t.OccurredAt, &t.Note, &t.CreatedAt, &t.UpdatedAt,
		&t.AccountName, &t.ToAccountName, &t.CategoryName)
	return t, err
}
