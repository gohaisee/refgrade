# security (OWASP API 2023 + static)

static heuristics only — live IDOR/BOLA reproduction, DAST, and K8s IAM are **runtime gaps** (printed in report footer)

## owasp api top 10 mapping

| owasp | id range | focus in go |
|-------|----------|-------------|
| API1 broken object level auth | sec-r01, sec-g04 | user id from client without ownership check |
| API2 broken authentication | sec-04, sec-05 | jwt alg, empty secret |
| API3 broken object property auth | sec-r03 | bind whole struct from json |
| API4 unrestricted resource consumption | sec-r04, sec-g02, sec-g03 | rate limit, body limit, gql depth |
| API5 broken function level auth | sec-r05 | admin routes without role |
| API6 unrestricted sensitive flows | sec-r06 | otp/payment without step-up (heuristic) |
| API7 ssrf | sec-09 | http get to user url |
| API8 security misconfiguration | sec-08, sec-14, sec-15 | debug, cors, pprof |
| API9 improper inventory | sec-r09 | shadow handlers (heuristic) |
| API10 unsafe consumption of apis | sec-r10 | webhook without signature |

## static checks (all modules)

| id | status | when | fix | severity |
|----|--------|------|-----|----------|
| sec-01 | implemented | `crypto/md5` / `sha1` for passwords | bcrypt/argon2/scrypt | fail |
| sec-02 | implemented | `tls.Config{InsecureSkipVerify: true}` | proper ca | fail |
| sec-03 | implemented | `math/rand` for tokens/session ids | `crypto/rand` | fail |
| sec-04 | implemented | jwt parse without algorithm pin | validate method in callback | fail |
| sec-05 | implemented | `jwt.SigningMethodNone` | reject | fail |
| sec-07 | implemented | `os/exec` with user-controlled args | allowlist commands | fail |
| sec-08 | implemented | `template.HTML(userInput)` | auto-escape; sanitize | fail |
| sec-09 | implemented | http client to url from user input | ssrf guard; block private ranges | warn |
| sec-10 | implemented | `.env` tracked in git | gitignore; rotate secrets | fail |
| sec-15 | implemented | `net/http/pprof` import without build tag | dev-only | warn |
| sec-16 | implemented | govulncheck not run (default) or CVE found (`--with-security`) | run `--with-security`; upgrade deps | warn / fail |

## overlap checks

| id | status | overlap | notes |
|----|--------|---------|-------|
| sec-13 | overlap → obs-02 | sensitive fields in log arguments | rest gate: n/a |
| sec-14 | overlap → rest-05 | cors wildcard + credentials | rest gate |
| sec-g01 | overlap → gql-05 | introspection without env gate | graphql gate |
| sec-g02 | overlap → gql-04 | no query depth limit | graphql gate |
| sec-g03 | overlap → gql-04 | no complexity/cost limit | graphql gate |
| sec-db01 | overlap → sql-02 | string-built sql | sql gate |
| sec-db02 | overlap → cfg-03 | hardcoded secrets in source | universal |
| sec-db03 | overlap → pgx-03 | `sslmode=disable` on remote | pgx gate |

## domain security (gated)

| id | status | gate | when | fix | severity |
|----|--------|------|------|-----|----------|
| sec-m01 | implemented | mongo | mongo uri password in source | secret manager | fail |
| sec-m02 | implemented | mongo | remote mongo without tls | mongodb+srv or tls=true | fail |
| sec-rd01 | implemented | redis | redis password in source | env only | fail |
| sec-rd02 | implemented | redis | redis url to remote without tls | rediss:// or tls options | fail |
| sec-mq01 | implemented | mq | broker url credentials in source | env / vault | fail |
| sec-mq02 | implemented | mq | plaintext amqp/nats to remote | amqps or tls dial | fail |

## rest-specific

| id | status | when | fix |
|----|--------|------|-----|
| sec-r01 | implemented | handler uses path `id` without authz check | verify actor owns resource |
| sec-r03 | implemented | `json.Unmarshal` into db model from request | dto + explicit fields |
| sec-r04 | implemented | login route without rate limiter | middleware limit |
| sec-r05 | implemented | `/admin` without role middleware | rbac check |
| sec-r06 | implemented | payment or otp flow without step-up auth | mfa or step-up before payout |
| sec-r09 | implemented | `http.Handle` registered outside cmd setup | central router only |
| sec-r10 | implemented | webhook handler without hmac verify | signature + replay id |

## graphql-specific

| id | status | when | fix |
|----|--------|------|-----|
| sec-g04 | implemented | `node(id:)` without ownership | authz in resolver |
| sec-g05 | implemented | unlimited alias batching | cost limit / persisted queries |
| sec-g06 | implemented | cookie session auth with get graphql query | csrf token or post-only |

## database

| id | status | when | fix |
|----|--------|------|-----|
| sec-db05 | implemented | dynamic table/column from user | allowlist |

## optional subprocess

`refgrade scan --with-security` runs `govulncheck` for **sec-16** when installed; without the flag **sec-16** warns that dependency CVE scan was skipped

## honest gaps (printed in report footer)

| topic | status |
|-------|--------|
| live IDOR / BOLA reproduction | runtime gap |
| DAST / fuzzing | runtime gap |
| infrastructure IAM / K8s RBAC | runtime gap |
| explain analyze for SQL perf | runtime gap |

i18n keys: `report.gap.idor`, `report.gap.dast`, `report.gap.k8s` in `locales/*.yaml`
