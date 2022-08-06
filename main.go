package main

import (
	"encoding/xml"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	router.GET("/", handleName)

	router.Run()
}

func handleName(c *gin.Context) {
	type Person struct {
		XMLName   xml.Name `xml:"person"`
		FirstName string   `xml:"firstname,attr"`
		LastName  string   `xml:"lastname,attr"`
	}
	c.XML(200, Person{
		FirstName: "Mahesh",
		LastName:  "Jagtap",
	})
}
