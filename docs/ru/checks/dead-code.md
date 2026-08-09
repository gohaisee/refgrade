# мёртвый код

всегда для модулей с `cmd/` или тестируемыми пакетами `internal/`

## инструменты

| tool | роль |
|------|------|
| `golang.org/x/tools/cmd/deadcode` | недостижимые func от entrypoints (rta) |
| staticcheck `U1000` | неиспользуемые types, vars, funcs в пакете |
| `go mod tidy -diff` | лишние зависимости в go.mod |

## проверки

| id | when | why | fix | severity |
|----|------|-----|-----|----------|
| dead-01 | func недостижима по `deadcode` | никто не вызывает; мешает сопровождению | удалить или подключить | warn |
| dead-02 | exported symbol в `internal/` недостижим даже с тестами | мёртвая поверхность API | удалить | info |
| dead-03 | `U1000` unused func/type/const | шум | удалить |
| dead-04 | `.go` файл вне сборки пакета (orphan) | путаница | удалить или починить package |
| dead-05 | пустой package (только `package foo`) | ошибка | удалить package |
| dead-06 | require в go.mod без импортов | шум supply chain | tidy |
| dead-07 | большие закомментированные блоки | скрывает логику | удалить (история в git) |
| dead-08 | test helper exported, но нужен в одном файле | лишняя видимость | unexport |

## ложные срабатывания (в finding пометить, не auto-fail)

- цели `//go:generate`
- reflection (`reflect.TypeOf`, plugins)
- build tags — гонять с разными `GOOS`/`tags` или `needs-review`
- `main` в multi-module repo — scope скана на один корень модуля

## cli

| command | поведение |
|---------|-----------|
| `refgrade scan` | dead-01…08 (dead-01/03/04/05/06/07 warn; dead-02/08 info) |
| `refgrade deadcode` | только dead-01…08; `--include-tests` передаёт `-test` в `deadcode`; `--tags` для go list и subprocess |

## область scan

| flag | поведение |
|------|-----------|
| `[path]` | подъём к ближайшему `go.mod` от path |
| `--module path` | явный корень модуля |
| `--all-modules` | scan вложенных `go.mod` под path (лимит глубины 8) |
| `--tags integration,e2e` | build tags для `go list` и subprocess |

## fixtures

см. `testdata/fixtures/` — mini-repos под конкретные dead-* id
