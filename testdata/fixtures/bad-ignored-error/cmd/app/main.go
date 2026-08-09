package main

import (
	"fmt"

	"github.com/gohaisee/refgrade/fixtures/bad-ignored-error/internal/service"
)

func main() {
	svc := service.New()
	fmt.Println(svc.Name())
}
