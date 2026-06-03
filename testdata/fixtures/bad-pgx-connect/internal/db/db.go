package db

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func Open(ctx context.Context) (*pgx.Conn, error) {
	return pgx.Connect(ctx, "postgres://localhost/app")
}
