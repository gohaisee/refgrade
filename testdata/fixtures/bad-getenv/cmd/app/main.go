package main

import (
	"fmt"
	"os"

	"github.com/gohaisee/refgrade/fixtures/bad-getenv/internal/service"
)

func main() {
	port := os.Getenv("PORT")
	svc := service.New()
	fmt.Println(port, svc.Name())
}
