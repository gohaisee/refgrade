# how it works

refgrade is a cli, not a linter replacement

## pipeline

1. load module (`go list -json ./...`)
2. **detect** stack — gin, gqlgen, pgx, mongo, …
3. run **universal** checks (config, layers, errors, tests, dead code)
4. run **stack** checks only when the library is present
5. mark missing stack as **n/a**, not fail
6. print report: text, markdown, or json

## commands

| command | role |
|---------|------|
| `scan` | main report |
| `detect` | print detected stack only |
| `init` | write `.refgrade.yaml` template |
| `explain <id>` | one check in human words |

## stack checks (phase 2)

| domain | gates | ids |
|--------|-------|-----|
| rest | gin, echo, chi, net/http | rest-01..10 |
| sql | pgx, gorm, sqlx, sqlc, ent, database/sql | sql-01..07, pgx-01..03, gorm-01..05, sqlx-01, sqlc-01, ent-01 |
| graphql | gqlgen, graphql-go, graphql | gql-01..09, gqlgen-01..02, ggl-01..02 |
| grpc | grpc, connect, grpc-gateway | grpc-01..06, conn-01..02, gw-01..02 |

## exit codes

- `0` — no fail-level findings
- `1` — at least one fail
- `2` — scan error (broken module, bad path)
