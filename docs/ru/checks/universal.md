# универсальные проверки

запускаются при каждом скане, если не исключены в `.refgrade.yaml`

## config

| id | status | when | why | fix | severity |
|----|--------|------|-----|-----|----------|
| cfg-01 | implemented | `os.Getenv` вне `cmd/`, тестов, `//go:build` integration/e2e | env разбросан по коду; сложно тестировать и ротировать секреты | загружать в `internal/config`; передавать struct из `main` | fail |
| cfg-02 | implemented | нет отдельного пакета конфигурации в приложениях с >3 env-переменными | magic strings для ключей | добавить `internal/config` со struct и валидацией | warn |
| cfg-03 | implemented | захардкоженные паттерны секретов в исходниках (`password=`, `sk-`, `BEGIN PRIVATE KEY`) | утечка через git | env/secret manager; никогда не коммитить значения | fail |
| cfg-04 | implemented | литерал ключа подписи jwt — пустая строка | обход аутентификации | падать при старте, если секрет не задан | fail |

## layout

| id | status | when | why | fix | severity |
|----|--------|------|-----|-----|----------|
| lay-01 | implemented | приложение без `cmd/<name>/main.go` | точка входа неочевидна | перенести `main` под `cmd/` | warn |
| lay-02 | implemented | бизнес-пакеты вне `internal/` без причины для `pkg/` | случайный публичный API | использовать `internal/` для кода приложения | warn |
| lay-03 | implemented | путь модуля в `go.mod` похож на заглушку (`example.com/foo`, `test`) | путаница с импортами | указать реальный путь модуля | info |

## layering

| id | status | when | why | fix | severity |
|----|--------|------|-----|-----|----------|
| lyr-01 | implemented | `internal/service` импортирует драйвер бд или ORM | бизнес-логика привязана к хранилищу | пакет repo/store; интерфейс в service | fail |
| lyr-02 | implemented | handler/resolver/grpc handler импортирует `internal/repo` | транспорт знает про SQL | вызывать только service | fail |
| lyr-03 | implemented | `New*` / `Init` открывает сеть или бд внутри | скрытые побочные эффекты; плохие тесты | подключения только в `main` | warn |
| lyr-04 | implemented | глобальный `var db` в пакетах service/handler | скрытое состояние; риск гонок | внедрять зависимости через struct | warn |

кастомные правила слоёв: `.refgrade.yaml` → `layers[].forbid`

## errors and resilience

| id | status | when | why | fix | severity |
|----|--------|------|-----|-----|----------|
| err-01 | implemented | `_ = err` или пустой `if err != nil {}` | тихие сбои | обработать или вернуть обёрнутую ошибку | warn |
| err-02 | implemented | `panic(` в `internal/` (не main/тесты) | падение процесса | возвращать error; recover на границе транспорта | warn |
| err-03 | implemented | `http.DefaultClient` или `http.Get` без кастомного клиента | нет таймаута; зависания | `http.Client{Timeout: ...}` + context | fail |
| err-04 | implemented | `http.Client` с нулевым таймаутом | то же, что у дефолтного | задать `Timeout` и дедлайны транспорта | fail |
| err-05 | implemented | `context.Background()` в http/grpc handler вместо ctx запроса | работа продолжается после отключения клиента | использовать `r.Context()` / ctx стрима | warn |
| err-06 | implemented | `go func()` внутри цикла без лимита воркеров | шторм горутин | семафор или пул воркеров | warn |

## tests

| id | status | when | why | fix | severity |
|----|--------|------|-----|-----|----------|
| tst-01 | implemented | файл логики `foo.go` без `foo_test.go` (с исключениями) | регрессии не ловятся | добавить table-driven тест рядом с исходником | warn |
| tst-02 | implemented | один `all_test.go` покрывает весь слой репозитория | сложно разбирать падения | разбить 1:1 с исходными файлами | warn |
| tst-03 | implemented | `t.Skip()` без build tag или short guard | CI скрывает сломанные тесты | починить тест или использовать `//go:build integration` | warn |

исключения для tst-01: `doc.go`, `*_gen.go`, `generated.go`, `mocks/`, тонкий `main.go`

## observability

| id | status | when | why | fix | severity |
|----|--------|------|-----|-----|----------|
| obs-01 | implemented | `fmt.Print*` в `internal/` | нет структуры; теряется в проде | `slog` / `zap` / `zerolog` | warn |
| obs-02 | implemented | строка лога содержит password, token, otp | утечка учётных данных | редактировать; логировать только id | fail |

## concurrency (light)

| id | status | when | why | fix | severity |
|----|--------|------|-----|-----|----------|
| con-01 | implemented | неограниченная отправка в канал на пути запроса без политики отбрасывания | рост памяти под нагрузкой | ограниченный буфер + backpressure | info |
