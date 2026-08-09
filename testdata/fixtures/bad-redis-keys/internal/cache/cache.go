package cache

import (
	"context"

	redis "github.com/redis/go-redis/v9"
)

func Purge(ctx context.Context, rdb *redis.Client) error {
	return rdb.Keys(ctx, "*").Err()
}
