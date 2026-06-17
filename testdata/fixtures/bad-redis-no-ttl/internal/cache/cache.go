package cache

import (
	"context"

	redis "github.com/redis/go-redis/v9"
)

func Save(ctx context.Context, rdb *redis.Client, key, value string) error {
	return rdb.Set(ctx, key, value, 0).Err()
}
