# redis

**n/a** when `github.com/redis/go-redis` not in module (v8/v9)

## detected import

`github.com/redis/go-redis/v9` (and legacy v8)

## checks

| id | when | why | fix | severity |
|----|------|-----|-----|----------|
| redis-01 | `redis.NewClient` per request | connection storm | singleton or injected client | fail |
| redis-02 | cache keys without ttl | stale forever / memory | always set expiration | warn |
| redis-03 | `KEYS *` or `FlushAll` in app code | blocks redis | `SCAN` for admin tools only | fail |
| redis-04 | cache-aside write updates cache directly instead of invalidate | inconsistent cache/db | delete key on write; read repopulates | warn |
| redis-05 | auth/session decisions from cache only | stale role | source of truth in db; short ttl | fail |
| redis-06 | get in loop without pipeline | high rtt | `Pipelined` or `MGet` | info |
| redis-07 | no key namespace prefix | collision between services | `app:entity:id` pattern | info |
| redis-08 | hot key miss without singleflight or lock | stampede on db | `singleflight` or redis lock on miss | info |

## pool

| id | when | fix |
|----|------|-----|
| redis-09 | `PoolSize` / `MinIdleConns` zero on prod path | tune pool |

## security

| id | when | fix |
|----|------|-----|
| sec-rd01 | password in source | env only |
| sec-rd02 | redis without tls on public network | tls + acl |
