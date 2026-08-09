# redis

**n/a** когда `github.com/redis/go-redis` нет в модуле (v8/v9)

## detected import

`github.com/redis/go-redis/v9` (и legacy v8)

## checks

| id | when | why | fix | severity |
|----|------|-----|-----|----------|
| redis-01 | `redis.NewClient` на каждый запрос | шторм соединений | singleton или внедрённый клиент | fail |
| redis-02 | ключи кэша без ttl | устаревшие данные навсегда / память | всегда задавать expiration | warn |
| redis-03 | `KEYS *` или `FlushAll` в коде приложения | блокирует redis | `SCAN` только в admin-инструментах | fail |
| redis-04 | cache-aside write обновляет кэш напрямую вместо invalidate | рассинхрон кэша и бд | удалять ключ при записи; чтение перезаполняет | warn |
| redis-05 | решения auth/session только из кэша | устаревшая роль | источник истины в бд; короткий ttl | fail |
| redis-06 | get в цикле без pipeline | высокий rtt | `Pipelined` или `MGet` | info |
| redis-07 | нет префикса namespace для ключей | коллизии между сервисами | паттерн `app:entity:id` | info |
| redis-08 | hot key miss без singleflight или lock | штурм бд | `singleflight` или redis lock при miss | info |

## pool

| id | when | fix |
|----|------|-----|
| redis-09 | `PoolSize` / `MinIdleConns` ноль на prod-пути | настроить pool |

## security

| id | when | fix |
|----|------|-----|
| sec-rd01 | пароль в исходниках | только env |
| sec-rd02 | redis без tls в публичной сети | tls + acl |
