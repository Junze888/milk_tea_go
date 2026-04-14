package db

import (
	"context"
	"fmt"

	"github.com/Junze888/milk_tea_go/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	pcfg, err := pgxpool.ParseConfig(cfg.PostgresDSN)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}
	pcfg.MaxConns = cfg.PGMaxConns
	pcfg.MinConns = cfg.PGMinConns
	pcfg.MaxConnLifetime = cfg.PGMaxConnLife
	pcfg.MaxConnIdleTime = cfg.PGMaxConnIdle
	pool, err := pgxpool.NewWithConfig(ctx, pcfg)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return pool, nil
}
