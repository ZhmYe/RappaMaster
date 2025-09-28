package redis

import (
	"RappaMaster/config"
	"RappaMaster/types"
	"context"
	"github.com/redis/go-redis/v9"
)

type RedisService struct {
	config.RedisConfig
	*redis.Client
}

func (rcs *RedisService) Init() error {
	rcs.Client = redis.NewClient(&redis.Options{
		Addr:     rcs.Addr,
		Password: rcs.Password,
		DB:       rcs.DB,
	})
	ctx := context.Background()
	_, err := rcs.Client.Ping(ctx).Result()
	if err != nil {
		return types.RaiseError(types.RedisError, "connect redis failed", err)
	}
	return nil
}

func NewRedisService(cfg config.RedisConfig) *RedisService {
	return &RedisService{
		RedisConfig: cfg,
		Client:      nil,
	}
}
