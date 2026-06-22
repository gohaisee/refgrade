# dogfood — scanning refgrade on itself

manual dogfood run using local `testdata/fixtures/` (no network clone required)

## prerequisites

```bash
go install golang.org/x/tools/cmd/deadcode@latest
go install honnef.co/go/tools/cmd/staticcheck@latest
go build -o bin/refgrade ./cmd/refgrade
export PATH="$PWD/bin:$PATH"
```

## fixture scan (tiny public pattern)

scan a minimal bad fixture that triggers known checks:

```bash
bin/refgrade scan testdata/fixtures/bad-weak-crypto
bin/refgrade scan testdata/fixtures/bad-math-rand-token
bin/refgrade scan testdata/fixtures/bad-dead-code
```

expected: `sec-01` or `sec-03` on crypto fixtures; `dead-01`…`dead-08` on dead-code fixture

## self-scan (refgrade repo root)

```bash
bin/refgrade scan .
bin/refgrade scan . --lang ru
bin/refgrade scan . --format json -o /tmp/refgrade-dogfood.json
```

record summary line and any unexpected `fail` findings below

## results template

| date | target | lang | fail | warn | notes |
|------|--------|------|------|------|-------|
| YYYY-MM-DD | `testdata/fixtures/bad-weak-crypto` | en | | | |
| YYYY-MM-DD | `testdata/fixtures/bad-dead-code` | en | | | |
| YYYY-MM-DD | `.` (self) | en | | | footer shows IDOR/DAST/K8s gaps |
| YYYY-MM-DD | `.` (self) | ru | | | |

## footer check

report footer must list runtime gaps from i18n keys:

- `report.gap.idor` — live IDOR/BOLA reproduction
- `report.gap.dast` — DAST / fuzzing
- `report.gap.k8s` — K8s IAM / infrastructure RBAC

see [checks/security-owasp.md](checks/security-owasp.md#honest-gaps-printed-in-report-footer)

## ci parity

local coverage gate (same threshold as CI):

```bash
go test ./internal/check/... -coverprofile=coverage.out
go tool cover -func=coverage.out | awk '/total:/'
# must be ≥ 75%
```
