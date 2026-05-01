# mongodb

**n/a** when `go.mongodb.org/mongo-driver` not in module

## detected import

`go.mongodb.org/mongo-driver/mongo`

## checks

| id | when | why | fix | severity |
|----|------|-----|-----|----------|
| mongo-01 | `mongo.Connect` per http/grpc request | socket exhaustion | one `mongo.Client` per process | fail |
| mongo-02 | db call with `context.Background()` in request path | no cancel on disconnect | pass request ctx | warn |
| mongo-03 | `Find` without limit on list handlers | unbounded memory | `SetLimit` | warn |
| mongo-04 | no `Ping` after connect in main | silent misconfig | ping with timeout at startup | info |
| mongo-05 | pool options unset on high-traffic service | latency under load | `SetMaxPoolSize`, `SetMinPoolSize` | info |
| mongo-06 | `$where` / js eval with user input | injection | avoid `$where`; use operators | fail |
| mongo-07 | user-controlled `$regex` without escape | re dos / injection | anchor and escape pattern | warn |

## layout

| id | when | fix |
|----|------|-----|
| mongo-08 | bson handling in handler package | repository package |

## security

| id | when | fix |
|----|------|-----|
| sec-m01 | connection uri with password in committed file | secret manager |
| sec-m02 | tls disabled for remote cluster | enable tls in uri |
