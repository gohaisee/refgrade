# security (OWASP API 2023 + static)

static heuristics only — runtime pentest and idor reproduction are **out of scope** for v1 (report as gap)

## owasp api top 10 mapping

| owasp | id range | focus in go |
|-------|----------|-------------|
| API1 broken object level auth | sec-r01, sec-g04 | user id from client without ownership check |
| API2 broken authentication | sec-04, sec-05 | jwt alg, empty secret |
| API3 broken object property auth | sec-r03 | bind whole struct from json |
| API4 unrestricted resource consumption | sec-r04, sec-g02, sec-g03 | rate limit, body limit, gql depth |
| API5 broken function level auth | sec-r05, sec-g07 | admin routes without role |
| API6 unrestricted sensitive flows | sec-r06 | otp/payment without step-up (heuristic) |
| API7 ssrf | sec-09 | http get to user url |
| API8 security misconfiguration | sec-08, sec-14, sec-15 | debug, cors, pprof |
| API9 improper inventory | sec-r09 | shadow handlers (heuristic) |
| API10 unsafe consumption of apis | sec-r10 | webhook without signature |

## static checks (all modules)

| id | when | fix | severity |
|----|------|-----|----------|
| sec-01 | `crypto/md5` / `sha1` for passwords | bcrypt/argon2/scrypt | fail |
| sec-02 | `tls.Config{InsecureSkipVerify: true}` | proper ca | fail |
| sec-03 | `math/rand` for tokens/session ids | `crypto/rand` | fail |
| sec-04 | jwt parse without algorithm pin | validate method in callback | fail |
| sec-05 | `jwt.SigningMethodNone` | reject | fail |
| sec-07 | `os/exec` with user-controlled args | allowlist commands | fail |
| sec-08 | `template.HTML(userInput)` | auto-escape; sanitize | fail |
| sec-09 | http client to url from user input | ssrf guard; block private ranges | warn |
| sec-10 | `.env` tracked in git | gitignore; rotate secrets | fail |
| sec-15 | `net/http/pprof` import without build tag | dev-only | warn |

## overlap and deferred ids (v1.0.0)

catalog rows below are **not** separate registry checks — use the implementing id or treat as deferred.

| id | status | implemented as | notes |
|----|--------|----------------|-------|
| sec-13 | covered-by | [obs-02](universal.md#observability) | sensitive fields in log arguments |
| sec-14 | covered-by | [rest-05](rest-gin-echo-chi.md) | cors wildcard + credentials (rest gate) |
| sec-g01 | covered-by | [gql-05](graphql.md) | introspection without env gate |
| sec-g02 | covered-by | [gql-04](graphql.md) | no query depth limit |
| sec-g03 | covered-by | [gql-04](graphql.md) | no complexity/cost limit |
| sec-db01 | covered-by | [sql-02](sql.md) | string-built sql |
| sec-db02 | covered-by | [cfg-03](universal.md#config) | hardcoded secrets in source |
| sec-db03 | covered-by | [pgx-03](sql.md) | `sslmode=disable` on remote |
| sec-16 | subprocess | `govulncheck` via `--with-security` | not a registry row; optional tool |
| sec-m01 | deferred | [cfg-03](universal.md#config) overlap | mongo uri password in repo |
| sec-m02 | deferred | — | remote mongo tls — manual review |
| sec-rd01 | deferred | [cfg-03](universal.md#config) overlap | redis password in source |
| sec-rd02 | deferred | — | redis tls on public network |
| sec-mq01 | deferred | [cfg-03](universal.md#config) overlap | broker url credentials in repo |
| sec-mq02 | deferred | — | plaintext amqp/nats to public internet |

## rest-specific

| id | when | fix |
|----|------|-----|
| sec-r01 | handler uses path `id` without authz check | verify actor owns resource |
| sec-r03 | `json.Unmarshal` into db model from request | dto + explicit fields |
| sec-r04 | login route without rate limiter | middleware limit |
| sec-r05 | `/admin` without role middleware | rbac check |
| sec-r10 | webhook handler without hmac verify | signature + replay id |

## graphql-specific (implemented)

| id | when | fix |
|----|------|-----|
| sec-g04 | `node(id:)` without ownership | authz in resolver |
| sec-g05 | unlimited alias batching | cost limit / persisted queries |
| sec-g06 | cookie auth + get queries | csrf token or post-only |

`sec-g01`…`sec-g03` — covered-by [gql-04](graphql.md) / [gql-05](graphql.md); see [overlap table](#overlap-and-deferred-ids-v100)

## database (implemented)

| id | when | fix |
|----|------|-----|
| sec-db05 | dynamic table/column from user | allowlist |

`sec-db01`…`sec-db03` — covered-by [sql-02](sql.md), [cfg-03](universal.md#config), [pgx-03](sql.md); see [overlap table](#overlap-and-deferred-ids-v100)

## optional subprocess

`refgrade scan --with-security` runs `govulncheck` when installed; does not replace manual review

## honest gaps (printed in report footer)

- live idor / bola reproduction
- dast / fuzzing
- infrastructure iam, k8s rbac
- explain analyze for sql perf
