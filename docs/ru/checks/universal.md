# универсальные проверки

запускаются при каждом скане, если не исключены в `.refgrade.yaml`

## config

| id | when | why | fix | severity |
|----|------|-----|-----|----------|
| cfg-01 | `os.Getenv` вне `cmd/`, тестов, `//go:build` integration/e2e | env разбросан по коду; сложно тестировать и ротировать секреты | загружать в `internal/config`; передавать struct из `main` | fail |
| cfg-02 | нет отдельного пакета конфигурации в приложениях с >3 env-переменными | magic strings для ключей | добавить `internal/config` со struct и валидацией | warn |
| cfg-03 | захардкоженные паттерны секретов в исходниках (`password=`, `sk-`, `BEGIN PRIVATE KEY`) | утечка через git | env/secret manager; никогда не коммитить значения | fail |
| cfg-04 | литерал ключа подписи jwt — пустая строка | обход аутентификации | падать при старте, если секрет не задан | fail |

## layout

| id | when | why | fix | severity |
|----|------|-----|-----|----------|
| lay-01 | приложение без `cmd/<name>/main.go` | точка входа неочевидна | перенести `main` под `cmd/` | warn |
| lay-02 | бизнес-пакеты вне `internal/` без причины для `pkg/` | случайный публичный API | использовать `internal/` для кода приложения | warn |
| lay-03 | путь модуля в `go.mod` похож на заглушку (`example.com/foo`, `test`) | путаница с импортами | указать реальный путь модуля | info |

## layering

| id | when | why | fix | severity |
|----|------|-----|-----|----------|
| lyr-01 | `internal/service` импортирует драйвер бд или ORM | бизнес-логика привязана к хранилищу | пакет repo/store; интерфейс в service | fail |
| lyr-02 | handler/resolver/grpc handler импортирует `internal/repo` | транспорт знает про SQL | вызывать только service | fail |
| lyr-03 | `New*` / `Init` открывает сеть или бд внутри | скрытые побочные эффекты; плохие тесты | подключения только в `main` | warn |
| lyr-04 | глобальный `var db` в пакетах service/handler | скрытое состояние; риск гонок | внедрять зависимости через struct | warn |

кастомные правила слоёв: `.refgrade.yaml` → `layers[].forbid`

## errors and resilience

| id | when | why | fix | severity |
|----|------|-----|-----|----------|
| err-01 | `_ = err` или пустой `if err != nil {}` | тихие сбои | обработать или вернуть обёрнутую ошибку | warn |
| err-02 | `panic(` в `internal/` (не main/тесты) | падение процесса | возвращать error; recover на границе транспорта | warn |
| err-03 | `http.DefaultClient` или `http.Get` без кастомного клиента | нет таймаута; зависания | `http.Client{Timeout: ...}` + context | fail |
| err-04 | `http.Client` с нулевым таймаутом | то же, что у дефолтного | задать `Timeout` и дедлайны транспорта | fail |
| err-05 | `context.Background()` в http/grpc handler вместо ctx запроса | работа продолжается после отключения клиента | использовать `r.Context()` / ctx стрима | warn |
| err-06 | `go func()` внутри цикла без лимита воркеров | шторм горутин | семафор или пул воркеров | warn |

## tests

| id | when | why | fix | severity |
|----|------|-----|-----|----------|
| tst-01 | файл логики `foo.go` без `foo_test.go` (с исключениями) | регрессии не ловятся | добавить table-driven тест рядом с исходником | warn |
| tst-02 | один `all_test.go` покрывает весь слой репозитория | сложно разбирать падения | разбить 1:1 с исходными файлами | warn |
| tst-03 | `t.Skip()` без build tag или short guard | CI скрывает сломанные тесты | починить тест или использовать `//go:build integration` | warn |

исключения для tst-01: `doc.go`, `*_gen.go`, `generated.go`, `mocks/`, тонкий `main.go`

## observability

| id | when | why | fix | severity |
|----|------|-----|-----|----------|
| obs-01 | `fmt.Print*` в `internal/` | нет структуры; теряется в проде | `slog` / `zap` / `zerolog` | warn |
| obs-02 | строка лога содержит password, token, otp | утечка учётных данных | редактировать; логировать только id | fail |

## concurrency (light)

| id | when | why | fix | severity |
|----|------|-----|-----|----------|
| con-01 | неограниченная отправка в канал на пути запроса без политики отбрасывания | рост памяти под нагрузкой | ограниченный буфер + backpressure | info |
