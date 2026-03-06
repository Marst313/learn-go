package routes

import (
	"net/http"

	"github.com/Marst/reminder-app/internal/utils"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine) {

	api := router.Group("/api/v1")
	{
		RegisterAuthRoutes(api)
		RegisterReminderRoutes(api)
	}

	router.NoRoute(func(c *gin.Context) {
		utils.JSONError(c, http.StatusNotFound, "Not found")
	})
}
