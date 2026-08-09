# config

optional file: `.refgrade.yaml` in module root

## example

```yaml
lang: en          # default report language: en | ru
profile: standard # minimal | standard | strict
kind: auto        # auto | library | http-api | graphql | grpc | worker

layers:
  - path: internal/service
    forbid:
      - github.com/jackc/pgx/v5
      - gorm.io/gorm

checks:
  getenv: fail
  test_pairing: warn

exclude:
  - "**/generated/**"
  - "**/*_gen.go"
```

## language

- `lang` in yaml
- flag `--lang en|ru`
- env `REFGRADE_LANG`

precedence: flag > env > yaml > `en`
