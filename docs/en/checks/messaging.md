# messaging — kafka, rabbitmq, nats

**n/a** when no message broker client in module

## detected imports

| broker | import |
|--------|--------|
| kafka | `github.com/segmentio/kafka-go` |
| rabbitmq | `github.com/rabbitmq/amqp091-go` |
| nats | `github.com/nats-io/nats.go` |

## universal (all brokers)

| id | status | when | why | fix | severity |
|----|--------|------|-----|-----|----------|
| mq-01 | implemented | consumer ack before successful process | message loss on crash | ack after success | fail |
| mq-02 | implemented | no idempotency on consumer handler | duplicate delivery breaks state | idempotent handler + dedup key | warn |
| mq-03 | implemented | no reconnect/backoff config | stall on network blip | library reconnect options | warn |
| mq-04 | implemented | unbounded goroutine per message | memory spike | worker pool with limit | warn |
| mq-05 | implemented | publish without context deadline | hang on broker down | ctx with timeout | warn |

## kafka (segmentio/kafka-go)

| id | status | when | fix |
|----|--------|------|-----|
| kafka-01 | implemented | consumer without group id | set `GroupID` |
| kafka-02 | overlap → mq-01 | kafka commit before successful process | commit after successful processing |
| kafka-03 | implemented | no partition key when order required | key by entity id |

## rabbitmq (amqp091-go)

| id | status | when | fix |
|----|--------|------|-----|
| rmq-01 | implemented | shared channel across goroutines | channel per consumer or mutex |
| rmq-02 | implemented | no dead-letter exchange on critical queues | configure dlx |
| rmq-03 | implemented | connection drop without recreate loop | reconnect wrapper |

## nats

| id | status | when | fix |
|----|--------|------|-----|
| nats-01 | implemented | core nats for must-not-lose work | jetstream with ack |
| nats-02 | implemented | jetstream consumer without ack/nak handling | explicit ack policy |

## design note

sync rpc (grpc/http) and async queue solve different problems — flag queue used where caller waits for peer data that could be local (info only, not auto-fail)

## security

| id | status | when | fix | severity |
|----|--------|------|-----|----------|
| sec-mq01 | implemented | broker url with credentials in repo | secrets via env | fail |
| sec-mq02 | implemented | plaintext amqp/nats to remote broker | amqps or tls dial | fail |

see [security-owasp.md](security-owasp.md) — sec-mq01, sec-mq02
