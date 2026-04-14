package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	cachepkg "github.com/Junze888/milk_tea_go/internal/cache"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func (h *API) ListTeas(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	ps, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if ps < 1 || ps > 100 {
		ps = 20
	}
	offset := (page - 1) * ps

	var cat *int64
	if s := c.Query("category_id"); s != "" {
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad category_id"})
			return
		}
		cat = &v
	}

	ctx := c.Request.Context()
	catKey := c.Query("category_id")
	if catKey == "" {
		catKey = "all"
	}
	cacheKey := cachepkg.KeyTeaList(catKey, page, ps)
	if h.Redis != nil {
		if s, err := h.Redis.Get(ctx, cacheKey).Result(); err == nil && s != "" {
			c.Data(http.StatusOK, "application/json; charset=utf-8", []byte(s))
			return
		}
	}

	list, err := h.Teas.List(ctx, cat, offset, ps)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db"})
		return
	}
	payload := gin.H{"items": list, "page": page, "page_size": ps}
	raw, _ := json.Marshal(payload)
	if h.Redis != nil {
		_ = cachepkg.SetJSON(ctx, h.Redis, cacheKey, string(raw), h.Cfg.TeaListCacheTTL)
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", raw)
}

func (h *API) GetTea(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	ctx := c.Request.Context()
	key := cachepkg.KeyTeaDetail(id)
	if h.Redis != nil {
		if s, err := h.Redis.Get(ctx, key).Result(); err == nil && s != "" {
			go func(tid int64) {
				_ = h.Teas.IncrView(context.Background(), tid)
			}(id)
			c.Data(http.StatusOK, "application/json; charset=utf-8", []byte(s))
			return
		}
	}

	t, err := h.Teas.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db"})
		return
	}
	_ = h.Teas.IncrView(ctx, id)

	raw, _ := json.Marshal(t)
	if h.Redis != nil {
		_ = cachepkg.SetJSON(ctx, h.Redis, key, string(raw), h.Cfg.TeaDetailCacheTTL)
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", raw)
}

func (h *API) ListTeaReviews(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	ps, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if ps < 1 || ps > 100 {
		ps = 20
	}
	offset := (page - 1) * ps

	list, err := h.Reviews.ListByTea(c.Request.Context(), id, offset, ps)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list, "page": page, "page_size": ps})
}
