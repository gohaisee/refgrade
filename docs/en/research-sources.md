# research sources

links used to build the check catalog — not required reading to use refgrade

## go http / rest frameworks

- [JetBrains — popular go web frameworks (2026)](https://blog.jetbrains.com/go/2026/04/28/popular-golang-web-frameworks/) — gin/echo/chi adoption
- [Go 1.22 net/http routing](https://go.dev/doc/go1.22) — stdlib method routing

## graphql

- [gqlgen documentation](https://gqlgen.com/)
- [graph-gophers/graphql-go](https://github.com/graph-gophers/graphql-go)
- [OWASP WSTG — testing GraphQL](https://owasp.org/www-project-web-security-testing-guide/stable/4-Web_Application_Security_Testing/12-API_Testing/01-Testing_GraphQL)
- [Postman — OWASP API 2023 and GraphQL](https://blog.postman.com/owasp-api-security-top-10-2023-and-graphql/)

## sql / orm

- [pgx documentation](https://pkg.go.dev/github.com/jackc/pgx/v5)
- [go.dev — deadcode tool blog](https://go.dev/blog/deadcode)
- community comparisons: gorm vs sqlx vs pgx vs sqlc (dev.to, glukhov.org posts, 2025–2026)

## mongo / redis

- mongo-driver connection pooling patterns (official docs + community guides)
- [redis cache-aside in go](https://redis.io/docs/latest/develop/use-cases/cache-aside/go/)
- [go-patterns — singleflight](https://go-patterns.dev/sync/singleflight)

## messaging

- kafka-go, amqp091-go, nats.go official docs
- at-least-once delivery and idempotent consumer patterns

## security

- [OWASP API Security Top 10 2023](https://owasp.org/API-Security/editions/2023/en/0x11-t10/)
- [Anthropic-Cybersecurity-Skills](https://github.com/mukul975/Anthropic-Cybersecurity-Skills) — api security domain patterns (apache 2.0, community project)

## dead code / static analysis

- `golang.org/x/tools/cmd/deadcode`
- [staticcheck U1000 / build tags](https://staticcheck.dev/docs/running-staticcheck/cli/build-tags/)

## disclaimer

checks are **heuristics** where noted; severity `warn` means human should confirm. refgrade does not claim exhaustive security audit
