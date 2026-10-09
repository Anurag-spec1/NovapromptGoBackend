package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourname/novapromptgobackend/internal/models"
)

type CategoryRepo struct{ pool *pgxpool.Pool }

func NewCategoryRepo(pool *pgxpool.Pool) *CategoryRepo { return &CategoryRepo{pool: pool} }

func (r *CategoryRepo) Create(ctx context.Context, c *models.Category) error {
	const q = `INSERT INTO categories (name, slug) VALUES ($1, $2)
	           RETURNING id, created_at`
	return r.pool.QueryRow(ctx, q, c.Name, c.Slug).Scan(&c.ID, &c.CreatedAt)
}

func (r *CategoryRepo) GetByID(ctx context.Context, id string) (*models.Category, error) {
	var c models.Category
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, slug, created_at FROM categories WHERE id = $1`, id).
		Scan(&c.ID, &c.Name, &c.Slug, &c.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &c, err
}

func (r *CategoryRepo) List(ctx context.Context) ([]models.Category, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, slug, created_at FROM categories ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	defer rows.Close()
	var out []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *CategoryRepo) Update(ctx context.Context, c *models.Category) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE categories SET name = $1, slug = $2 WHERE id = $3`,
		c.Name, c.Slug, c.ID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *CategoryRepo) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM categories WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}