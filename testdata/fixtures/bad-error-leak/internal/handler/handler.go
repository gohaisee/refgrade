package handler

import "github.com/gin-gonic/gin"

func Handle(c *gin.Context, err error) {
	c.JSON(500, gin.H{"error": err.Error()})
}
