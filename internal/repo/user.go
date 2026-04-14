package repo

import (
	"context"
	"errors"
	"net/netip"

	"github.com/Junze888/milk_tea_go/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}

func (r *UserRepo) Create(ctx context.Context, username, email, hash, nickname string) (int64, error) {
	const q = `
INSERT INTO users (username, email, password_hash, nickname)
VALUES ($1,$2,$3,$4)
RETURNING id`
	var id int64
	err := r.pool.QueryRow(ctx, q, username, email, hash, nickname).Scan(&id)
	return id, err
}

func (r *UserRepo) ByUsername(ctx context.Context, username string) (*model.User, string, error) {
	const q = `
SELECT id, username, email, password_hash, nickname, avatar_url, status, last_login_at, created_at
FROM users WHERE username=$1`
	var u model.User
	var hash string
	err := r.pool.QueryRow(ctx, q, username).Scan(
		&u.ID, &u.Username, &u.Email, &hash, &u.Nickname, &u.AvatarURL, &u.Status, &u.LastLoginAt, &u.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, "", ErrUserNotFound
	}
	return &u, hash, err
}

func (r *UserRepo) ByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)`, email).Scan(&exists)
	return exists, err
}

func (r *UserRepo) ByID(ctx context.Context, id int64) (*model.User, error) {
	const q = `
SELECT id, username, email, nickname, avatar_url, status, last_login_at, created_at
FROM users WHERE id=$1`
	var u model.User
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&u.ID, &u.Username, &u.Email, &u.Nickname, &u.AvatarURL, &u.Status, &u.LastLoginAt, &u.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	return &u, err
}

func (r *UserRepo) TouchLogin(ctx context.Context, id int64, ip netip.Addr) error {
	if ip.IsValid() {
		_, err := r.pool.Exec(ctx, `UPDATE users SET last_login_at=NOW(), last_login_ip=$2::inet, updated_at=NOW() WHERE id=$1`, id, ip.String())
		return err
	}
	_, err := r.pool.Exec(ctx, `UPDATE users SET last_login_at=NOW(), updated_at=NOW() WHERE id=$1`, id)
	return err
}
