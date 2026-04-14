package handler

import (
	"github.com/Junze888/milk_tea_go/internal/config"
	"github.com/Junze888/milk_tea_go/internal/middleware"
	"github.com/gin-gonic/gin"
)

func (h *API) Mount(r *gin.Engine, cfg *config.Config) {
	r.Use(middleware.CORS())
	r.GET("/healthz", h.Healthz)
	r.GET("/readyz", h.Readyz)

	v1 := r.Group("/api/v1")
	v1.Use(middleware.RateLimit(cfg))

	auth := v1.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.Refresh)
		auth.POST("/logout", h.Logout)
	}

	v1.GET("/categories", h.ListCategories)
	v1.GET("/teas", h.ListTeas)
	v1.GET("/teas/:id", h.GetTea)
	v1.GET("/teas/:id/reviews", h.ListTeaReviews)
	v1.GET("/leaderboard", h.Leaderboard)

	need := v1.Group("")
	need.Use(middleware.JWTAuth(cfg))
	{
		need.GET("/me", h.Me)
		need.POST("/teas/:id/reviews", h.UpsertReview)
	}
}
