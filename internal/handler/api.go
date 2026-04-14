package handler

import (
	"github.com/Junze888/milk_tea_go/internal/config"
	"github.com/Junze888/milk_tea_go/internal/repo"
	"github.com/redis/go-redis/v9"
)

type API struct {
	Cfg     *config.Config
	Users   *repo.UserRepo
	Tokens  *repo.RefreshTokenRepo
	Cats    *repo.CategoryRepo
	Teas    *repo.TeaRepo
	Reviews *repo.ReviewRepo
	Redis   *redis.Client
}

func NewAPI(
	cfg *config.Config,
	u *repo.UserRepo,
	t *repo.RefreshTokenRepo,
	c *repo.CategoryRepo,
	te *repo.TeaRepo,
	rv *repo.ReviewRepo,
	rdb *redis.Client,
) *API {
	return &API{
		Cfg: cfg, Users: u, Tokens: t, Cats: c, Teas: te, Reviews: rv, Redis: rdb,
	}
}
