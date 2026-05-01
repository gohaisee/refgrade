# mongodb

**n/a** когда `go.mongodb.org/mongo-driver` нет в модуле

## detected import

`go.mongodb.org/mongo-driver/mongo`

## checks

| id | when | why | fix | severity |
|----|------|-----|-----|----------|
| mongo-01 | `mongo.Connect` на каждый http/grpc запрос | исчерпание сокетов | один `mongo.Client` на процесс | fail |
| mongo-02 | вызов бд с `context.Background()` на пути запроса | нет отмены при disconnect | передавать ctx запроса | warn |
| mongo-03 | `Find` без limit в list handlers | неограниченная память | `SetLimit` | warn |
| mongo-04 | нет `Ping` после connect в main | тихая misconfig | ping с таймаутом при старте | info |
| mongo-05 | опции pool не заданы на высоконагруженном сервисе | латентность под нагрузкой | `SetMaxPoolSize`, `SetMinPoolSize` | info |
| mongo-06 | `$where` / js eval с пользовательским вводом | инъекция | избегать `$where`; использовать операторы | fail |
| mongo-07 | пользовательский `$regex` без escape | re dos / инъекция | якорить и экранировать паттерн | warn |

## layout

| id | when | fix |
|----|------|-----|
| mongo-08 | обработка bson в пакете handler | пакет repository |

## security

| id | when | fix |
|----|------|-----|
| sec-m01 | connection uri с паролем в закоммиченном файле | secret manager |
| sec-m02 | tls отключён для удалённого кластера | включить tls в uri |
