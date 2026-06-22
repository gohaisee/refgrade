# grpc

**n/a** когда нет grpc/connect сервера

## detected imports

| library | import path |
|---------|-------------|
| grpc-go | `google.golang.org/grpc` |
| connect | `connectrpc.com/connect` |
| grpc-gateway | `github.com/grpc-ecosystem/grpc-gateway/v2` |

## checks

| id | status | when | why | fix | severity |
|----|--------|------|-----|-----|----------|
| grpc-01 | implemented | `grpc.Dial` / insecure credentials вне тестового кода | mitm | tls или `credentials.NewClientTLSFromCert` | fail |
| grpc-02 | implemented | `grpc.NewServer()` без unary/stream interceptors | нет auth/logging/recovery | цепочка interceptors | warn |
| grpc-03 | implemented | тело rpc-метода >50 строк с inline sql | логика в сгенерированном handler | слой service | warn |
| grpc-04 | implemented | metadata `x-user-id` доверяется для authz | подделанные metadata | валидировать jwt; маппить из токена | fail |
| grpc-05 | implemented | reflection включена на публичном listener | утечка поверхности API | reflection только на admin-порту / dev | warn |
| grpc-06 | implemented | нет `GracefulStop` при shutdown | потерянные rpc | обработчик сигналов + graceful stop | info |

## connect

| id | status | when | fix |
|----|--------|------|-----|
| conn-01 | implemented | клиент без опции timeout | задать connect timeout |
| conn-02 | implemented | те же пробелы interceptors, что у grpc-02 | добавить connect interceptors |

## grpc-gateway

| id | status | when | fix |
|----|--------|------|-----|
| gw-01 | implemented | http-мост открывает rpc без того же auth, что у grpc | выровнять auth middleware |
| gw-02 | implemented | gateway зарегистрирован до auth middleware | исправить порядок middleware |
