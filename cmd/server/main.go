package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Junze888/milk_tea_go/internal/cache"
	"github.com/Junze888/milk_tea_go/internal/config"
	"github.com/Junze888/milk_tea_go/internal/db"
	"github.com/Junze888/milk_tea_go/internal/handler"
	"github.com/Junze888/milk_tea_go/internal/migrate"
	"github.com/Junze888/milk_tea_go/internal/repo"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()
	if cfg.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	ctx := context.Background()

	pool, err := db.NewPool(ctx, cfg)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pool.Close()

	rdb := cache.NewRedis(cfg)
	defer func() { _ = rdb.Close() }()
	if err := cache.Ping(ctx, rdb); err != nil {
		log.Fatalf("redis: %v", err)
	}

	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "migrations"
	}
	if err := migrate.Run(ctx, pool, migrationsDir); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	users := repo.NewUserRepo(pool)
	tokens := repo.NewRefreshTokenRepo(pool)
	cats := repo.NewCategoryRepo(pool)
	teas := repo.NewTeaRepo(pool)
	reviews := repo.NewReviewRepo(pool)

	api := handler.NewAPI(cfg, users, tokens, cats, teas, reviews, rdb)

	r := gin.New()
	r.Use(gin.Recovery())
	api.Mount(r, cfg)

	srv := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      r,
		ReadTimeout:  cfg.HTTPReadTO,
		WriteTimeout: cfg.HTTPWriteTO,
		IdleTimeout:  cfg.HTTPIdleTO,
	}

	go func() {
		log.Printf("listening %s env=%s", cfg.HTTPAddr, cfg.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTO)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown: %v", err)
	}
}
