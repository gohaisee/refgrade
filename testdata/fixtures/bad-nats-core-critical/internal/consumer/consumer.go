package consumer

import "github.com/nats-io/nats.go"

func RunPayment(nc *nats.Conn) {
	nc.Subscribe("payments.critical", func(msg *nats.Msg) {})
}
