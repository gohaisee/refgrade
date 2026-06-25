# checks catalog

rules by domain — universal + stack phase 2–3 implemented in `internal/check/`

| file | when it runs |
|------|----------------|
| [checks/universal.md](checks/universal.md) | every go module |
| [checks/rest-gin-echo-chi.md](checks/rest-gin-echo-chi.md) | `net/http`, gin, echo, chi |
| [checks/graphql.md](checks/graphql.md) | gqlgen, graphql-go, code-first graphql |
| [checks/grpc.md](checks/grpc.md) | grpc-go, connect, grpc-gateway |
| [checks/sql.md](checks/sql.md) | pgx, gorm, sqlx, sqlc, ent |
| [checks/mongodb.md](checks/mongodb.md) | mongo-driver |
| [checks/redis.md](checks/redis.md) | go-redis |
| [checks/messaging.md](checks/messaging.md) | kafka, rabbitmq, nats |
| [checks/security-owasp.md](checks/security-owasp.md) | static security |
| [checks/dead-code.md](checks/dead-code.md) | modules with `cmd/` or `internal/` |

scan status: `ok` · `warn` · `fail` · `n/a`

catalog row status: `implemented` · `overlap → X` · `runtime gap`

each row: **id** · **status** · **when** · **why** · **fix** · **severity**

## implementation status (v2)

| phase | domain | registered checks |
|-------|--------|-------------------|
| 1 | universal (cfg, lay, lyr, err, tst, obs, con) | 23 |
| 2 | rest, sql, graphql, grpc | 51 |
| 3 | dead code, mongo, redis, messaging, security static | 59 |
| | **total** | **133** |

overlap security ids: [checks/security-owasp.md](checks/security-owasp.md#overlap-checks)

runtime gaps (footer): [checks/security-owasp.md](checks/security-owasp.md#honest-gaps-printed-in-report-footer)

dogfood guide: [dogfood.md](dogfood.md)

[research-sources.md](research-sources.md)
