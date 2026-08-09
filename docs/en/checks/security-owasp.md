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
| sec-13 | log of password/token/otp | redact | fail |
| sec-14 | cors `*` + credentials | explicit origins | fail |
| sec-15 | `net/http/pprof` import without build tag | dev-only | warn |
| sec-16 | `govulncheck` findings (flag `--with-security`) | upgrade dep | fail |

## rest-specific

| id | when | fix |
|----|------|-----|
| sec-r01 | handler uses path `id` without authz check | verify actor owns resource |
| sec-r03 | `json.Unmarshal` into db model from request | dto + explicit fields |
| sec-r04 | login route without rate limiter | middleware limit |
| sec-r05 | `/admin` without role middleware | rbac check |
| sec-r10 | webhook handler without hmac verify | signature + replay id |

## graphql-specific

| id | when | fix |
|----|------|-----|
| sec-g01 | introspection on in prod build | gate by env |
| sec-g02 | no max query depth | server limit |
| sec-g03 | no complexity/cost analysis | extension |
| sec-g04 | `node(id:)` without ownership | authz in resolver |
| sec-g05 | unlimited alias batching | cost limit / persisted queries |
| sec-g06 | cookie auth + get queries | csrf token or post-only |

## database

| id | when | fix |
|----|------|-----|
| sec-db01 | string-built sql | parameters |
| sec-db02 | dsn password in repo | env |
| sec-db03 | `sslmode=disable` remote | tls |
| sec-db05 | dynamic table/column from user | allowlist |

## optional subprocess

`refgrade scan --with-security` runs `govulncheck` when installed; does not replace manual review

## honest gaps (printed in report footer)

- live idor / bola reproduction
- dast / fuzzing
- infrastructure iam, k8s rbac
- explain analyze for sql perf
