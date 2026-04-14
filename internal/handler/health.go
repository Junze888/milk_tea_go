package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *API) Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *API) Readyz(c *gin.Context) {
	ctx := c.Request.Context()
	if err := h.Users.Ping(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"postgres": err.Error()})
		return
	}
	if err := h.Redis.Ping(ctx).Err(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"redis": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"postgres": "ok", "redis": "ok"})
}
