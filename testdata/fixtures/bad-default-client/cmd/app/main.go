package main

import (
	"fmt"
	"net/http"

	"github.com/gohaisee/refgrade/fixtures/bad-default-client/internal/service"
)

func main() {
	svc := service.New()
	resp, err := http.Get("http://example.com")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	fmt.Println(svc.Name())
}
