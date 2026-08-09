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

| id | status | when | why | fix | severity |
|----|--------|------|-----|-----|----------|
| sql-01 | implemented | `Query`/`Exec`/orm call inside `for` loop | query in a loop on lists | batch `WHERE id = ANY($1)` or join | warn |
| sql-02 | implemented | sql built with `fmt.Sprintf` and user input | injection | parameterized queries only | fail |
| sql-03 | implemented | function named `List*` / `FindAll*` without limit in sql | full table scan | hard `LIMIT`; keyset pagination | warn |
| sql-04 | implemented | `SELECT *` in hot path (heuristic) | over-fetch | list columns explicitly | info |
| sql-05 | implemented | new pool/connection per request | exhaust fds | single pool per process | fail |
| sql-06 | implemented | `float64` for money fields | rounding bugs | integer minor units or decimal type | warn |
| sql-07 | implemented | offset pagination on large tables (comment/heuristic) | slow pages | keyset (`WHERE id > $cursor`) | info |

## pgx

| id | status | when | fix | severity |
|----|--------|------|-----|----------|
| pgx-01 | implemented | `pgx.Connect` instead of pool in app | use `pgxpool.New` | warn |
| pgx-02 | implemented | compare pg errors by string | use `pgconn.PgError` codes | info |
| pgx-03 | implemented | `sslmode=disable` in dsn for remote host | enable tls | fail |

## gorm

| id | status | when | fix | severity |
|----|--------|------|-----|----------|
| gorm-01 | implemented | `db.Raw("... "+userInput)` | bind parameters | fail |
| gorm-02 | implemented | deep `Preload` chains on list endpoints | join or batch | warn |
| gorm-03 | implemented | `AutoMigrate` in `main` for prod apps | goose/golang-migrate | warn |
| gorm-04 | implemented | `Debug()` enabled in non-dev | remove debug | warn |
| gorm-05 | implemented | package-level `var DB *gorm.DB` | inject via constructor | warn |

## sqlx

| id | status | when | fix |
|----|--------|------|-----|
| sqlx-01 | implemented | `Get`/`Select` without context | pass `ctx` |

## sqlc / ent

| id | status | when | fix |
|----|--------|------|-----|
| sqlc-01 | implemented | hand-edited `*.sql.go` | regenerate from sql |
| ent-01 | implemented | edge traversal in loop without eager load | batch query or eager edge |

## overlap checks

| id | status | overlap | notes |
|----|--------|---------|-------|
| sec-db01 | overlap → sql-02 | string-built sql | security gate |
| sec-db03 | overlap → pgx-03 | `sslmode=disable` on remote | security gate |
| sec-db05 | implemented | dynamic table/column from user | sql gate |

## security

see [security-owasp.md](security-owasp.md) — sec-db01…sec-db05
