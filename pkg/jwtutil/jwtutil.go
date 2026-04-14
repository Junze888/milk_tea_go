package jwtutil

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid token")
)

type AccessClaims struct {
	UserID   int64  `json:"uid"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type RefreshClaims struct {
	UserID   int64  `json:"uid"`
	JTI      string `json:"jti"`
	FamilyID string `json:"fid"`
	jwt.RegisteredClaims
}

func SignAccess(secret string, userID int64, username string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := AccessClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			ID:        uuid.NewString(),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(secret))
}

func ParseAccess(secret, token string) (*AccessClaims, error) {
	var claims AccessClaims
	t, err := jwt.ParseWithClaims(token, &claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil || !t.Valid {
		return nil, ErrInvalidToken
	}
	return &claims, nil
}

func SignRefresh(secret string, userID int64, jti, familyID uuid.UUID, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := RefreshClaims{
		UserID:   userID,
		JTI:      jti.String(),
		FamilyID: familyID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			ID:        jti.String(),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(secret))
}

func ParseRefresh(secret, token string) (*RefreshClaims, error) {
	var claims RefreshClaims
	t, err := jwt.ParseWithClaims(token, &claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil || !t.Valid {
		return nil, ErrInvalidToken
	}
	return &claims, nil
}
