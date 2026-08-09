package main

import "github.com/gin-gonic/gin"

func main() {
	cfg := struct {
		AllowOrigins     []string
		AllowCredentials bool
	}{AllowOrigins: []string{"*"}, AllowCredentials: true}
	_ = cfg
	_ = gin.New()
}
