package utils

import (
	structs "kapacitor-alerts-api/structs"
	"log"

	"github.com/gin-gonic/gin"
)

// ReportError - Send an error message to the client
func ReportError(err error, c *gin.Context, msg string) {
	log.Println(err)
	var er structs.ErrorResponse
	if msg != "" {
		er.Error = msg
	} else if err != nil {
		er.Error = err.Error()
	} else {
		er.Error = "Internal server error"
	}
	c.JSON(500, er)
}

// ReportInvalidRequest - Send a 400 bad request message to the client
func ReportInvalidRequest(c *gin.Context, msg string) {
	var er structs.ErrorResponse
	er.Error = msg
	c.JSON(400, er)
}
