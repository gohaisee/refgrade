package repo

import "github.com/jackc/pgx/v5"

type Repo struct{}

func New() *Repo { return &Repo{} }

func (r *Repo) Ping(db *pgx.Conn) error { return nil }
