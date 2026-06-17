package consumer

import amqp "github.com/rabbitmq/amqp091-go"

func Handle(d amqp.Delivery) error {
	d.Ack(false)
	process(d.Body)
	return nil
}

func process(b []byte) {}
