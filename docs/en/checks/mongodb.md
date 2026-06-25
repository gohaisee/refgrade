# mongodb

**n/a** when `go.mongodb.org/mongo-driver` not in module

## detected import

`go.mongodb.org/mongo-driver/mongo`

## checks

| id | status | when | why | fix | severity |
|----|--------|------|-----|-----|----------|
| mongo-01 | implemented | `mongo.Connect` per http/grpc request | socket exhaustion | one `mongo.Client` per process | fail |
| mongo-02 | implemented | db call with `context.Background()` in request path | no cancel on disconnect | pass request ctx | warn |
| mongo-03 | implemented | `Find` without limit on list handlers | unbounded memory | `SetLimit` | warn |
| mongo-04 | implemented | no `Ping` after connect in main | silent misconfig | ping with timeout at startup | info |
| mongo-05 | implemented | pool options unset on high-traffic service | latency under load | `SetMaxPoolSize`, `SetMinPoolSize` | info |
| mongo-06 | implemented | `$where` / js eval with user input | injection | avoid `$where`; use operators | fail |
| mongo-07 | implemented | user-controlled `$regex` without escape | re dos / injection | anchor and escape pattern | warn |

## layout

| id | status | when | fix |
|----|--------|------|-----|
| mongo-08 | implemented | bson handling in handler package | repository package |

## security

| id | status | when | fix | severity |
|----|--------|------|-----|----------|
| sec-m01 | implemented | connection uri with password in committed file | secret manager | fail |
| sec-m02 | implemented | remote mongo uri without tls | mongodb+srv or tls=true | fail |

see [security-owasp.md](security-owasp.md) — sec-m01, sec-m02
