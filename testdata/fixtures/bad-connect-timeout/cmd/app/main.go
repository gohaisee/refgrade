package main

import (
	"github.com/gohaisee/refgrade/fixtures/bad-connect-timeout/internal/client"
)

func main() {
	_ = client.NewGreeterClient()
}
