# rest — gin, echo, chi, net/http

**n/a** когда в модуле нет http-сервера (чистая библиотека, только worker)

## detected imports

| library | import path |
|---------|-------------|
| stdlib | `net/http` |
| gin | `github.com/gin-gonic/gin` |
| echo v4/v5 | `github.com/labstack/echo/v4`, `.../v5` |
| chi | `github.com/go-chi/chi/v5` |
| gorilla/mux (legacy) | `github.com/gorilla/mux` |

## checks

| id | status | when | why | fix | severity |
|----|--------|------|-----|-----|----------|
| rest-01 | implemented | функция handler >80 строк без вызовов service | бизнес-логика в транспорте | вынести в service; handler маппит dto | warn |
| rest-02 | implemented | импорт sql/db в пакете handler | нарушение слоёв | перенести в repo | fail |
| rest-03 | implemented | `http.Server` без `ReadHeaderTimeout` / `ReadTimeout` | slowloris; зависшие соединения | задать таймауты на сервере | warn |
| rest-04 | implemented | нет лимита размера тела на upload/post маршрутах | dos большим телом | `MaxBytesReader` или middleware с лимитом | warn |
| rest-05 | implemented | cors разрешает `*` с `AllowCredentials: true` | уязвимость безопасности в браузере | явный список origin | fail |
| rest-06 | implemented | auth читает заголовок `X-User-Id` / `X-Is-Admin` | тривиальный обход | только jwt/session на стороне сервера | fail |
| rest-07 | implemented | текст внутренней ошибки в json-теле ответа | утечка информации | маппить на безопасные коды; детали логировать на сервере | warn |
| rest-08 | implemented | `gin.SetMode(DebugMode)` вне dev-сборки | stack trace клиентам | release mode через env | warn |
| rest-09 | implemented | gorilla/mux в go.mod без заметки о миграции | неподдерживаемый роутер | chi или stdlib 1.22+ routes | info |
| rest-10 | implemented | запуск http-сервера без recover middleware | panic в handler может уронить процесс | recover middleware на внешнем крае | warn |

## overlap checks

| id | status | overlap | примечание |
|----|--------|---------|------------|
| sec-14 | overlap → rest-05 | cors wildcard + credentials | security gate |

## middleware order (recommended)

снаружи → внутрь: **recover** → request id → logger → timeout → security headers → cors → auth → rate limit → handler

нет recover на внешнем крае → warn `rest-10`

## framework notes

**gin:** явная обработка ошибок binding; не игнорировать ошибки `ShouldBind`

**echo:** подключить `HTTPErrorHandler`; зафиксировать алгоритм jwt в `echo-jwt`

**chi:** использовать `middleware.Timeout`; общий auth middleware на sub-роутерах

**stdlib 1.22+:** предпочитать method-aware `ServeMux`; избегать глобального `http.HandleFunc` в библиотеках
