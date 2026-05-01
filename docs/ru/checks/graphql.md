# graphql

**n/a** когда нет зависимости graphql-сервера

## detected imports

| library | import path |
|---------|-------------|
| gqlgen | `github.com/99designs/gqlgen` |
| graphql-go | `github.com/graph-gophers/graphql-go` |
| graphql (code-first) | `github.com/graphql-go/graphql` |

## checks

| id | when | why | fix | severity |
|----|------|-----|-----|----------|
| gql-01 | импорт sql/db driver в пакете resolver | resolver ходит в хранилище | делегировать в service/repo | fail |
| gql-02 | field resolver вызывает бд на каждую строку родителя (эвристика: запрос в resolver без batch) | медленные списки; запрос в цикле | dataloader, join в repo или sql batch | warn |
| gql-03 | ручное редактирование `generated.go` / `models_gen.go` | перезапишется при regen | менять только schema + `gqlgen generate` | fail |
| gql-04 | нет лимита complexity или depth в настройке сервера | дорогой query dos | расширение gqlgen или кастомный middleware | warn |
| gql-05 | introspection включена без env gate | дамп схемы в проде | отключить или закрыть introspection auth | warn |
| gql-06 | маршрут graphiql/playground зарегистрирован безусловно | публичный UI схемы | build tag или маршрут только для dev | fail |
| gql-07 | `map[string]interface{}` в рукописных resolvers (gqlgen) | потеря типобезопасности | сгенерированные models | warn |
| gql-08 | файл resolver >500 строк без разбиения | сложный review | разбить по домену или тонкие делегаты | info |
| gql-09 | get-based graphql queries без лимита размера | отравление кэша / слишком длинные url | только post или лимит query string | info |

## gqlgen-specific

| id | when | fix |
|----|------|-----|
| gqlgen-01 | нет `gqlgen.yml` | добавить конфиг; зафиксировать generate в ci |
| gqlgen-02 | federation `entity.resolvers` без auth на id | проверять, что актор владеет entity |

## graphql-go / code-first

| id | when | fix |
|----|------|-----|
| ggl-01 | нет panic recovery на границе execute | обернуть executor |
| ggl-02 | только legacy `graphql-go/graphql` | для нового кода рассмотреть gqlgen (info) |

## security overlap

см. [security-owasp.md](security-owasp.md) — sec-g01…sec-g10
