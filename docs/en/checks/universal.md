# universal checks

run on every scan unless excluded in `.refgrade.yaml`

## config

| id | when | why | fix | severity |
|----|------|-----|-----|----------|
| cfg-01 | `os.Getenv` outside `cmd/`, tests, `//go:build` integration/e2e | env scattered in code; hard to test and rotate secrets | load in `internal/config`; inject struct in `main` | fail |
| cfg-02 | no dedicated config package in apps with >3 env vars | magic strings for keys | add `internal/config` with struct + validation | warn |
| cfg-03 | hardcoded secret patterns in source (`password=`, `sk-`, `BEGIN PRIVATE KEY`) | leak via git | env/secret manager; never commit values | fail |
| cfg-04 | jwt signing key literal empty string | auth bypass | fail fast at startup if secret missing | fail |

## layout

| id | when | why | fix | severity |
|----|------|-----|-----|----------|
| lay-01 | application without `cmd/<name>/main.go` | entrypoint unclear | move `main` under `cmd/` | warn |
| lay-02 | business packages outside `internal/` without `pkg/` reason | accidental public API | use `internal/` for app code | warn |
| lay-03 | `go.mod` module path looks placeholder (`example.com/foo`, `test`) | import confusion | match real module path | info |

## layering

| id | when | why | fix | severity |
|----|------|-----|-----|----------|
| lyr-01 | `internal/service` imports db driver or ORM | business logic tied to storage | repo/store package; interface in service | fail |
| lyr-02 | handler/resolver/grpc handler imports `internal/repo` | transport knows SQL | call service only | fail |
| lyr-03 | `New*` / `Init` opens network or db inside | hidden side effects; bad tests | wire connections in `main` only | warn |
| lyr-04 | global `var db` in service/handler packages | hidden state; race risk | inject deps via struct | warn |

custom layer rules: `.refgrade.yaml` → `layers[].forbid`

## errors and resilience

| id | when | why | fix | severity |
|----|------|-----|-----|----------|
| err-01 | `_ = err` or empty `if err != nil {}` | silent failures | handle or return wrapped error | warn |
| err-02 | `panic(` in `internal/` (not main/tests) | process crash | return error; recover at transport edge | warn |
| err-03 | `http.DefaultClient` or `http.Get` without custom client | no timeout; hangs | `http.Client{Timeout: ...}` + context | fail |
| err-04 | `http.Client` with zero timeout | same as default | set `Timeout` and transport deadlines | fail |
| err-05 | `context.Background()` in http/grpc handler instead of request ctx | work continues after client disconnect | use `r.Context()` / stream ctx | warn |
| err-06 | `go func()` inside loop without worker limit | goroutine storm | semaphore or worker pool | warn |

## tests

| id | when | why | fix | severity |
|----|------|-----|-----|----------|
| tst-01 | logic file `foo.go` without `foo_test.go` (with exclusions) | regressions undetected | add table-driven test beside source | warn |
| tst-02 | single `all_test.go` covers whole repo layer | hard to navigate failures | split 1:1 with source files | warn |
| tst-03 | `t.Skip()` without build tag or short guard | CI hides broken tests | fix test or use `//go:build integration` | warn |

exclusions for tst-01: `doc.go`, `*_gen.go`, `generated.go`, `mocks/`, thin `main.go`

## observability

| id | when | why | fix | severity |
|----|------|-----|-----|----------|
| obs-01 | `fmt.Print*` in `internal/` | no structure; lost in prod | `slog` / `zap` / `zerolog` | warn |
| obs-02 | log line includes password, token, otp field | credential leak | redact; log ids only | fail |

## concurrency (light)

| id | when | why | fix | severity |
|----|------|-----|-----|----------|
| con-01 | unbounded channel send in request path without drop policy | memory growth under load | bounded buffer + backpressure | info |
