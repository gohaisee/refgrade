# messaging — kafka, rabbitmq, nats

**n/a** when no message broker client in module

## detected imports

| broker | import |
|--------|--------|
| kafka | `github.com/segmentio/kafka-go` |
| rabbitmq | `github.com/rabbitmq/amqp091-go` |
| nats | `github.com/nats-io/nats.go` |

## universal (all brokers)

| id | when | why | fix | severity |
|----|------|-----|-----|----------|
| mq-01 | consumer ack before successful process | message loss on crash | ack after success | fail |
| mq-02 | no idempotency on consumer handler | duplicate delivery breaks state | idempotent handler + dedup key | warn |
| mq-03 | no reconnect/backoff config | stall on network blip | library reconnect options | warn |
| mq-04 | unbounded goroutine per message | memory spike | worker pool with limit | warn |
| mq-05 | publish without context deadline | hang on broker down | ctx with timeout | warn |

## kafka (segmentio/kafka-go)

| id | when | fix |
|----|------|-----|
| kafka-01 | consumer without group id | set `GroupID` |
| kafka-02 | commit before process (overlap mq-01) | commit after |
| kafka-03 | no partition key when order required | key by entity id |

## rabbitmq (amqp091-go)

| id | when | fix |
|----|------|-----|
| rmq-01 | shared channel across goroutines | channel per consumer or mutex |
| rmq-02 | no dead-letter exchange on critical queues | configure dlx |
| rmq-03 | connection drop without recreate loop | reconnect wrapper |

## nats

| id | when | fix |
|----|------|-----|
| nats-01 | core nats for must-not-lose work | jetstream with ack |
| nats-02 | jetstream consumer without ack/nak handling | explicit ack policy |

## design note

sync rpc (grpc/http) and async queue solve different problems — flag queue used where caller waits for peer data that could be local (info only, not auto-fail)

## security (catalog only — deferred in v1.0.0)

| id | status | when | fix |
|----|--------|------|-----|
| sec-mq01 | deferred (overlap [cfg-03](universal.md#config)) | broker url with credentials in repo | secrets via env |
| sec-mq02 | deferred | plaintext amqp/nats to public internet | tls |
