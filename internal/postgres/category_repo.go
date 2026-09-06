package postgres

import (
	"context"
	"errors"

	"github.com/alginugraha/monify/internal/domain"
	"github.com/jackc/pgx/v5"
)

type CategoryRepo struct {
	db DBTX
}

func NewCategoryRepo(db DBTX) *CategoryRepo {
	return &CategoryRepo{db: db}
}

func (r *CategoryRepo) List(ctx context.Context) ([]domain.Category, error) {
	const q = `SELECT id, name, kind, parent_id, created_at, updated_at
	           FROM categories WHERE deleted_at IS NULL
	           ORDER BY kind, COALESCE(parent_id::text, id::text), parent_id NULLS FIRST, name`
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Category
	for rows.Next() {
		var c domain.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Kind, &c.ParentID, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *CategoryRepo) ByID(ctx context.Context, id string) (domain.Category, error) {
	const q = `SELECT id, name, kind, parent_id, created_at, updated_at
	           FROM categories WHERE id=$1 AND deleted_at IS NULL`
	var c domain.Category
	err := r.db.QueryRow(ctx, q, id).Scan(&c.ID, &c.Name, &c.Kind, &c.ParentID, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Category{}, domain.ErrNotFound
	}
	return c, err
}

func (r *CategoryRepo) Create(ctx context.Context, c domain.Category) (domain.Category, error) {
	const q = `INSERT INTO categories (name, kind, parent_id) VALUES ($1,$2,$3)
	           RETURNING id, name, kind, parent_id, created_at, updated_at`
	var out domain.Category
	err := r.db.QueryRow(ctx, q, c.Name, c.Kind, c.ParentID).
		Scan(&out.ID, &out.Name, &out.Kind, &out.ParentID, &out.CreatedAt, &out.UpdatedAt)
	return out, err
}

func (r *CategoryRepo) Update(ctx context.Context, c domain.Category) error {
	const q = `UPDATE categories SET name=$2, kind=$3, parent_id=$4, updated_at=now()
	           WHERE id=$1 AND deleted_at IS NULL`
	ct, err := r.db.Exec(ctx, q, c.ID, c.Name, c.Kind, c.ParentID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *CategoryRepo) SoftDelete(ctx context.Context, id string) error {
	ct, err := r.db.Exec(ctx,
		`UPDATE categories SET deleted_at=now(), updated_at=now() WHERE id=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
