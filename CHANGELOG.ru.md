# changelog

все значимые изменения refgrade документируются здесь.

формат следует [keep a changelog](https://keepachangelog.com/ru/1.1.0/).

## [Unreleased]

### добавлено

- ci coverage gate для `internal/check` (минимум 75%)
- dogfood-гайд (`docs/en/dogfood.md`, `docs/ru/dogfood.md`)
- колонка status у каждого check id: `implemented`, `overlap → X` или `runtime gap`

### изменено

- i18n footer отчёта: отдельные ключи для runtime gaps IDOR/BOLA, DAST и K8s IAM (`report.gap.idor`, `report.gap.dast`, `report.gap.k8s`)

## [1.0.0] - 2026-08-09

первый стабильный релиз — сканер готовности go-сервисов к рефакторингу по полному стековому каталогу.

### добавлено

**фаза 1 — ядро cli**

- команды: `scan`, `detect`, `explain`, `init`
- universal-проверки: config, layout, layering, errors, tests, observability, concurrency
- i18n: `en` и `ru` через встроенные `locales/*.yaml`
- конфиг: `.refgrade.yaml` (lang, profile, kind, exclude, layers, checks)
- детект стека и gating — проверки дают `n/a`, если стек не обнаружен

**фаза 2 — стековый каталог (76 проверок)**

- rest: gin, echo, chi, net/http (`rest-01`…`rest-10`)
- sql: pgx, gorm, sqlx, sqlc, ent, database/sql (`sql-01`…`sql-07`, `pgx-01`…`pgx-03`, `gorm-01`…`gorm-05`, `sqlx-01`, `sqlc-01`, `ent-01`)
- graphql: gqlgen, graphql-go, code-first (`gql-01`…`gql-09`, `gqlgen-01`…`gqlgen-02`, `ggl-01`…`ggl-02`)
- grpc: grpc-go, connect, grpc-gateway (`grpc-01`…`grpc-06`, `conn-01`…`conn-02`, `gw-01`…`gw-02`)

**фаза 3 — данные, messaging, безопасность, dead code (57 проверок → 133 всего)**

- mongodb: `mongo-01`…`mongo-08`
- redis: `redis-01`…`redis-09`
- messaging: kafka, rabbitmq, nats (`mq-01`…`mq-05`, `kafka-01`…`kafka-03`, `rmq-01`…`rmq-03`, `nats-01`…`nats-02`)
- security static: эвристики в духе owasp (`sec-01`…`sec-10`, `sec-15`, `sec-r*`, `sec-g04`…`sec-g06`, `sec-db05`)
- dead code расширен: `dead-01`…`dead-08` (ast + subprocess)

**инфраструктура**

- загрузка yaml-конфига и profile overrides
- вывод sarif 2.1.0 (`--format sarif`)
- ci workflow и composite github action (`action.yml`)
- subprocess-инструменты: `deadcode`, `staticcheck` (U1000), `go mod tidy`, `govulncheck` (`--with-security`)
- только stdlib в коде сканера — без сторонних зависимостей

### примечания

- `scan` выходит с **1** только при `fail`; `warn` печатается, но exit **0**
- `sec-16` (`govulncheck`) только с `--with-security`; не строка в registry
- id каталога с пометками deferred / covered-by — [docs/ru/checks/security-owasp.md](docs/ru/checks/security-owasp.md)

[1.0.0]: https://github.com/gohaisee/refgrade/releases/tag/v1.0.0
