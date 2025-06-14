package errors

import (
	"github.com/gin-gonic/gin"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/proxy/internal/models"
)

// WriteError writes an error response to the gin context
func WriteError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, models.ErrorResponse{
		Error: message,
		Code:  statusCode,
	})
}

// HandleError handles both gRPC and HTTP errors
func HandleError(c *gin.Context, err error) {
	statusCode, message := HandleGRPCError(err)
	WriteError(c, statusCode, message)
}
