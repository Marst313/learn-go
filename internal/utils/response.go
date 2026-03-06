package utils

import (
	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	ErrorCode int    `json:"errorCode"`
}

type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func JSONError(c *gin.Context, status int, err string) {
	c.AbortWithStatusJSON(status, &ErrorResponse{
		Success:   false,
		Message:   err,
		ErrorCode: status,
	})
}

func JSONSuccess(c *gin.Context, status int, message string, data any) {
	c.JSON(status, &SuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}
