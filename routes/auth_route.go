package routes

import (
	"github.com/Marst/reminder-app/internal/handlers"
	"github.com/Marst/reminder-app/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(api *gin.RouterGroup) {
	authRouter := api.Group("/auth")
	{
		authRouter.POST("/register", handlers.Register)
		authRouter.POST("/login", handlers.Login)
		authRouter.POST("/logout", handlers.Logout)

		authRouter.Use(middleware.AuthMiddleware())
		authRouter.GET("/refresh-cookies", handlers.RefreshCookies)
		authRouter.GET("/profile", handlers.Profile)
	}
}
