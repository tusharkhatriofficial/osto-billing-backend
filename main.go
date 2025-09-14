package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	//helthcheck endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"Status":  "ok",
			"Message": "osto billing backend is running",
		})
	})

	r.Run(":8080")
}
