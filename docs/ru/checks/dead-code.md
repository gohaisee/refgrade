# мёртвый код

всегда для модулей с `cmd/` или тестируемыми пакетами `internal/`

## инструменты

| tool | роль |
|------|------|
| `golang.org/x/tools/cmd/deadcode` | недостижимые func от entrypoints (rta) |
| staticcheck `U1000` | неиспользуемые types, vars, funcs в пакете |
| `go mod tidy -diff` | лишние зависимости в go.mod |

## проверки

| id | status | when | why | fix | severity |
|----|--------|------|-----|-----|----------|
| dead-01 | implemented | func недостижима по `deadcode` | никто не вызывает; мешает сопровождению | удалить или подключить | warn |
| dead-02 | implemented | exported symbol в `internal/` недостижим даже с тестами | мёртвая поверхность API | удалить | info |
| dead-03 | implemented | `U1000` unused func/type/const | шум | удалить | warn |
| dead-04 | implemented | `.go` файл вне сборки пакета (orphan) | путаница | удалить или починить package | warn |
| dead-05 | implemented | пустой package (только `package foo`) | ошибка | удалить package | warn |
| dead-06 | implemented | require в go.mod без импортов | шум supply chain | tidy | warn |
| dead-07 | implemented | большие закомментированные блоки | скрывает логику | удалить (история в git) | warn |
| dead-08 | implemented | test helper exported, но нужен в одном файле | лишняя видимость | unexport | info |

## ложные срабатывания (в finding пометить, не auto-fail)

- цели `//go:generate`
- reflection (`reflect.TypeOf`, plugins)
- build tags — гонять с разными `GOOS`/`tags` или `needs-review`
- `main` в multi-module repo — scope скана на один корень модуля

## cli

| command | поведение |
|---------|-----------|
| `refgrade scan` | dead-01…08 (dead-01/03/04/05/06/07 warn; dead-02/08 info) |
| `refgrade deadcode` | отдельной командой нет — dead-проверки в `scan`, если `deadcode` и `staticcheck` на PATH |

## fixtures

см. `testdata/fixtures/` — mini-repos под конкретные dead-* id
