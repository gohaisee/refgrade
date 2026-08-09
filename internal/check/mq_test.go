package check

import (
	"context"
	"testing"
)

func TestMq01_earlyAckBeforeProcess(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "internal/consumer/consumer.go"
	path := writeGoFile(t, dir, rel, `package consumer

import amqp "github.com/rabbitmq/amqp091-go"

func Handle(d amqp.Delivery) error {
	d.Ack(false)
	process(d.Body)
	return nil
}

func process(b []byte) {}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewMq01().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "mq-01" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestMq01_ackAfterProcessOk(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "internal/consumer/consumer.go"
	path := writeGoFile(t, dir, rel, `package consumer

import amqp "github.com/rabbitmq/amqp091-go"

func Handle(d amqp.Delivery) error {
	if err := process(d.Body); err != nil {
		return err
	}
	d.Ack(false)
	return nil
}

func process(b []byte) error { return nil }
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewMq01().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestMq04_goFuncInLoop(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "internal/consumer/consumer.go"
	path := writeGoFile(t, dir, rel, `package consumer

import (
	"context"

	"github.com/segmentio/kafka-go"
)

func Run(r *kafka.Reader) {
	for {
		msg, _ := r.ReadMessage(context.Background())
		go func() { process(msg) }()
	}
}

func process(msg kafka.Message) {}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewMq04().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "mq-04" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestKafka01_readerWithoutGroup(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "internal/consumer/consumer.go"
	path := writeGoFile(t, dir, rel, `package consumer

import "github.com/segmentio/kafka-go"

func New() *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "events",
	})
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewKafka01().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "kafka-01" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestRmq01_sharedChannelGoroutines(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "internal/consumer/consumer.go"
	path := writeGoFile(t, dir, rel, `package consumer

import amqp "github.com/rabbitmq/amqp091-go"

var shared *amqp.Channel

func Run(conn *amqp.Connection) {
	shared, _ = conn.Channel()
	go worker(shared)
	go worker(shared)
}

func worker(ch *amqp.Channel) {}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewRmq01().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) < 1 || findings[0].ID != "rmq-01" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestNats01_coreSubscribeCritical(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "internal/consumer/consumer.go"
	path := writeGoFile(t, dir, rel, `package consumer

import "github.com/nats-io/nats.go"

func RunPayment(nc *nats.Conn) {
	nc.Subscribe("payments.critical", func(msg *nats.Msg) {})
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewNats01().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "nats-01" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestMq05_publishWithoutTimeout(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "internal/producer/producer.go"
	path := writeGoFile(t, dir, rel, `package producer

import (
	"context"

	"github.com/segmentio/kafka-go"
)

func Send(w *kafka.Writer, msg kafka.Message) error {
	return w.WriteMessages(context.Background(), msg)
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewMq05().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "mq-05" {
		t.Fatalf("findings = %+v", findings)
	}
}
