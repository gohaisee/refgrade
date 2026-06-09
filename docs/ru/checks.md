# каталог проверок

правила по доменам — universal + стековая фаза 2–3 реализованы в `internal/check/`

| файл | когда запускается |
|------|-------------------|
| [checks/universal.md](checks/universal.md) | любой go-модуль |
| [checks/rest-gin-echo-chi.md](checks/rest-gin-echo-chi.md) | `net/http`, gin, echo, chi |
| [checks/graphql.md](checks/graphql.md) | gqlgen, graphql-go, code-first graphql |
| [checks/grpc.md](checks/grpc.md) | grpc-go, connect, grpc-gateway |
| [checks/sql.md](checks/sql.md) | pgx, gorm, sqlx, sqlc, ent |
| [checks/mongodb.md](checks/mongodb.md) | mongo-driver |
| [checks/redis.md](checks/redis.md) | go-redis |
| [checks/messaging.md](checks/messaging.md) | kafka, rabbitmq, nats |
| [checks/security-owasp.md](checks/security-owasp.md) | статическая безопасность |
| [checks/dead-code.md](checks/dead-code.md) | модули с `cmd/` или `internal/` |

статусы: `ok` · `warn` · `fail` · `n/a`

каждая строка: **id** · **when** · **why** · **fix** · **severity**

[research-sources.md](research-sources.md)
