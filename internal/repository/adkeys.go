package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourname/novapromptgobackend/internal/models"
)

type AdKeyRepo struct{ pool *pgxpool.Pool }

func NewAdKeyRepo(pool *pgxpool.Pool) *AdKeyRepo { return &AdKeyRepo{pool: pool} }

func (r *AdKeyRepo) Upsert(ctx context.Context, k *models.AdKey) error {
	raw, err := json.Marshal(k.Value)
	if err != nil {
		return err
	}
	const q = `
		INSERT INTO ad_keys (key, value)
		VALUES ($1, $2)
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value
		RETURNING updated_at`
	return r.pool.QueryRow(ctx, q, k.Key, raw).Scan(&k.UpdatedAt)
}

func (r *AdKeyRepo) Get(ctx context.Context, key string) (*models.AdKey, error) {
	var (
		k   models.AdKey
		raw []byte
	)
	err := r.pool.QueryRow(ctx,
		`SELECT key, value, updated_at FROM ad_keys WHERE key = $1`, key).
		Scan(&k.Key, &raw, &k.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(raw, &k.Value); err != nil {
		return nil, err
	}
	return &k, nil
}

func (r *AdKeyRepo) List(ctx context.Context) ([]models.AdKey, error) {
	rows, err := r.pool.Query(ctx, `SELECT key, value, updated_at FROM ad_keys ORDER BY key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.AdKey
	for rows.Next() {
		var (
			k   models.AdKey
			raw []byte
		)
		if err := rows.Scan(&k.Key, &raw, &k.UpdatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(raw, &k.Value)
		out = append(out, k)
	}
	return out, rows.Err()
}