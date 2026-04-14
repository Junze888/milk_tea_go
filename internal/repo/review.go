package repo

import (
	"context"

	"github.com/Junze888/milk_tea_go/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReviewRepo struct {
	pool *pgxpool.Pool
}

func NewReviewRepo(pool *pgxpool.Pool) *ReviewRepo {
	return &ReviewRepo{pool: pool}
}

func (r *ReviewRepo) ListByTea(ctx context.Context, teaID int64, offset, limit int) ([]model.Review, error) {
	rows, err := r.pool.Query(ctx, `
SELECT r.id, r.tea_id, r.user_id, u.nickname, r.stars, r.title, r.comment, r.created_at
FROM reviews r
JOIN users u ON u.id = r.user_id
WHERE r.tea_id=$1 AND r.status=1
ORDER BY r.created_at DESC
OFFSET $2 LIMIT $3`, teaID, offset, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Review
	for rows.Next() {
		var rv model.Review
		if err := rows.Scan(&rv.ID, &rv.TeaID, &rv.UserID, &rv.UserName, &rv.Stars, &rv.Title, &rv.Comment, &rv.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, rv)
	}
	return out, rows.Err()
}

func (r *ReviewRepo) Upsert(ctx context.Context, teaID, userID int64, stars int16, title, comment string) error {
	const q = `
INSERT INTO reviews (tea_id, user_id, stars, title, comment, status)
VALUES ($1,$2,$3,$4,$5,1)
ON CONFLICT (tea_id, user_id) DO UPDATE SET
  stars = EXCLUDED.stars,
  title = EXCLUDED.title,
  comment = EXCLUDED.comment,
  updated_at = NOW()`
	_, err := r.pool.Exec(ctx, q, teaID, userID, stars, title, comment)
	return err
}

func (r *ReviewRepo) Leaderboard(ctx context.Context, limit int) ([]model.LeaderboardRow, error) {
	rows, err := r.pool.Query(ctx, `
SELECT t.id, t.name, t.shop_name, COALESCE(s.avg_rating, 0), COALESCE(s.review_count, 0)
FROM tea_stats s
JOIN teas t ON t.id = s.tea_id
WHERE t.status = 1 AND s.review_count > 0
ORDER BY s.avg_rating DESC, s.review_count DESC
LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.LeaderboardRow
	rank := 1
	for rows.Next() {
		var row model.LeaderboardRow
		if err := rows.Scan(&row.TeaID, &row.Name, &row.ShopName, &row.AvgRating, &row.ReviewCount); err != nil {
			return nil, err
		}
		row.Rank = rank
		rank++
		out = append(out, row)
	}
	return out, rows.Err()
}
