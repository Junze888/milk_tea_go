package middleware

import (
	"net/http"
	"strings"

	"github.com/Junze888/milk_tea_go/internal/config"
	"github.com/Junze888/milk_tea_go/pkg/jwtutil"
	"github.com/gin-gonic/gin"
)

const (
	CtxUserIDKey   = "uid"
	CtxUsernameKey = "username"
)

func JWTAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if h == "" || !strings.HasPrefix(strings.ToLower(h), "bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		raw := strings.TrimSpace(h[7:])
		claims, err := jwtutil.ParseAccess(cfg.JWTAccessSecret, raw)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid access token"})
			return
		}
		c.Set(CtxUserIDKey, claims.UserID)
		c.Set(CtxUsernameKey, claims.Username)
		c.Next()
	}
}

func MustUserID(c *gin.Context) (int64, bool) {
	v, ok := c.Get(CtxUserIDKey)
	if !ok {
		return 0, false
	}
	id, ok := v.(int64)
	return id, ok
}
