package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// gin router
func main() {
	r := gin.Default()
	r.GET("/hello", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello World!",
		})
	})
	log.Fatal(http.ListenAndServe(":8080", r))
}
