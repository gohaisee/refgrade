package main

import (
	"net/http"

	"github.com/gohaisee/refgrade/fixtures/bad-mass-assign/internal/handler"
)

func main() {
	http.HandleFunc("/users", handler.Update)
}
