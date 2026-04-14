package repo

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RefreshTokenRepo struct {
	pool *pgxpool.Pool
}

func NewRefreshTokenRepo(pool *pgxpool.Pool) *RefreshTokenRepo {
	return &RefreshTokenRepo{pool: pool}
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func (r *RefreshTokenRepo) Insert(ctx context.Context, userID int64, jti, familyID uuid.UUID, tokenRaw string, exp time.Time, ip, ua, device string) error {
	const q = `
INSERT INTO refresh_tokens (user_id, jti, token_hash, family_id, expires_at, ip, user_agent, device_id)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`
	h := hashToken(tokenRaw)
	return execErr(r.pool.Exec(ctx, q, userID, jti, h, familyID, exp, nullableIP(ip), ua, device))
}

func nullableIP(ip string) interface{} {
	if ip == "" {
		return nil
	}
	return ip
}

func execErr(_ interface{}, err error) error {
	return err
}

func (r *RefreshTokenRepo) RevokeJTI(ctx context.Context, jti uuid.UUID, reason string) error {
	_, err := r.pool.Exec(ctx, `UPDATE refresh_tokens SET revoked_at=NOW(), revoked_reason=$2 WHERE jti=$1 AND revoked_at IS NULL`, jti, reason)
	return err
}

func (r *RefreshTokenRepo) RevokeFamily(ctx context.Context, family uuid.UUID, reason string) error {
	_, err := r.pool.Exec(ctx, `UPDATE refresh_tokens SET revoked_at=NOW(), revoked_reason=$2 WHERE family_id=$1 AND revoked_at IS NULL`, family, reason)
	return err
}

// ValidByRaw 校验 refresh token 原文是否在库中且未吊销
func (r *RefreshTokenRepo) ValidByRaw(ctx context.Context, tokenRaw string, jti uuid.UUID) (userID int64, familyID uuid.UUID, ok bool, err error) {
	h := hashToken(tokenRaw)
	const q = `
SELECT user_id, family_id FROM refresh_tokens
WHERE jti=$1 AND token_hash=$2 AND revoked_at IS NULL AND expires_at > NOW()`
	err = r.pool.QueryRow(ctx, q, jti, h).Scan(&userID, &familyID)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, uuid.Nil, false, nil
	}
	if err != nil {
		return 0, uuid.Nil, false, err
	}
	return userID, familyID, true, nil
}
