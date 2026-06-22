# graphql

**n/a** когда нет зависимости graphql-сервера

## detected imports

| library | import path |
|---------|-------------|
| gqlgen | `github.com/99designs/gqlgen` |
| graphql-go | `github.com/graph-gophers/graphql-go` |
| graphql (code-first) | `github.com/graphql-go/graphql` |

## checks

| id | status | when | why | fix | severity |
|----|--------|------|-----|-----|----------|
| gql-01 | implemented | импорт sql/db driver в пакете resolver | resolver ходит в хранилище | делегировать в service/repo | fail |
| gql-02 | implemented | field resolver вызывает бд на каждую строку родителя (эвристика: запрос в resolver без batch) | медленные списки; запрос в цикле | dataloader, join в repo или sql batch | warn |
| gql-03 | implemented | ручное редактирование `generated.go` / `models_gen.go` | перезапишется при regen | менять только schema + `gqlgen generate` | fail |
| gql-04 | implemented | нет лимита complexity или depth в настройке сервера | дорогой query dos | расширение gqlgen или кастомный middleware | warn |
| gql-05 | implemented | introspection включена без env gate | дамп схемы в проде | отключить или закрыть introspection auth | warn |
| gql-06 | implemented | маршрут graphiql/playground зарегистрирован безусловно | публичный UI схемы | build tag или маршрут только для dev | fail |
| gql-07 | implemented | `map[string]interface{}` в рукописных resolvers (gqlgen) | потеря типобезопасности | сгенерированные models | warn |
| gql-08 | implemented | файл resolver >500 строк без разбиения | сложный review | разбить по домену или тонкие делегаты | info |
| gql-09 | implemented | get-based graphql queries без лимита размера | отравление кэша / слишком длинные url | только post или лимит query string | info |

## overlap checks

| id | status | overlap | примечание |
|----|--------|---------|------------|
| sec-g01 | overlap → gql-05 | introspection без env gate | security gate |
| sec-g02 | overlap → gql-04 | нет лимита глубины запроса | security gate |
| sec-g03 | overlap → gql-04 | нет complexity/cost limit | security gate |

## gqlgen-specific

| id | status | when | fix |
|----|--------|------|-----|
| gqlgen-01 | implemented | нет `gqlgen.yml` | добавить конфиг; зафиксировать generate в ci |
| gqlgen-02 | implemented | federation `entity.resolvers` без auth на id | проверять, что актор владеет entity |

## graphql-go / code-first

| id | status | when | fix |
|----|--------|------|-----|
| ggl-01 | implemented | нет panic recovery на границе execute | обернуть executor |
| ggl-02 | implemented | только legacy `graphql-go/graphql` | для нового кода рассмотреть gqlgen (info) |

## security

см. [security-owasp.md](security-owasp.md) — sec-g01…sec-g10
