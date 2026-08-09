# refgrade

cli that scans a go repo and prints a refactor readiness report (`ok` / `warn` / `fail` / `n/a`)

## docs

| language | path |
|----------|------|
| english | [docs/en/README.md](docs/en/README.md) |
| русский | [docs/ru/README.md](docs/ru/README.md) |

check catalog: [docs/en/checks.md](docs/en/checks.md) · [docs/ru/checks.md](docs/ru/checks.md)

## quick start

```bash
go run ./cmd/refgrade scan ./...
go run ./cmd/refgrade scan testdata/fixtures/bad-getenv          # exit 1, cfg-01
go run ./cmd/refgrade scan testdata/fixtures/bad-default-client  # exit 1, err-03
go run ./cmd/refgrade scan testdata/fixtures/bad-ignored-error    # exit 0, err-01 warn
go run ./cmd/refgrade scan testdata/fixtures/good-minimal        # exit 0
go run ./cmd/refgrade detect .
go run ./cmd/refgrade explain cfg-01 --lang ru
go run ./cmd/refgrade explain err-01 --lang ru
go test ./...
```

## status

**phase 3** dead code (dead-01…08) + stack catalog — see [docs/en/checks.md](docs/en/checks.md)

`scan` exits **1** only on `fail` findings; `warn` prints in the report but exit **0**

## languages

`en` and `ru` — precedence: `--lang` > `REFGRADE_LANG` > `.refgrade.yaml` `lang:` > `en` — [docs/en/i18n.md](docs/en/i18n.md)
