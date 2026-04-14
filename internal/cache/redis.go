package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/Junze888/milk_tea_go/internal/config"
	"github.com/redis/go-redis/v9"
)

func NewRedis(cfg *config.Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:         cfg.RedisAddr,
		Password:     cfg.RedisPassword,
		DB:           cfg.RedisDB,
		PoolSize:     cfg.RedisPoolSize,
		MinIdleConns: cfg.RedisMinIdle,
		PoolTimeout:  cfg.RedisPoolTO,
	})
}

func Ping(ctx context.Context, c *redis.Client) error {
	if err := c.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("ping redis: %w", err)
	}
	return nil
}

func KeyTeaDetail(id int64) string  { return fmt.Sprintf("tea:detail:%d", id) }
func KeyTeaList(cat string, page, ps int) string {
	return fmt.Sprintf("tea:list:%s:%d:%d", cat, page, ps)
}
func KeyLeaderboard() string { return "leaderboard:teas" }

func InvalidateTeaCache(ctx context.Context, rdb *redis.Client, teaID int64) error {
	if err := rdb.Del(ctx, KeyTeaDetail(teaID)).Err(); err != nil {
		return err
	}
	if err := rdb.Del(ctx, KeyLeaderboard()).Err(); err != nil {
		return err
	}
	iter := rdb.Scan(ctx, 0, "tea:list:*", 500).Iterator()
	for iter.Next(ctx) {
		if err := rdb.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}

func SetJSON(ctx context.Context, rdb *redis.Client, key string, val string, ttl time.Duration) error {
	return rdb.Set(ctx, key, val, ttl).Err()
}
