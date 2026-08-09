# messaging — kafka, rabbitmq, nats

**n/a** когда в модуле нет клиента message broker

## detected imports

| broker | import |
|--------|--------|
| kafka | `github.com/segmentio/kafka-go` |
| rabbitmq | `github.com/rabbitmq/amqp091-go` |
| nats | `github.com/nats-io/nats.go` |

## universal (all brokers)

| id | status | when | why | fix | severity |
|----|--------|------|-----|-----|----------|
| mq-01 | implemented | consumer ack до успешной обработки | потеря сообщений при падении | ack после успеха | fail |
| mq-02 | implemented | нет идемпотентности в consumer handler | дубликаты ломают состояние | идемпотентный handler + dedup key | warn |
| mq-03 | implemented | нет reconnect/backoff config | зависание при сетевом сбое | опции reconnect библиотеки | warn |
| mq-04 | implemented | неограниченная горутина на сообщение | скачок памяти | пул воркеров с лимитом | warn |
| mq-05 | implemented | publish без deadline в context | зависание при недоступном broker | ctx с таймаутом | warn |

## kafka (segmentio/kafka-go)

| id | status | when | fix |
|----|--------|------|-----|
| kafka-01 | implemented | consumer без group id | задать `GroupID` |
| kafka-02 | overlap → mq-01 | kafka commit до успешной обработки | commit после успешной обработки |
| kafka-03 | implemented | нет partition key когда нужен порядок | ключ по id entity |

## rabbitmq (amqp091-go)

| id | status | when | fix |
|----|--------|------|-----|
| rmq-01 | implemented | общий channel между горутинами | channel на consumer или mutex |
| rmq-02 | implemented | нет dead-letter exchange на критичных очередях | настроить dlx |
| rmq-03 | implemented | обрыв соединения без цикла recreate | обёртка reconnect |

## nats

| id | status | when | fix |
|----|--------|------|-----|
| nats-01 | implemented | core nats для работы, которую нельзя терять | jetstream с ack |
| nats-02 | implemented | jetstream consumer без ack/nak handling | явная политика ack |

## design note

синхронный rpc (grpc/http) и async queue решают разные задачи — флаг, если очередь используется там, где вызывающий ждёт данные peer, которые могли бы быть локальными (только info, не auto-fail)

## security

| id | status | when | fix | severity |
|----|--------|------|-----|----------|
| sec-mq01 | implemented | url broker с credentials в репозитории | секреты через env | fail |
| sec-mq02 | implemented | plaintext amqp/nats на remote broker | amqps или tls dial | fail |

см. [security-owasp.md](security-owasp.md) — sec-mq01, sec-mq02
