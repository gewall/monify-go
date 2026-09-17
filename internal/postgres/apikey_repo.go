package postgres

import (
	"context"
	"errors"

	"github.com/alginugraha/monify/internal/domain"
	"github.com/jackc/pgx/v5"
)

type APIKeyRepo struct {
	db DBTX
}

func NewAPIKeyRepo(db DBTX) *APIKeyRepo {
	return &APIKeyRepo{db: db}
}

func (r *APIKeyRepo) Create(ctx context.Context, k domain.APIKey) (domain.APIKey, error) {
	const q = `INSERT INTO api_keys (user_id, name, key_hash, key_prefix)
	           VALUES ($1,$2,$3,$4)
	           RETURNING id, user_id, name, key_hash, key_prefix, created_at`
	var out domain.APIKey
	err := r.db.QueryRow(ctx, q, k.UserID, k.Name, k.KeyHash, k.KeyPrefix).
		Scan(&out.ID, &out.UserID, &out.Name, &out.KeyHash, &out.KeyPrefix, &out.CreatedAt)
	return out, err
}

// ByHash looks up a non-revoked key by its hashed token.
func (r *APIKeyRepo) ByHash(ctx context.Context, hash string) (domain.APIKey, error) {
	const q = `SELECT id, user_id, name, key_hash, key_prefix, last_used_at, created_at, revoked_at
	           FROM api_keys WHERE key_hash = $1 AND revoked_at IS NULL`
	var k domain.APIKey
	err := r.db.QueryRow(ctx, q, hash).Scan(&k.ID, &k.UserID, &k.Name, &k.KeyHash, &k.KeyPrefix,
		&k.LastUsedAt, &k.CreatedAt, &k.RevokedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.APIKey{}, domain.ErrNotFound
	}
	return k, err
}

func (r *APIKeyRepo) Touch(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `UPDATE api_keys SET last_used_at = now() WHERE id = $1`, id)
	return err
}
