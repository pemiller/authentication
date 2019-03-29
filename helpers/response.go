package helpers

import (
	"github.com/gin-gonic/gin"
)

// PrepareErrorResponse places the error message into a JSON interface
func PrepareErrorResponse(message string, err error) gin.H {
	resp := gin.H{
		"message": message,
	}

	if err != nil {
		resp["error"] = err
	}

	return resp
}
