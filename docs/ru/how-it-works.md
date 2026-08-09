# как это работает

refgrade — это cli, не замена golangci-lint

## pipeline

1. загружаем модуль (`go list -json ./...`)
2. **detect** стека — gin, gqlgen, pgx, mongo, …
3. гоняем **универсальные** проверки (config, слои, ошибки, тесты, мёртвый код)
4. **стековые** проверки — только если библиотека найдена
5. если стека нет — ставим **n/a**, не fail
6. печатаем отчёт: text, markdown или json

## команды

| команда | зачем |
|---------|--------|
| `scan` | основной отчёт |
| `detect` | только найденный стек |
| `init` | шаблон `.refgrade.yaml` |
| `explain <id>` | одна проверка человеческим языком |

## стековые проверки (фаза 2–3)

| домен | gates | ids |
|-------|-------|-----|
| rest | gin, echo, chi, net/http | rest-01..10 |
| sql | pgx, gorm, sqlx, sqlc, ent, database/sql | sql-01..07, pgx-01..03, gorm-01..05, sqlx-01, sqlc-01, ent-01 |
| graphql | gqlgen, graphql-go, graphql | gql-01..09, gqlgen-01..02, ggl-01..02 |
| grpc | grpc, connect, grpc-gateway | grpc-01..06, conn-01..02, gw-01..02 |
| mongodb | mongo-driver | mongo-01..08 |
| redis | go-redis | redis-01..09 |
| messaging | kafka, rabbitmq, nats | mq-01..05, kafka-01..03, rmq-01..03, nats-01..02 |
| security | static (универсальные + gated) | sec-01..10, sec-15, sec-r*, sec-g04..06, sec-db05 |
| dead code | универсальные + subprocess | dead-01..08, sec-16 через `--with-security` |

**133** проверки в `internal/check/registry.go` (фаза 3 добавляет **57** к 76 фазы 2).

## коды выхода

- `0` — нет findings уровня fail
- `1` — есть хотя бы один fail
- `2` — ошибка скана (битый модуль, плохой путь)
