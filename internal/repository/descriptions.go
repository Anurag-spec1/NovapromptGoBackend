package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourname/novapromptgobackend/internal/models"
)

type DescriptionRepo struct{ pool *pgxpool.Pool }

func NewDescriptionRepo(pool *pgxpool.Pool) *DescriptionRepo {
	return &DescriptionRepo{pool: pool}
}

func (r *DescriptionRepo) Upsert(ctx context.Context, d *models.Description) error {
	const q = `
		INSERT INTO descriptions (image_id, body)
		VALUES ($1, $2)
		ON CONFLICT (image_id) DO UPDATE SET body = EXCLUDED.body
		RETURNING id, updated_at`
	return r.pool.QueryRow(ctx, q, d.ImageID, d.Body).Scan(&d.ID, &d.UpdatedAt)
}

func (r *DescriptionRepo) GetByImageID(ctx context.Context, imageID string) (*models.Description, error) {
	var d models.Description
	err := r.pool.QueryRow(ctx,
		`SELECT id, image_id, body, updated_at FROM descriptions WHERE image_id = $1`,
		imageID).Scan(&d.ID, &d.ImageID, &d.Body, &d.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &d, err
}

func (r *DescriptionRepo) DeleteByImageID(ctx context.Context, imageID string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM descriptions WHERE image_id = $1`, imageID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}