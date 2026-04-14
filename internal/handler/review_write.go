package handler

import (
	"errors"
	"net/http"
	"strconv"

	cachepkg "github.com/Junze888/milk_tea_go/internal/cache"
	"github.com/Junze888/milk_tea_go/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type reviewBody struct {
	Stars   int16  `json:"stars" binding:"required,min=1,max=5"`
	Title   string `json:"title" binding:"max=128"`
	Comment string `json:"comment" binding:"max=2000"`
}

func (h *API) UpsertReview(c *gin.Context) {
	teaID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || teaID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad tea id"})
		return
	}
	uid, ok := middleware.MustUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var body reviewBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := c.Request.Context()
	if _, err := h.Teas.GetByID(ctx, teaID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "tea not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db"})
		return
	}
	if err := h.Reviews.Upsert(ctx, teaID, uid, body.Stars, body.Title, body.Comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db"})
		return
	}
	if h.Redis != nil {
		_ = cachepkg.InvalidateTeaCache(ctx, h.Redis, teaID)
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
