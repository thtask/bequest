package app

import (
	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (app *Application) errorResponse(c *gin.Context, status int, message string) {
	c.JSON(status, &Response{
		Success: false,
		Message: message,
	})
}

func (app *Application) successResponse(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, &Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}
