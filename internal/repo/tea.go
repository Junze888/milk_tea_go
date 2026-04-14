package repo

import (
	"context"
	"encoding/json"

	"github.com/Junze888/milk_tea_go/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TeaRepo struct {
	pool *pgxpool.Pool
}

func NewTeaRepo(pool *pgxpool.Pool) *TeaRepo {
	return &TeaRepo{pool: pool}
}

func (r *TeaRepo) List(ctx context.Context, categoryID *int64, offset, limit int) ([]model.Tea, error) {
	var (
		rows pgx.Rows
		err  error
	)
	if categoryID != nil {
		rows, err = r.pool.Query(ctx, `
SELECT t.id, t.category_id, t.name, t.shop_name, t.subtitle, t.description, t.image_url, t.tags,
       t.price_min_cent, t.price_max_cent, t.currency, t.sugar_level, t.is_recommended, t.status, t.view_count,
       COALESCE(s.avg_rating, 0), COALESCE(s.review_count, 0), t.created_at
FROM teas t
LEFT JOIN tea_stats s ON s.tea_id = t.id
WHERE t.status=1 AND t.category_id=$1
ORDER BY t.is_recommended DESC, t.id DESC
OFFSET $2 LIMIT $3`, *categoryID, offset, limit)
	} else {
		rows, err = r.pool.Query(ctx, `
SELECT t.id, t.category_id, t.name, t.shop_name, t.subtitle, t.description, t.image_url, t.tags,
       t.price_min_cent, t.price_max_cent, t.currency, t.sugar_level, t.is_recommended, t.status, t.view_count,
       COALESCE(s.avg_rating, 0), COALESCE(s.review_count, 0), t.created_at
FROM teas t
LEFT JOIN tea_stats s ON s.tea_id = t.id
WHERE t.status=1
ORDER BY t.is_recommended DESC, t.id DESC
OFFSET $1 LIMIT $2`, offset, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTeas(rows)
}

func scanTeas(rows pgx.Rows) ([]model.Tea, error) {
	var out []model.Tea
	for rows.Next() {
		var t model.Tea
		var tags []string
		if err := rows.Scan(
			&t.ID, &t.CategoryID, &t.Name, &t.ShopName, &t.Subtitle, &t.Description, &t.ImageURL, &tags,
			&t.PriceMinCent, &t.PriceMaxCent, &t.Currency, &t.SugarLevel, &t.IsRecommended, &t.Status, &t.ViewCount,
			&t.AvgRating, &t.ReviewCount, &t.CreatedAt,
		); err != nil {
			return nil, err
		}
		t.Tags = tags
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *TeaRepo) GetByID(ctx context.Context, id int64) (*model.Tea, error) {
	row := r.pool.QueryRow(ctx, `
SELECT t.id, t.category_id, t.name, t.shop_name, t.subtitle, t.description, t.image_url, t.tags,
       t.price_min_cent, t.price_max_cent, t.currency, t.sugar_level, t.is_recommended, t.status, t.view_count,
       COALESCE(s.avg_rating, 0), COALESCE(s.review_count, 0), t.created_at
FROM teas t
LEFT JOIN tea_stats s ON s.tea_id = t.id
WHERE t.id=$1 AND t.status=1`, id)
	var t model.Tea
	var tags []string
	err := row.Scan(
		&t.ID, &t.CategoryID, &t.Name, &t.ShopName, &t.Subtitle, &t.Description, &t.ImageURL, &tags,
		&t.PriceMinCent, &t.PriceMaxCent, &t.Currency, &t.SugarLevel, &t.IsRecommended, &t.Status, &t.ViewCount,
		&t.AvgRating, &t.ReviewCount, &t.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	t.Tags = tags
	return &t, nil
}

func (r *TeaRepo) IncrView(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `UPDATE teas SET view_count = view_count + 1, updated_at=NOW() WHERE id=$1`, id)
	return err
}

// TeaJSON 用于 Redis 缓存序列化
func TeaToJSON(t *model.Tea) ([]byte, error) {
	return json.Marshal(t)
}

func TeaFromJSON(b []byte) (*model.Tea, error) {
	var t model.Tea
	err := json.Unmarshal(b, &t)
	return &t, err
}
