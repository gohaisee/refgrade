# grpc

**n/a** when no grpc/connect server

## detected imports

| library | import path |
|---------|-------------|
| grpc-go | `google.golang.org/grpc` |
| connect | `connectrpc.com/connect` |
| grpc-gateway | `github.com/grpc-ecosystem/grpc-gateway/v2` |

## checks

| id | when | why | fix | severity |
|----|------|-----|-----|----------|
| grpc-01 | `grpc.Dial` / insecure credentials in non-test code | mitm | tls or `credentials.NewClientTLSFromCert` | fail |
| grpc-02 | `grpc.NewServer()` without unary/stream interceptors | no auth/logging/recovery | chain interceptors | warn |
| grpc-03 | rpc method body >50 lines with sql inline | logic in generated handler | service layer | warn |
| grpc-04 | metadata `x-user-id` trusted for authz | spoofed metadata | validate jwt; map from token | fail |
| grpc-05 | reflection enabled on public listener | api surface leak | reflection only on admin port / dev | warn |
| grpc-06 | missing `GracefulStop` on shutdown | dropped rpcs | signal handler + graceful stop | info |

## connect

| id | when | fix |
|----|------|-----|
| conn-01 | client without timeout option | set connect timeout |
| conn-02 | same interceptor gaps as grpc-02 | add connect interceptors |

## grpc-gateway

| id | when | fix |
|----|------|-----|
| gw-01 | http bridge exposes rpc without same auth as grpc | align auth middleware |
| gw-02 | gateway registered before auth middleware | fix middleware order |
