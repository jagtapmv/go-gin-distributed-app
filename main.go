package main

import "github.com/gin-gonic/gin"

func main() {
	router := gin.Default()

	router.GET("/:name", handleName)

	router.Run()
}

func handleName(c *gin.Context) {
	userName := c.Params.ByName("name")
	c.JSON(200, gin.H{
		"Message": "Hello " + userName,
	})
}
