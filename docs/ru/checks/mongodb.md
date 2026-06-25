# mongodb

**n/a** когда `go.mongodb.org/mongo-driver` нет в модуле

## detected import

`go.mongodb.org/mongo-driver/mongo`

## checks

| id | status | when | why | fix | severity |
|----|--------|------|-----|-----|----------|
| mongo-01 | implemented | `mongo.Connect` на каждый http/grpc запрос | исчерпание сокетов | один `mongo.Client` на процесс | fail |
| mongo-02 | implemented | вызов бд с `context.Background()` на пути запроса | нет отмены при disconnect | передавать ctx запроса | warn |
| mongo-03 | implemented | `Find` без limit в list handlers | неограниченная память | `SetLimit` | warn |
| mongo-04 | implemented | нет `Ping` после connect в main | тихая misconfig | ping с таймаутом при старте | info |
| mongo-05 | implemented | опции pool не заданы на высоконагруженном сервисе | латентность под нагрузкой | `SetMaxPoolSize`, `SetMinPoolSize` | info |
| mongo-06 | implemented | `$where` / js eval с пользовательским вводом | инъекция | избегать `$where`; использовать операторы | fail |
| mongo-07 | implemented | пользовательский `$regex` без escape | re dos / инъекция | якорить и экранировать паттерн | warn |

## layout

| id | status | when | fix |
|----|--------|------|-----|
| mongo-08 | implemented | обработка bson в пакете handler | пакет repository |

## security

| id | status | when | fix | severity |
|----|--------|------|-----|----------|
| sec-m01 | implemented | connection uri с паролем в закоммиченном файле | secret manager | fail |
| sec-m02 | implemented | remote mongo uri без tls | mongodb+srv или tls=true | fail |

см. [security-owasp.md](security-owasp.md) — sec-m01, sec-m02
