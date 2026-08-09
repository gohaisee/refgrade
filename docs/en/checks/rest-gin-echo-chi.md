# rest — gin, echo, chi, net/http

**n/a** when module has no http server (pure library, worker-only)

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
| rest-01 | implemented | handler function >80 lines without service calls | business logic in transport | extract to service; handler maps dto | warn |
| rest-02 | implemented | sql/db import in handler package | layer violation | move to repo | fail |
| rest-03 | implemented | `http.Server` without `ReadHeaderTimeout` / `ReadTimeout` | slowloris; hung connections | set timeouts on server | warn |
| rest-04 | implemented | no max body size on upload/post routes | large body dos | `MaxBytesReader` or middleware limit | warn |
| rest-05 | implemented | cors allows `*` with `AllowCredentials: true` | browser security bug | explicit origin list | fail |
| rest-06 | implemented | auth reads `X-User-Id` / `X-Is-Admin` header | trivial bypass | jwt/session server-side only | fail |
| rest-07 | implemented | internal error text returned in json body | info leak | map to safe codes; log detail server-side | warn |
| rest-08 | implemented | `gin.SetMode(DebugMode)` in non-dev build path | stack traces to clients | release mode via env | warn |
| rest-09 | implemented | gorilla/mux in go.mod without migration note | unmaintained router | chi or stdlib 1.22+ routes | info |
| rest-10 | implemented | http server setup without recover middleware | handler panic can crash the process | add recover middleware at the outer edge | warn |

## overlap checks

| id | status | overlap | notes |
|----|--------|---------|-------|
| sec-14 | overlap → rest-05 | cors wildcard + credentials | security gate |

## middleware order (recommended)

outer → inner: **recover** → request id → logger → timeout → security headers → cors → auth → rate limit → handler

missing recover at outer edge → warn `rest-10`

## framework notes

**gin:** prefer explicit binding error handling; don't ignore `ShouldBind` errors

**echo:** wire `HTTPErrorHandler`; pin jwt algorithm in `echo-jwt`

**chi:** use `middleware.Timeout`; share auth middleware on sub-routers

**stdlib 1.22+:** prefer method-aware `ServeMux`; avoid global `http.HandleFunc` in libraries
