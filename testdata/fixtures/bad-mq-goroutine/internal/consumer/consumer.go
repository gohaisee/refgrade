package consumer

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
