package handler

import (
	"context"

	redis "github.com/redis/go-redis/v9"
)

func SessionFromCache(ctx context.Context, rdb *redis.Client, token string) (bool, error) {
	_, err := rdb.Get(ctx, "session:"+token).Result()
	return err == nil, err
}
