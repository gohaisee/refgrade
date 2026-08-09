# changelog

all notable changes to refgrade are documented here.

format follows [keep a changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

## [2.0.0] - 2026-08-09

full catalog release — **152** registry checks (up from **133**), security overlap aliases, stack-gated domain rules, and deadcode CLI ([#6](https://github.com/gohaisee/refgrade/pull/6)).

### added

**catalog expansion (133 → 152)**

- security overlap aliases: `sec-13`, `sec-14`, `sec-g01`…`sec-g03`, `sec-db01`…`sec-db03` — same logic as base checks, separate catalog ids
- graphql security extensions: `sec-g07`…`sec-g10`
- domain security (mongo / redis / messaging): `sec-m01`…`sec-m02`, `sec-rd01`…`sec-rd02`, `sec-mq01`…`sec-mq02`
- `sec-16` (`govulncheck`) registered in the catalog; warns when `--with-security` is off, runs the tool when on
- integration fixtures for mongo, redis, and mq security negatives under `testdata/fixtures/`

**cli**

- `refgrade deadcode` — focused `dead-01`…`dead-08` run with `--include-tests` and build tags
- `scan --all-modules` — scan every nested `go.mod` under the path; merge JSON/SARIF output across modules

**quality and docs**

- registry guard: `TestRegistry_hasAtLeast150Checks`
- dogfood scan guide (`docs/en/dogfood.md`, `docs/ru/dogfood.md`)
- catalog row status on every check id: `implemented`, `overlap → X`, or `runtime gap`
- en/ru check docs refreshed (security-owasp and stack pages)

### changed

- overlap aliases skip when the delegated base check runs; alias still runs when the base is disabled in config
- report footer i18n: split keys for runtime gaps (IDOR/BOLA, DAST, K8s IAM)
- deadcode scan path honors `IncludeTests` in project load and engine
- broader security and domain check test coverage

### fixed

- review fixes from PR #6: overlap config delegation, multi-module report merge, deadcode CLI and engine edge cases

### notes

- `scan` exits **1** only on `fail` findings; `warn` prints but exit **0**
- overlap and runtime-gap semantics — [docs/en/checks/security-owasp.md](docs/en/checks/security-owasp.md)

## [1.0.0] - 2026-08-09

first stable release — full stack refactor readiness scanner for go services.

### added

**phase 1 — core cli**

- commands: `scan`, `detect`, `explain`, `init`
- universal checks: config, layout, layering, errors, tests, observability, concurrency
- i18n: `en` and `ru` via embedded `locales/*.yaml`
- config: `.refgrade.yaml` (lang, profile, kind, exclude, layers, checks)
- stack detection and gating — checks emit `n/a` when stack not present

**phase 2 — stack catalog (76 checks)**

- rest: gin, echo, chi, net/http (`rest-01`…`rest-10`)
- sql: pgx, gorm, sqlx, sqlc, ent, database/sql (`sql-01`…`sql-07`, `pgx-01`…`pgx-03`, `gorm-01`…`gorm-05`, `sqlx-01`, `sqlc-01`, `ent-01`)
- graphql: gqlgen, graphql-go, code-first (`gql-01`…`gql-09`, `gqlgen-01`…`gqlgen-02`, `ggl-01`…`ggl-02`)
- grpc: grpc-go, connect, grpc-gateway (`grpc-01`…`grpc-06`, `conn-01`…`conn-02`, `gw-01`…`gw-02`)

**phase 3 — data, messaging, security, dead code (57 checks → 133 total)**

- mongodb: `mongo-01`…`mongo-08`
- redis: `redis-01`…`redis-09`
- messaging: kafka, rabbitmq, nats (`mq-01`…`mq-05`, `kafka-01`…`kafka-03`, `rmq-01`…`rmq-03`, `nats-01`…`nats-02`)
- security static: owasp-oriented heuristics (`sec-01`…`sec-10`, `sec-15`, `sec-r*`, `sec-g04`…`sec-g06`, `sec-db05`)
- dead code extended: `dead-01`…`dead-08` (ast + subprocess)

**infrastructure**

- yaml config loading and profile overrides
- sarif 2.1.0 output (`--format sarif`)
- ci workflow and composite github action (`action.yml`)
- subprocess tools: `deadcode`, `staticcheck` (U1000), `go mod tidy`, `govulncheck` (`--with-security`)
- stdlib-only tool code — no third-party deps in the scanner itself

### notes

- `scan` exits **1** only on `fail` findings; `warn` prints but exit **0**
- `sec-16` (`govulncheck`) runs only with `--with-security`; not a registry row
- catalog ids marked deferred or covered-by in [docs/en/checks/security-owasp.md](docs/en/checks/security-owasp.md)

[Unreleased]: https://github.com/gohaisee/refgrade/compare/v2.0.0...HEAD
[2.0.0]: https://github.com/gohaisee/refgrade/releases/tag/v2.0.0
[1.0.0]: https://github.com/gohaisee/refgrade/releases/tag/v1.0.0
