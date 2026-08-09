package handler

import (
	"net/http"

	redis "github.com/redis/go-redis/v9"
)

func Cache(w http.ResponseWriter, r *http.Request) {
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	_ = client
}
