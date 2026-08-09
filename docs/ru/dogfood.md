# dogfood — скан refgrade на себе

ручной dogfood через локальные `testdata/fixtures/` (без network clone)

## prerequisites

```bash
go install golang.org/x/tools/cmd/deadcode@latest
go install honnef.co/go/tools/cmd/staticcheck@latest
go build -o bin/refgrade ./cmd/refgrade
export PATH="$PWD/bin:$PATH"
```

## скан fixture (минимальный паттерн)

```bash
bin/refgrade scan testdata/fixtures/bad-weak-crypto
bin/refgrade scan testdata/fixtures/bad-math-rand-token
bin/refgrade scan testdata/fixtures/bad-dead-code
```

ожидается: `sec-01` или `sec-03` на crypto-fixtures; `dead-01`…`dead-08` на dead-code fixture

## self-scan (корень репозитория)

```bash
bin/refgrade scan .
bin/refgrade scan . --lang ru
bin/refgrade scan . --format json -o /tmp/refgrade-dogfood.json
```

записать summary и неожиданные `fail` ниже

## шаблон результатов

| дата | target | lang | fail | warn | notes |
|------|--------|------|------|------|-------|
| YYYY-MM-DD | `testdata/fixtures/bad-weak-crypto` | en | | | |
| YYYY-MM-DD | `testdata/fixtures/bad-dead-code` | en | | | |
| YYYY-MM-DD | `.` (self) | en | | | footer с IDOR/DAST/K8s gaps |
| YYYY-MM-DD | `.` (self) | ru | | | |

## проверка footer

footer отчёта должен перечислять runtime gaps из i18n:

- `report.gap.idor` — live IDOR/BOLA
- `report.gap.dast` — DAST / fuzzing
- `report.gap.k8s` — K8s IAM / infrastructure RBAC

см. [checks/security-owasp.md](checks/security-owasp.md#честные-пробелы-в-footer-отчёта)

## паритет с ci

локальный coverage gate (тот же порог, что в CI):

```bash
go test ./internal/check/... -coverprofile=coverage.out
go tool cover -func=coverage.out | awk '/total:/'
# должно быть ≥ 75%
```
