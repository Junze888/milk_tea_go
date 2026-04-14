package repo

import (
	"context"

	"github.com/Junze888/milk_tea_go/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CategoryRepo struct {
	pool *pgxpool.Pool
}

func NewCategoryRepo(pool *pgxpool.Pool) *CategoryRepo {
	return &CategoryRepo{pool: pool}
}

func (r *CategoryRepo) ListVisible(ctx context.Context) ([]model.Category, error) {
	rows, err := r.pool.Query(ctx, `
SELECT id, name, slug, description, icon_url, cover_url, sort_order, is_hot, status, created_at
FROM categories WHERE status=1 ORDER BY sort_order DESC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Category
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &c.IconURL, &c.CoverURL, &c.SortOrder, &c.IsHot, &c.Status, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
