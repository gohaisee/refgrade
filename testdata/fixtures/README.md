# fixtures

mini go modules used in tests — each directory has its own `go.mod` and targets specific check ids

| fixture | checks |
|---------|--------|
| `bad-getenv/` | cfg-01 (fail) — `os.Getenv` in `internal/service` |
| `bad-default-client/` | err-03 (fail) — `http.DefaultClient` in `internal/service` |
| `bad-ignored-error/` | err-01 (warn) — `_ = err` and empty `if err != nil {}` |
| `good-minimal/` | baseline — getenv only in `cmd/` and `internal/config` |
| `bad-layers/` | lyr-01, lyr-02 (planned) |
| `bad-sql-loop/` | sql-01 (planned) |
| `bad-gql-resolver/` | gql-01 (planned) |
| `bad-dead-code/` | dead-01 (planned) |

run scanner against a fixture:

```bash
go run ./cmd/refgrade scan testdata/fixtures/bad-getenv
```
