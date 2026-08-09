package consumer

import "github.com/segmentio/kafka-go"

func New() *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "events",
	})
}
