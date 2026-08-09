# messaging — kafka, rabbitmq, nats

**n/a** когда в модуле нет клиента message broker

## detected imports

| broker | import |
|--------|--------|
| kafka | `github.com/segmentio/kafka-go` |
| rabbitmq | `github.com/rabbitmq/amqp091-go` |
| nats | `github.com/nats-io/nats.go` |

## universal (all brokers)

| id | when | why | fix | severity |
|----|------|-----|-----|----------|
| mq-01 | consumer ack до успешной обработки | потеря сообщений при падении | ack после успеха | fail |
| mq-02 | нет идемпотентности в consumer handler | дубликаты ломают состояние | идемпотентный handler + dedup key | warn |
| mq-03 | нет reconnect/backoff config | зависание при сетевом сбое | опции reconnect библиотеки | warn |
| mq-04 | неограниченная горутина на сообщение | скачок памяти | пул воркеров с лимитом | warn |
| mq-05 | publish без deadline в context | зависание при недоступном broker | ctx с таймаутом | warn |

## kafka (segmentio/kafka-go)

| id | when | fix |
|----|------|-----|
| kafka-01 | consumer без group id | задать `GroupID` |
| kafka-02 | commit до обработки (пересечение с mq-01) | commit после |
| kafka-03 | нет partition key когда нужен порядок | ключ по id entity |

## rabbitmq (amqp091-go)

| id | when | fix |
|----|------|-----|
| rmq-01 | общий channel между горутинами | channel на consumer или mutex |
| rmq-02 | нет dead-letter exchange на критичных очередях | настроить dlx |
| rmq-03 | обрыв соединения без цикла recreate | обёртка reconnect |

## nats

| id | when | fix |
|----|------|-----|
| nats-01 | core nats для работы, которую нельзя терять | jetstream с ack |
| nats-02 | jetstream consumer без ack/nak handling | явная политика ack |

## design note

синхронный rpc (grpc/http) и async queue решают разные задачи — флаг, если очередь используется там, где вызывающий ждёт данные peer, которые могли бы быть локальными (только info, не auto-fail)

## security (только каталог — deferred в v1.0.0)

| id | статус | when | fix |
|----|--------|------|-----|
| sec-mq01 | deferred (overlap [cfg-03](universal.md#config)) | url broker с credentials в репозитории | секреты через env |
| sec-mq02 | deferred | plaintext amqp/nats в публичный интернет | tls |
