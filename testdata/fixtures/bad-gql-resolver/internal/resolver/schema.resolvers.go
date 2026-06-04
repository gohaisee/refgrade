package resolver

import "github.com/jackc/pgx/v5"

type Resolver struct {
	pool *pgx.Conn
}
