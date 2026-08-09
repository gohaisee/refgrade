package main

import (
	"fmt"
	"os"

	"github.com/gohaisee/refgrade/fixtures/good-minimal/internal/config"
	"github.com/gohaisee/refgrade/fixtures/good-minimal/internal/service"
)

func main() {
	cfg := config.LoadFromEnv(os.Getenv)
	svc := service.New(cfg.ServiceName)
	fmt.Println(cfg.Port, svc.Name())
}
