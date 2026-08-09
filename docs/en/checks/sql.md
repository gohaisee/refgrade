# sql and relational data

**n/a** when no sql driver or orm in module

## detected stack

| tool | import / signal |
|------|-----------------|
| pgx v5 | `github.com/jackc/pgx/v5` |
| gorm | `gorm.io/gorm` |
| sqlx | `github.com/jmoiron/sqlx` |
| sqlc | `github.com/sqlc-dev/sqlc` (tool / generated queries) |
| ent | `entgo.io/ent` |
| pq (legacy pg) | `github.com/lib/pq` |
| mysql | `github.com/go-sql-driver/mysql` |
| sqlite | `modernc.org/sqlite`, `github.com/mattn/go-sqlite3` |

## universal sql checks

| id | when | why | fix | severity |
|----|------|-----|-----|----------|
| sql-01 | `Query`/`Exec`/orm call inside `for` loop | query in a loop on lists | batch `WHERE id = ANY($1)` or join | warn |
| sql-02 | sql built with `fmt.Sprintf` and user input | injection | parameterized queries only | fail |
| sql-03 | function named `List*` / `FindAll*` without limit in sql | full table scan | hard `LIMIT`; keyset pagination | warn |
| sql-04 | `SELECT *` in hot path (heuristic) | over-fetch | list columns explicitly | info |
| sql-05 | new pool/connection per request | exhaust fds | single pool per process | fail |
| sql-06 | `float64` for money fields | rounding bugs | integer minor units or decimal type | warn |
| sql-07 | offset pagination on large tables (comment/heuristic) | slow pages | keyset (`WHERE id > $cursor`) | info |

## pgx

| id | when | fix | severity |
|----|------|-----|----------|
| pgx-01 | `pgx.Connect` instead of pool in app | use `pgxpool.New` | warn |
| pgx-02 | compare pg errors by string | use `pgconn.PgError` codes | info |
| pgx-03 | `sslmode=disable` in dsn for remote host | enable tls | fail |

## gorm

| id | when | fix | severity |
|----|------|-----|----------|
| gorm-01 | `db.Raw("... "+userInput)` | bind parameters | fail |
| gorm-02 | deep `Preload` chains on list endpoints | join or batch | warn |
| gorm-03 | `AutoMigrate` in `main` for prod apps | goose/golang-migrate | warn |
| gorm-04 | `Debug()` enabled in non-dev | remove debug | warn |
| gorm-05 | package-level `var DB *gorm.DB` | inject via constructor | warn |

## sqlx

| id | when | fix |
|----|------|-----|
| sqlx-01 | `Get`/`Select` without context | pass `ctx` |

## sqlc / ent

| id | when | fix |
|----|------|-----|
| sqlc-01 | hand-edited `*.sql.go` | regenerate from sql |
| ent-01 | edge traversal in loop without eager load | batch query or eager edge |

## security

see [security-owasp.md](security-owasp.md) — sec-db01…sec-db05
