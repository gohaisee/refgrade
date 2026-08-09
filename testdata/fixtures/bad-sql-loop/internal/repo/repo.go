package repo

import (
	"context"

	_ "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Activate(ctx context.Context, pool *pgxpool.Pool, ids []int) error {
	for _, id := range ids {
		_, err := pool.Exec(ctx, "UPDATE users SET active = true WHERE id = $1", id)
		if err != nil {
			return err
		}
	}
	return nil
}
