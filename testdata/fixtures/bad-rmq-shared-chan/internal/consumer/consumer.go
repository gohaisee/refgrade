package consumer

import amqp "github.com/rabbitmq/amqp091-go"

var shared *amqp.Channel

func Run(conn *amqp.Connection) {
	shared, _ = conn.Channel()
	go worker(shared)
	go worker(shared)
}

func worker(ch *amqp.Channel) {}
