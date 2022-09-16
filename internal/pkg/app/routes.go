package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (a *Application) Routes() http.Handler {
	e := gin.Default()
	e.Use(gin.Recovery())

	v1 := e.Group("/api/v1")
	{
		v1.POST("/answers", a.CreateAnswer)
		v1.GET("/answers/:key", a.FindAnswerByKey)
		v1.PUT("/answers/:key", a.UpdateAnswer)
		v1.DELETE("/answers/:key", a.DeleteAnswer)
		v1.GET("/answers/:key/history", a.FindHistoryByKey)
	}

	return e
}
