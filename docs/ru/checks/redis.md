# redis

**n/a** когда `github.com/redis/go-redis` нет в модуле (v8/v9)

## detected import

`github.com/redis/go-redis/v9` (и legacy v8)

## checks

| id | status | when | why | fix | severity |
|----|--------|------|-----|-----|----------|
| redis-01 | implemented | `redis.NewClient` на каждый запрос | шторм соединений | singleton или внедрённый клиент | fail |
| redis-02 | implemented | ключи кэша без ttl | устаревшие данные навсегда / память | всегда задавать expiration | warn |
| redis-03 | implemented | `KEYS *` или `FlushAll` в коде приложения | блокирует redis | `SCAN` только в admin-инструментах | fail |
| redis-04 | implemented | cache-aside write обновляет кэш напрямую вместо invalidate | рассинхрон кэша и бд | удалять ключ при записи; чтение перезаполняет | warn |
| redis-05 | implemented | решения auth/session только из кэша | устаревшая роль | источник истины в бд; короткий ttl | fail |
| redis-06 | implemented | get в цикле без pipeline | высокий rtt | `Pipelined` или `MGet` | info |
| redis-07 | implemented | нет префикса namespace для ключей | коллизии между сервисами | паттерн `app:entity:id` | info |
| redis-08 | implemented | hot key miss без singleflight или lock | штурм бд | `singleflight` или redis lock при miss | info |

## pool

| id | status | when | fix |
|----|--------|------|-----|
| redis-09 | implemented | `PoolSize` / `MinIdleConns` ноль на prod-пути | настроить pool |

## security

| id | status | when | fix | severity |
|----|--------|------|-----|----------|
| sec-rd01 | implemented | пароль в исходниках | только env | fail |
| sec-rd02 | implemented | redis url на remote без tls | rediss:// или tls options | fail |

см. [security-owasp.md](security-owasp.md) — sec-rd01, sec-rd02
