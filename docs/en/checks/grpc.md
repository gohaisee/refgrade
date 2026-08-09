# grpc

**n/a** when no grpc/connect server

## detected imports

| library | import path |
|---------|-------------|
| grpc-go | `google.golang.org/grpc` |
| connect | `connectrpc.com/connect` |
| grpc-gateway | `github.com/grpc-ecosystem/grpc-gateway/v2` |

## checks

| id | status | when | why | fix | severity |
|----|--------|------|-----|-----|----------|
| grpc-01 | implemented | `grpc.Dial` / insecure credentials in non-test code | mitm | tls or `credentials.NewClientTLSFromCert` | fail |
| grpc-02 | implemented | `grpc.NewServer()` without unary/stream interceptors | no auth/logging/recovery | chain interceptors | warn |
| grpc-03 | implemented | rpc method body >50 lines with sql inline | logic in generated handler | service layer | warn |
| grpc-04 | implemented | metadata `x-user-id` trusted for authz | spoofed metadata | validate jwt; map from token | fail |
| grpc-05 | implemented | reflection enabled on public listener | api surface leak | reflection only on admin port / dev | warn |
| grpc-06 | implemented | missing `GracefulStop` on shutdown | dropped rpcs | signal handler + graceful stop | info |

## connect

| id | status | when | fix |
|----|--------|------|-----|
| conn-01 | implemented | client without timeout option | set connect timeout |
| conn-02 | implemented | same interceptor gaps as grpc-02 | add connect interceptors |

## grpc-gateway

| id | status | when | fix |
|----|--------|------|-----|
| gw-01 | implemented | http bridge exposes rpc without same auth as grpc | align auth middleware |
| gw-02 | implemented | gateway registered before auth middleware | fix middleware order |
