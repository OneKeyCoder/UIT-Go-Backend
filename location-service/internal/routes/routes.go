package routes

import (
	"location-service/internal/handlers"

	"github.com/gin-gonic/gin"
)

func InitRoute(router *gin.Engine, locationHandlers *handlers.Handlers) {
	router.POST("/", func(c *gin.Context) {
		locationHandlers.SetCurrentLocation(c.Writer, c.Request)
	})

	router.GET("/", func(c *gin.Context) {
		locationHandlers.GetCurrentLocation(c.Writer, c.Request)
	})

	router.GET("/nearest", func(c *gin.Context) {
		locationHandlers.FindNearestUsers(c.Writer, c.Request)
	})
}
