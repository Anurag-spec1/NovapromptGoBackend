package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourname/novapromptgobackend/internal/models"
)

type TagRepo struct{ pool *pgxpool.Pool }

func NewTagRepo(pool *pgxpool.Pool) *TagRepo { return &TagRepo{pool: pool} }

func (r *TagRepo) Create(ctx context.Context, t *models.Tag) error {
	const q = `INSERT INTO tags (name) VALUES ($1)
	           ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
	           RETURNING id`
	return r.pool.QueryRow(ctx, q, t.Name).Scan(&t.ID)
}

func (r *TagRepo) List(ctx context.Context) ([]models.Tag, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, name FROM tags ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}
	defer rows.Close()
	var out []models.Tag
	for rows.Next() {
		var t models.Tag
		if err := rows.Scan(&t.ID, &t.Name); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *TagRepo) Update(ctx context.Context, t *models.Tag) error {
	tag, err := r.pool.Exec(ctx, `UPDATE tags SET name = $1 WHERE id = $2`, t.Name, t.ID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *TagRepo) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM tags WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *TagRepo) AttachToImage(ctx context.Context, imageID, tagID string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO image_tags (image_id, tag_id) VALUES ($1, $2)
		 ON CONFLICT DO NOTHING`, imageID, tagID)
	return err
}

func (r *TagRepo) DetachFromImage(ctx context.Context, imageID, tagID string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM image_tags WHERE image_id = $1 AND tag_id = $2`, imageID, tagID)
	return err
}