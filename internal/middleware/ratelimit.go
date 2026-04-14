package middleware

import (
	"net/http"
	"sync"

	"github.com/Junze888/milk_tea_go/internal/config"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type ipLimiter struct {
	mu sync.Mutex
	m  map[string]*rate.Limiter
	r  rate.Limit
	b  int
}

func newIPLimiter(rps float64, burst int) *ipLimiter {
	return &ipLimiter{
		m: make(map[string]*rate.Limiter),
		r: rate.Limit(rps),
		b: burst,
	}
}

func (il *ipLimiter) get(ip string) *rate.Limiter {
	il.mu.Lock()
	defer il.mu.Unlock()
	l, ok := il.m[ip]
	if ok {
		return l
	}
	l = rate.NewLimiter(il.r, il.b)
	il.m[ip] = l
	return l
}

// RateLimit 按客户端 IP 限流（单机高并发下保护应用；前置 Nginx/网关可再叠一层）
func RateLimit(cfg *config.Config) gin.HandlerFunc {
	il := newIPLimiter(cfg.RateLimitRPS, cfg.RateBurst)
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !il.get(ip).Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit"})
			return
		}
		c.Next()
	}
}
