package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *API) ListCategories(c *gin.Context) {
	list, err := h.Cats.ListVisible(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}
