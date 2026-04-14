package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Env           string
	HTTPAddr      string
	HTTPReadTO    time.Duration
	HTTPWriteTO   time.Duration
	HTTPIdleTO    time.Duration
	ShutdownTO    time.Duration
	PostgresDSN   string
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	// JWT
	JWTAccessSecret  string
	JWTRefreshSecret string
	JWTAccessTTL     time.Duration
	JWTRefreshTTL    time.Duration
	// 连接池（单机 ~5000 并发读多写少场景：调大 PG/Redis 池，按机器内存与 max_connections 再压测）
	PGMaxConns       int32
	PGMinConns       int32
	PGMaxConnLife    time.Duration
	PGMaxConnIdle    time.Duration
	RedisPoolSize    int
	RedisMinIdle     int
	RedisPoolTO      time.Duration
	// 限流：每 IP 每秒 token（可按网关/CDN 前置再调大）
	RateLimitRPS float64
	RateBurst    int
	// 缓存 TTL
	TeaDetailCacheTTL time.Duration
	TeaListCacheTTL   time.Duration
	LeaderboardTTL    time.Duration
}

func getenv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func mustAtoi(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func mustParseDur(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}

func mustParseFloat(key string, def float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return def
	}
	return f
}

func Load() *Config {
	return &Config{
		Env:              getenv("APP_ENV", "dev"),
		HTTPAddr:         getenv("HTTP_ADDR", ":8080"),
		HTTPReadTO:       mustParseDur("HTTP_READ_TIMEOUT", 15*time.Second),
		HTTPWriteTO:      mustParseDur("HTTP_WRITE_TIMEOUT", 15*time.Second),
		HTTPIdleTO:       mustParseDur("HTTP_IDLE_TIMEOUT", 120*time.Second),
		ShutdownTO:       mustParseDur("HTTP_SHUTDOWN_TIMEOUT", 30*time.Second),
		PostgresDSN:      getenv("POSTGRES_DSN", "postgres://milktea:milktea@127.0.0.1:5432/milktea?sslmode=disable"),
		RedisAddr:        getenv("REDIS_ADDR", "127.0.0.1:6379"),
		RedisPassword:    getenv("REDIS_PASSWORD", ""),
		RedisDB:          mustAtoi("REDIS_DB", 0),
		JWTAccessSecret:  getenv("JWT_ACCESS_SECRET", "change-me-access-secret-32bytes!!"),
		JWTRefreshSecret: getenv("JWT_REFRESH_SECRET", "change-me-refresh-secret-32bytes!"),
		JWTAccessTTL:     mustParseDur("JWT_ACCESS_TTL", 15*time.Minute),
		JWTRefreshTTL:    mustParseDur("JWT_REFRESH_TTL", 168*time.Hour),
		PGMaxConns:       int32(mustAtoi("PG_MAX_CONNS", 120)),
		PGMinConns:       int32(mustAtoi("PG_MIN_CONNS", 10)),
		PGMaxConnLife:    mustParseDur("PG_MAX_CONN_LIFETIME", time.Hour),
		PGMaxConnIdle:    mustParseDur("PG_MAX_CONN_IDLE", 15*time.Minute),
		RedisPoolSize:    mustAtoi("REDIS_POOL_SIZE", 200),
		RedisMinIdle:     mustAtoi("REDIS_MIN_IDLE", 20),
		RedisPoolTO:      mustParseDur("REDIS_POOL_TIMEOUT", 5*time.Second),
		RateLimitRPS:     mustParseFloat("RATE_LIMIT_RPS", 200),
		RateBurst:        mustAtoi("RATE_LIMIT_BURST", 400),
		TeaDetailCacheTTL: mustParseDur("CACHE_TEA_DETAIL_TTL", 5*time.Minute),
		TeaListCacheTTL:   mustParseDur("CACHE_TEA_LIST_TTL", 2*time.Minute),
		LeaderboardTTL:    mustParseDur("CACHE_LEADERBOARD_TTL", 1*time.Minute),
	}
}
