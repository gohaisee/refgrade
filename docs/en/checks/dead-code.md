# dead code

always run for modules with `cmd/` or testable `internal/` packages

## tools

| tool | role |
|------|------|
| `golang.org/x/tools/cmd/deadcode` | unreachable functions from entrypoints (rta) |
| staticcheck `U1000` | unused types, vars, funcs per package |
| `go mod tidy -diff` | unused module dependencies |

## checks

| id | status | when | why | fix | severity |
|----|--------|------|-----|-----|----------|
| dead-01 | implemented | function reported unreachable by `deadcode` | never called; maintenance drag | delete or wire up | warn |
| dead-02 | implemented | exported symbol in `internal/` unreachable even with tests | dead api surface | remove | info |
| dead-03 | implemented | `U1000` unused func/type/const | clutter | delete | warn |
| dead-04 | implemented | `.go` file not in any package build (orphan) | confusion | remove or fix package | warn |
| dead-05 | implemented | empty package (only `package foo`) | mistake | delete package | warn |
| dead-06 | implemented | require in go.mod unused by imports | supply chain noise | tidy | warn |
| dead-07 | implemented | large commented-out code blocks | hides real logic | delete (git history keeps it) | warn |
| dead-08 | implemented | test helper only used in one file but exported | narrow visibility | unexport | info |

## false positives (document in finding, don't auto-fail)

- `//go:generate` targets
- reflection (`reflect.TypeOf`, plugins)
- build tags — run scan with multiple `GOOS`/`tags` or mark `needs-review`
- `main` in multi-module repo — scope scan to one module root

## cli

| command | behavior |
|---------|----------|
| `refgrade scan` | dead-01…08 (dead-01/03/04/05/06/07 warn; dead-02/08 info) |
| `refgrade deadcode` | not a separate command — dead checks run inside `scan` when `deadcode` and `staticcheck` are on PATH |

## fixtures

see `testdata/fixtures/` — mini repos that must trigger specific dead-* ids
