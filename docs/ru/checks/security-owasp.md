# безопасность (OWASP API 2023 + статика)

только статические эвристики — runtime pentest и воспроизведение idor **вне scope** v1 (в отчёте как gap)

## маппинг owasp api top 10

| owasp | id | фокус в go |
|-------|-----|------------|
| API1 broken object level auth | sec-r01, sec-g04 | id от клиента без проверки владельца |
| API2 broken authentication | sec-04, sec-05 | jwt alg, пустой secret |
| API3 broken object property auth | sec-r03 | bind всего struct из json |
| API4 unrestricted resource consumption | sec-r04, sec-g02, sec-g03 | rate limit, body limit, глубина gql |
| API5 broken function level auth | sec-r05, sec-g07 | admin-ручки без роли |
| API6 unrestricted sensitive flows | sec-r06 | otp/платёж без step-up (эвристика) |
| API7 ssrf | sec-09 | http get на url от пользователя |
| API8 security misconfiguration | sec-08, sec-14, sec-15 | debug, cors, pprof |
| API9 improper inventory | sec-r09 | теневые handlers (эвристика) |
| API10 unsafe consumption of apis | sec-r10 | webhook без подписи |

## статические проверки (все модули)

| id | when | fix | severity |
|----|------|-----|----------|
| sec-01 | `crypto/md5` / `sha1` для паролей | bcrypt/argon2/scrypt | fail |
| sec-02 | `tls.Config{InsecureSkipVerify: true}` | нормальный ca | fail |
| sec-03 | `math/rand` для токенов/session id | `crypto/rand` | fail |
| sec-04 | jwt parse без фиксации algorithm | проверять method в callback | fail |
| sec-05 | `jwt.SigningMethodNone` | отклонять | fail |
| sec-07 | `os/exec` с аргументами от пользователя | allowlist команд | fail |
| sec-08 | `template.HTML(userInput)` | auto-escape; sanitize | fail |
| sec-09 | http client на url от пользователя | ssrf guard; блок private ranges | warn |
| sec-10 | `.env` в git | gitignore; ротация секретов | fail |
| sec-15 | `net/http/pprof` без build tag | только dev | warn |

## overlap и deferred id (v1.0.0)

строки каталога ниже **не** отдельные проверки в registry — смотри implementing id или пометку deferred.

| id | статус | реализовано как | примечание |
|----|--------|-----------------|------------|
| sec-13 | covered-by | [obs-02](universal.md#observability) | чувствительные поля в аргументах лога |
| sec-14 | covered-by | [rest-05](rest-gin-echo-chi.md) | cors wildcard + credentials (rest gate) |
| sec-g01 | covered-by | [gql-05](graphql.md) | introspection без env gate |
| sec-g02 | covered-by | [gql-04](graphql.md) | нет лимита глубины запроса |
| sec-g03 | covered-by | [gql-04](graphql.md) | нет complexity/cost limit |
| sec-db01 | covered-by | [sql-02](sql.md) | sql из конкатенации строк |
| sec-db02 | covered-by | [cfg-03](universal.md#config) | захардкоженные секреты в исходниках |
| sec-db03 | covered-by | [pgx-03](sql.md) | `sslmode=disable` на remote |
| sec-16 | subprocess | `govulncheck` через `--with-security` | не строка registry; опциональный tool |
| sec-m01 | deferred | overlap [cfg-03](universal.md#config) | пароль mongo uri в репо |
| sec-m02 | deferred | — | tls удалённого mongo — ручной review |
| sec-rd01 | deferred | overlap [cfg-03](universal.md#config) | пароль redis в исходниках |
| sec-rd02 | deferred | — | redis tls в публичной сети |
| sec-mq01 | deferred | overlap [cfg-03](universal.md#config) | credentials broker url в репо |
| sec-mq02 | deferred | — | plaintext amqp/nats в публичный интернет |

## rest

| id | when | fix |
|----|------|-----|
| sec-r01 | handler берёт `id` из path без authz | проверить, что actor владеет ресурсом |
| sec-r03 | `json.Unmarshal` в db model из запроса | dto + явные поля |
| sec-r04 | login без rate limiter | middleware limit |
| sec-r05 | `/admin` без role middleware | rbac check |
| sec-r10 | webhook без hmac verify | signature + replay id |

## graphql (реализовано)

| id | when | fix |
|----|------|-----|
| sec-g04 | `node(id:)` без ownership | authz в resolver |
| sec-g05 | неограниченный alias batching | cost limit / persisted queries |
| sec-g06 | cookie auth + get queries | csrf token или post-only |

`sec-g01`…`sec-g03` — covered-by [gql-04](graphql.md) / [gql-05](graphql.md); см. [таблицу overlap](#overlap-и-deferred-id-v100)

## database (реализовано)

| id | when | fix |
|----|------|-----|
| sec-db05 | table/column из user input | allowlist |

`sec-db01`…`sec-db03` — covered-by [sql-02](sql.md), [cfg-03](universal.md#config), [pgx-03](sql.md); см. [таблицу overlap](#overlap-и-deferred-id-v100)

## опциональный subprocess

`refgrade scan --with-security` запускает `govulncheck`, если установлен; не заменяет ручной review

## честные пробелы (в footer отчёта)

- live idor / bola
- dast / fuzzing
- infrastructure iam, k8s rbac
- explain analyze для sql perf
