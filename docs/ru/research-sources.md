# источники ресерча

ссылки, на которых собран каталог проверок — читать не обязательно, чтобы пользоваться refgrade

## go http / rest

- [JetBrains — popular go web frameworks (2026)](https://blog.jetbrains.com/go/2026/04/28/popular-golang-web-frameworks/) — доля gin/echo/chi
- [Go 1.22 net/http routing](https://go.dev/doc/go1.22) — method routing в stdlib

## graphql

- [gqlgen documentation](https://gqlgen.com/)
- [graph-gophers/graphql-go](https://github.com/graph-gophers/graphql-go)
- [OWASP WSTG — testing GraphQL](https://owasp.org/www-project-web-security-testing-guide/stable/4-Web_Application_Security_Testing/12-API_Testing/01-Testing_GraphQL)
- [Postman — OWASP API 2023 and GraphQL](https://blog.postman.com/owasp-api-security-top-10-2023-and-graphql/)

## sql / orm

- [pgx documentation](https://pkg.go.dev/github.com/jackc/pgx/v5)
- [go.dev — deadcode tool blog](https://go.dev/blog/deadcode)
- сравнения gorm vs sqlx vs pgx vs sqlc (dev.to, glukhov.org, 2025–2026)

## mongo / redis

- паттерны пула mongo-driver (официальные docs + community)
- [redis cache-aside in go](https://redis.io/docs/latest/develop/use-cases/cache-aside/go/)
- [go-patterns — singleflight](https://go-patterns.dev/sync/singleflight)

## messaging

- официальные docs kafka-go, amqp091-go, nats.go
- at-least-once delivery и идемпотентные consumers

## security

- [OWASP API Security Top 10 2023](https://owasp.org/API-Security/editions/2023/en/0x11-t10/)
- [Anthropic-Cybersecurity-Skills](https://github.com/mukul975/Anthropic-Cybersecurity-Skills) — паттерны api security (apache 2.0, community)

## dead code / static analysis

- `golang.org/x/tools/cmd/deadcode`
- [staticcheck U1000 / build tags](https://staticcheck.dev/docs/running-staticcheck/cli/build-tags/)

## disclaimer

где помечено — это эвристики; `warn` значит «подтверди человеком». refgrade не обещает полный security audit
