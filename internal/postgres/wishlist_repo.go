package postgres

import (
	"context"
	"errors"

	"github.com/alginugraha/monify/internal/domain"
	"github.com/jackc/pgx/v5"
)

type WishlistRepo struct {
	db DBTX
}

func NewWishlistRepo(db DBTX) *WishlistRepo {
	return &WishlistRepo{db: db}
}

const wishColumns = `id, name, target_amount_minor, saved_amount_minor, priority,
	target_date, url, status, created_at, updated_at`

func scanWish(row scannable) (domain.WishlistItem, error) {
	var w domain.WishlistItem
	err := row.Scan(&w.ID, &w.Name, &w.TargetAmountMinor, &w.SavedAmountMinor, &w.Priority,
		&w.TargetDate, &w.URL, &w.Status, &w.CreatedAt, &w.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.WishlistItem{}, domain.ErrNotFound
	}
	return w, err
}

// List returns items ordered for sequential funding: active first, by priority
// then target date then creation.
func (r *WishlistRepo) List(ctx context.Context) ([]domain.WishlistItem, error) {
	rows, err := r.db.Query(ctx, `SELECT `+wishColumns+` FROM wishlist_items
		WHERE deleted_at IS NULL
		ORDER BY (status <> 'active'), priority, target_date NULLS LAST, created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.WishlistItem
	for rows.Next() {
		w, err := scanWish(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

func (r *WishlistRepo) ByID(ctx context.Context, id string) (domain.WishlistItem, error) {
	return scanWish(r.db.QueryRow(ctx,
		`SELECT `+wishColumns+` FROM wishlist_items WHERE id=$1 AND deleted_at IS NULL`, id))
}

func (r *WishlistRepo) Create(ctx context.Context, w domain.WishlistItem) (domain.WishlistItem, error) {
	return scanWish(r.db.QueryRow(ctx, `
INSERT INTO wishlist_items (name, target_amount_minor, saved_amount_minor, priority, target_date, url, status)
VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING `+wishColumns,
		w.Name, w.TargetAmountMinor, w.SavedAmountMinor, w.Priority, w.TargetDate, w.URL, w.Status))
}

func (r *WishlistRepo) Update(ctx context.Context, w domain.WishlistItem) error {
	ct, err := r.db.Exec(ctx, `
UPDATE wishlist_items SET
  name=$2, target_amount_minor=$3, saved_amount_minor=$4, priority=$5,
  target_date=$6, url=$7, status=$8, updated_at=now()
WHERE id=$1 AND deleted_at IS NULL`,
		w.ID, w.Name, w.TargetAmountMinor, w.SavedAmountMinor, w.Priority,
		w.TargetDate, w.URL, w.Status)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *WishlistRepo) SoftDelete(ctx context.Context, id string) error {
	ct, err := r.db.Exec(ctx,
		`UPDATE wishlist_items SET deleted_at=now(), updated_at=now() WHERE id=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
