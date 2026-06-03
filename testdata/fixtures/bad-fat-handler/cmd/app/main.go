package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gohaisee/refgrade/fixtures/bad-fat-handler/internal/handler"
)

func main() {
	r := gin.New()
	r.GET("/", handler.Handle)
	_ = r
}
