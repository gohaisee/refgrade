package main

import "github.com/gin-gonic/gin"

func main() {
	r := gin.New()
	r.POST("/upload", func(c *gin.Context) {})
	_ = r
}
