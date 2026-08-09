package main

import "github.com/gin-gonic/gin"

func main() {
	r := gin.New()
	r.GET("/admin/users", func(c *gin.Context) {})
	r.Run(":8080")
}
