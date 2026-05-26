package config

import _ "github.com/jackc/pgx/v5"

const DSN = "postgres://user:pass@db.example.com:5432/app?sslmode=disable"
