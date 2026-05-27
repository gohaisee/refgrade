package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gohaisee/refgrade/fixtures/bad-handler-sql/internal/handler"
)

func main() {
	r := gin.New()
	h := handler.New(nil)
	r.GET("/", h.List)
	_ = r
}
