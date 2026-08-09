package client

import (
	"net/http"

	"connectrpc.com/connect"
)

func NewGreeterClient() *connect.Client[struct{}, struct{}] {
	return connect.NewClient[struct{}, struct{}](http.DefaultClient, "http://localhost:8080")
}
