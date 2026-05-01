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
| sec-13 | лог password/token/otp | редактировать | fail |
| sec-14 | cors `*` + credentials | явный список origins | fail |
| sec-15 | `net/http/pprof` без build tag | только dev | warn |
| sec-16 | находки `govulncheck` (флаг `--with-security`) | обновить dep | fail |

## rest

| id | when | fix |
|----|------|-----|
| sec-r01 | handler берёт `id` из path без authz | проверить, что actor владеет ресурсом |
| sec-r03 | `json.Unmarshal` в db model из запроса | dto + явные поля |
| sec-r04 | login без rate limiter | middleware limit |
| sec-r05 | `/admin` без role middleware | rbac check |
| sec-r10 | webhook без hmac verify | signature + replay id |

## graphql

| id | when | fix |
|----|------|-----|
| sec-g01 | introspection включён в prod build | gate по env |
| sec-g02 | нет лимита глубины запроса | server limit |
| sec-g03 | нет complexity/cost analysis | extension |
| sec-g04 | `node(id:)` без ownership | authz в resolver |
| sec-g05 | неограниченный alias batching | cost limit / persisted queries |
| sec-g06 | cookie auth + get queries | csrf token или post-only |

## database

| id | when | fix |
|----|------|-----|
| sec-db01 | sql из конкатенации строк | parameters |
| sec-db02 | пароль dsn в репо | env |
| sec-db03 | `sslmode=disable` на remote | tls |
| sec-db05 | table/column из user input | allowlist |

## опциональный subprocess

`refgrade scan --with-security` запускает `govulncheck`, если установлен; не заменяет ручной review

## честные пробелы (в footer отчёта)

- live idor / bola
- dast / fuzzing
- infrastructure iam, k8s rbac
- explain analyze для sql perf
