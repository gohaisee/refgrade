# grpc

**n/a** когда нет grpc/connect сервера

## detected imports

| library | import path |
|---------|-------------|
| grpc-go | `google.golang.org/grpc` |
| connect | `connectrpc.com/connect` |
| grpc-gateway | `github.com/grpc-ecosystem/grpc-gateway/v2` |

## checks

| id | when | why | fix | severity |
|----|------|-----|-----|----------|
| grpc-01 | `grpc.Dial` / insecure credentials вне тестового кода | mitm | tls или `credentials.NewClientTLSFromCert` | fail |
| grpc-02 | `grpc.NewServer()` без unary/stream interceptors | нет auth/logging/recovery | цепочка interceptors | warn |
| grpc-03 | тело rpc-метода >50 строк с inline sql | логика в сгенерированном handler | слой service | warn |
| grpc-04 | metadata `x-user-id` доверяется для authz | подделанные metadata | валидировать jwt; маппить из токена | fail |
| grpc-05 | reflection включена на публичном listener | утечка поверхности API | reflection только на admin-порту / dev | warn |
| grpc-06 | нет `GracefulStop` при shutdown | потерянные rpc | обработчик сигналов + graceful stop | info |

## connect

| id | when | fix |
|----|------|-----|
| conn-01 | клиент без опции timeout | задать connect timeout |
| conn-02 | те же пробелы interceptors, что у grpc-02 | добавить connect interceptors |

## grpc-gateway

| id | when | fix |
|----|------|-----|
| gw-01 | http-мост открывает rpc без того же auth, что у grpc | выровнять auth middleware |
| gw-02 | gateway зарегистрирован до auth middleware | исправить порядок middleware |
