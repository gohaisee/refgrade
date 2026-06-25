# redis

**n/a** when `github.com/redis/go-redis` not in module (v8/v9)

## detected import

`github.com/redis/go-redis/v9` (and legacy v8)

## checks

| id | status | when | why | fix | severity |
|----|--------|------|-----|-----|----------|
| redis-01 | implemented | `redis.NewClient` per request | connection storm | singleton or injected client | fail |
| redis-02 | implemented | cache keys without ttl | stale forever / memory | always set expiration | warn |
| redis-03 | implemented | `KEYS *` or `FlushAll` in app code | blocks redis | `SCAN` for admin tools only | fail |
| redis-04 | implemented | cache-aside write updates cache directly instead of invalidate | inconsistent cache/db | delete key on write; read repopulates | warn |
| redis-05 | implemented | auth/session decisions from cache only | stale role | source of truth in db; short ttl | fail |
| redis-06 | implemented | get in loop without pipeline | high rtt | `Pipelined` or `MGet` | info |
| redis-07 | implemented | no key namespace prefix | collision between services | `app:entity:id` pattern | info |
| redis-08 | implemented | hot key miss without singleflight or lock | stampede on db | `singleflight` or redis lock on miss | info |

## pool

| id | status | when | fix |
|----|--------|------|-----|
| redis-09 | implemented | `PoolSize` / `MinIdleConns` zero on prod path | tune pool |

## security

| id | status | when | fix | severity |
|----|--------|------|-----|----------|
| sec-rd01 | implemented | password in source | env only | fail |
| sec-rd02 | implemented | redis url to remote without tls | rediss:// or tls options | fail |

see [security-owasp.md](security-owasp.md) — sec-rd01, sec-rd02
