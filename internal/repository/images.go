package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourname/novapromptgobackend/internal/models"
)

var ErrNotFound = errors.New("not found")

type ImageRepo struct {
	pool *pgxpool.Pool
}

func NewImageRepo(pool *pgxpool.Pool) *ImageRepo {
	return &ImageRepo{pool: pool}
}

func (r *ImageRepo) Create(ctx context.Context, img *models.Image) error {
	const q = `
		INSERT INTO images (title, cloudinary_url, category_id)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`
	return r.pool.QueryRow(ctx, q, img.Title, img.CloudinaryURL, img.CategoryID).
		Scan(&img.ID, &img.CreatedAt, &img.UpdatedAt)
}

func (r *ImageRepo) GetByID(ctx context.Context, id string) (*models.Image, error) {
	const q = `
		SELECT id, title, cloudinary_url, category_id, created_at, updated_at
		FROM images WHERE id = $1`
	var img models.Image
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&img.ID, &img.Title, &img.CloudinaryURL, &img.CategoryID, &img.CreatedAt, &img.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get image: %w", err)
	}
	return &img, nil
}

func (r *ImageRepo) List(ctx context.Context, limit, offset int) ([]models.Image, error) {
	const q = `
		SELECT id, title, cloudinary_url, category_id, created_at, updated_at
		FROM images
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`
	rows, err := r.pool.Query(ctx, q, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list images: %w", err)
	}
	defer rows.Close()

	var out []models.Image
	for rows.Next() {
		var img models.Image
		if err := rows.Scan(&img.ID, &img.Title, &img.CloudinaryURL, &img.CategoryID, &img.CreatedAt, &img.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, img)
	}
	return out, rows.Err()
}

func (r *ImageRepo) Update(ctx context.Context, img *models.Image) error {
	const q = `
		UPDATE images
		SET title = $1, cloudinary_url = $2, category_id = $3
		WHERE id = $4
		RETURNING updated_at`
	err := r.pool.QueryRow(ctx, q, img.Title, img.CloudinaryURL, img.CategoryID, img.ID).
		Scan(&img.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func (r *ImageRepo) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM images WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete image: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}