# конфиг

опциональный файл: `.refgrade.yaml` в корне модуля

## пример

```yaml
lang: ru          # язык отчёта по умолчанию: en | ru
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

## язык

- `lang` в yaml
- флаг `--lang en|ru`
- env `REFGRADE_LANG`

приоритет: флаг > env > yaml > `en`
