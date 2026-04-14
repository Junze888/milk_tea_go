package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	cachepkg "github.com/Junze888/milk_tea_go/internal/cache"
	"github.com/gin-gonic/gin"
)

func (h *API) Leaderboard(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit < 1 || limit > 200 {
		limit = 50
	}
	ctx := c.Request.Context()
	key := cachepkg.KeyLeaderboard()
	if h.Redis != nil {
		if s, err := h.Redis.Get(ctx, key).Result(); err == nil && s != "" {
			c.Data(http.StatusOK, "application/json; charset=utf-8", []byte(s))
			return
		}
	}
	rows, err := h.Reviews.Leaderboard(ctx, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db"})
		return
	}
	payload := gin.H{"items": rows}
	raw, _ := json.Marshal(payload)
	if h.Redis != nil {
		_ = cachepkg.SetJSON(ctx, h.Redis, key, string(raw), h.Cfg.LeaderboardTTL)
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", raw)
}
