package handler

import (
	"net"
	"net/http"
	"net/netip"
	"strings"
	"time"

	"github.com/Junze888/milk_tea_go/internal/middleware"
	"github.com/Junze888/milk_tea_go/internal/repo"
	"github.com/Junze888/milk_tea_go/pkg/jwtutil"
	"github.com/Junze888/milk_tea_go/pkg/password"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type registerReq struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=72"`
	Nickname string `json:"nickname" binding:"max=64"`
}

type loginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type refreshReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type logoutReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *API) Register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := c.Request.Context()
	exists, err := h.Users.ByEmail(ctx, strings.ToLower(strings.TrimSpace(req.Email)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db"})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		return
	}
	hash, err := password.Hash(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "hash"})
		return
	}
	nick := strings.TrimSpace(req.Nickname)
	if nick == "" {
		nick = req.Username
	}
	id, err := h.Users.Create(ctx, strings.TrimSpace(req.Username), strings.ToLower(strings.TrimSpace(req.Email)), hash, nick)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			c.JSON(http.StatusConflict, gin.H{"error": "username or email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create user"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "username": req.Username})
}

func (h *API) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := c.Request.Context()
	u, hash, err := h.Users.ByUsername(ctx, strings.TrimSpace(req.Username))
	if err != nil {
		if err == repo.ErrUserNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db"})
		return
	}
	if u.Status != 1 {
		c.JSON(http.StatusForbidden, gin.H{"error": "user disabled"})
		return
	}
	if !password.Verify(hash, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	ip := parseClientIP(c.ClientIP())
	_ = h.Users.TouchLogin(ctx, u.ID, ip)

	access, err := jwtutil.SignAccess(h.Cfg.JWTAccessSecret, u.ID, u.Username, h.Cfg.JWTAccessTTL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token"})
		return
	}
	family := uuid.New()
	jti := uuid.New()
	refreshRaw, err := jwtutil.SignRefresh(h.Cfg.JWTRefreshSecret, u.ID, jti, family, h.Cfg.JWTRefreshTTL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token"})
		return
	}
	exp := time.Now().Add(h.Cfg.JWTRefreshTTL)
	if err := h.Tokens.Insert(ctx, u.ID, jti, family, refreshRaw, exp, c.ClientIP(), c.GetHeader("User-Agent"), ""); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"access_token":  access,
		"refresh_token": refreshRaw,
		"expires_in":    int(h.Cfg.JWTAccessTTL.Seconds()),
		"token_type":    "Bearer",
		"user": gin.H{
			"id": u.ID, "username": u.Username, "nickname": u.Nickname, "avatar_url": u.AvatarURL,
		},
	})
}

func (h *API) Refresh(c *gin.Context) {
	var req refreshReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := c.Request.Context()
	claims, err := jwtutil.ParseRefresh(h.Cfg.JWTRefreshSecret, req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}
	jti, err := uuid.Parse(claims.JTI)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}
	uid, familyID, ok, err := h.Tokens.ValidByRaw(ctx, req.RefreshToken, jti)
	if err != nil || !ok || uid != claims.UserID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh revoked or expired"})
		return
	}
	_ = h.Tokens.RevokeJTI(ctx, jti, "rotated")

	u, err := h.Users.ByID(ctx, uid)
	if err != nil || u.Status != 1 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}
	access, err := jwtutil.SignAccess(h.Cfg.JWTAccessSecret, u.ID, u.Username, h.Cfg.JWTAccessTTL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token"})
		return
	}
	newJTI := uuid.New()
	refreshRaw, err := jwtutil.SignRefresh(h.Cfg.JWTRefreshSecret, u.ID, newJTI, familyID, h.Cfg.JWTRefreshTTL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token"})
		return
	}
	exp := time.Now().Add(h.Cfg.JWTRefreshTTL)
	if err := h.Tokens.Insert(ctx, u.ID, newJTI, familyID, refreshRaw, exp, c.ClientIP(), c.GetHeader("User-Agent"), ""); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"access_token":  access,
		"refresh_token": refreshRaw,
		"expires_in":    int(h.Cfg.JWTAccessTTL.Seconds()),
		"token_type":    "Bearer",
	})
}

func (h *API) Logout(c *gin.Context) {
	var req logoutReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := c.Request.Context()
	claims, err := jwtutil.ParseRefresh(h.Cfg.JWTRefreshSecret, req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}
	jti, err := uuid.Parse(claims.JTI)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}
	uid, _, ok, err := h.Tokens.ValidByRaw(ctx, req.RefreshToken, jti)
	if err != nil || !ok || uid != claims.UserID {
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}
	_ = h.Tokens.RevokeJTI(ctx, jti, "logout")
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *API) Me(c *gin.Context) {
	id, ok := middleware.MustUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	u, err := h.Users.ByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, u)
}

func parseClientIP(s string) netip.Addr {
	s = strings.TrimSpace(s)
	if s == "" {
		return netip.Addr{}
	}
	if ip, err := netip.ParseAddr(s); err == nil {
		return ip
	}
	host, _, err := net.SplitHostPort(s)
	if err != nil {
		return netip.Addr{}
	}
	if ip, err := netip.ParseAddr(host); err == nil {
		return ip
	}
	return netip.Addr{}
}
