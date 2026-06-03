package main

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/gohaisee/refgrade/fixtures/bad-error-leak/internal/handler"
)

func main() {
	r := gin.New()
	r.GET("/", func(c *gin.Context) {
		handler.Handle(c, errors.New("db connection failed"))
	})
	_ = r
}
