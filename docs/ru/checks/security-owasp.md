# безопасность (OWASP API 2023 + статика)

только статические эвристики — live IDOR/BOLA, DAST и K8s IAM — **runtime gap** (в footer отчёта)

## маппинг owasp api top 10

| owasp | id | фокус в go |
|-------|-----|------------|
| API1 broken object level auth | sec-r01, sec-g04 | id от клиента без проверки владельца |
| API2 broken authentication | sec-04, sec-05 | jwt alg, пустой secret |
| API3 broken object property auth | sec-r03 | bind всего struct из json |
| API4 unrestricted resource consumption | sec-r04, sec-g02, sec-g03 | rate limit, body limit, глубина gql |
| API5 broken function level auth | sec-r05 | admin-ручки без роли |
| API6 unrestricted sensitive flows | sec-r06 | otp/платёж без step-up (эвристика) |
| API7 ssrf | sec-09 | http get на url от пользователя |
| API8 security misconfiguration | sec-08, sec-14, sec-15 | debug, cors, pprof |
| API9 improper inventory | sec-r09 | теневые handlers (эвристика) |
| API10 unsafe consumption of apis | sec-r10 | webhook без подписи |

## статические проверки (все модули)

| id | status | when | fix | severity |
|----|--------|------|-----|----------|
| sec-01 | implemented | `crypto/md5` / `sha1` для паролей | bcrypt/argon2/scrypt | fail |
| sec-02 | implemented | `tls.Config{InsecureSkipVerify: true}` | нормальный ca | fail |
| sec-03 | implemented | `math/rand` для токенов/session id | `crypto/rand` | fail |
| sec-04 | implemented | jwt parse без фиксации algorithm | проверять method в callback | fail |
| sec-05 | implemented | `jwt.SigningMethodNone` | отклонять | fail |
| sec-07 | implemented | `os/exec` с аргументами от пользователя | allowlist команд | fail |
| sec-08 | implemented | `template.HTML(userInput)` | auto-escape; sanitize | fail |
| sec-09 | implemented | http client на url от пользователя | ssrf guard; блок private ranges | warn |
| sec-10 | implemented | `.env` в git | gitignore; ротация секретов | fail |
| sec-15 | implemented | `net/http/pprof` без build tag | только dev | warn |
| sec-16 | implemented | govulncheck не запущен (по умолчанию) или найден CVE (`--with-security`) | `--with-security`; обновить deps | warn / fail |

## overlap checks

| id | status | overlap | примечание |
|----|--------|---------|------------|
| sec-13 | overlap → obs-02 | чувствительные поля в аргументах лога | universal |
| sec-14 | overlap → rest-05 | cors wildcard + credentials | rest gate |
| sec-g01 | overlap → gql-05 | introspection без env gate | graphql gate |
| sec-g02 | overlap → gql-04 | нет лимита глубины запроса | graphql gate |
| sec-g03 | overlap → gql-04 | нет complexity/cost limit | graphql gate |
| sec-db01 | overlap → sql-02 | sql из конкатенации строк | sql gate |
| sec-db02 | overlap → cfg-03 | захардкоженные секреты в исходниках | universal |
| sec-db03 | overlap → pgx-03 | `sslmode=disable` на remote | pgx gate |

## domain security (gated)

| id | status | gate | when | fix | severity |
|----|--------|------|------|-----|----------|
| sec-m01 | implemented | mongo | пароль mongo uri в исходнике | secret manager | fail |
| sec-m02 | implemented | mongo | remote mongo без tls | mongodb+srv или tls=true | fail |
| sec-rd01 | implemented | redis | пароль redis в исходнике | только env | fail |
| sec-rd02 | implemented | redis | redis url на remote без tls | rediss:// или tls options | fail |
| sec-mq01 | implemented | mq | credentials broker url в исходнике | env / vault | fail |
| sec-mq02 | implemented | mq | plaintext amqp/nats на remote broker | amqps или tls dial | fail |

## rest

| id | status | when | fix |
|----|--------|------|-----|
| sec-r01 | implemented | handler берёт `id` из path без authz | проверить, что actor владеет ресурсом |
| sec-r03 | implemented | `json.Unmarshal` в db model из запроса | dto + явные поля |
| sec-r04 | implemented | login без rate limiter | middleware limit |
| sec-r05 | implemented | `/admin` без role middleware | rbac check |
| sec-r06 | implemented | otp/платёж без step-up auth | mfa или step-up перед payout |
| sec-r09 | implemented | `http.Handle` вне cmd setup | только central router |
| sec-r10 | implemented | webhook без hmac verify | signature + replay id |

## graphql

| id | status | when | fix |
|----|--------|------|-----|
| sec-g04 | implemented | `node(id:)` без ownership | authz в resolver |
| sec-g05 | implemented | неограниченный alias batching | cost limit / persisted queries |
| sec-g06 | implemented | cookie auth + get queries | csrf token или post-only |

## database

| id | status | when | fix |
|----|--------|------|-----|
| sec-db05 | implemented | table/column из user input | allowlist |

## опциональный subprocess

`refgrade scan --with-security` запускает `govulncheck` для **sec-16**, если установлен; без флага **sec-16** предупреждает, что CVE scan пропущен

## честные пробелы (в footer отчёта)

| тема | status |
|------|--------|
| live IDOR / BOLA | runtime gap |
| DAST / fuzzing | runtime gap |
| infrastructure IAM / K8s RBAC | runtime gap |
| explain analyze для sql perf | runtime gap |

i18n-ключи: `report.gap.idor`, `report.gap.dast`, `report.gap.k8s` в `locales/*.yaml`
