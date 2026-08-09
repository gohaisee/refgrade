package main

import (
	"os"

	"github.com/gohaisee/refgrade/fixtures/bad-no-config/internal/worker"
)

func main() {
	_ = worker.Load()
	_ = os.Getenv("A")
}
