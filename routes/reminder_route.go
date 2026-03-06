package routes

import (
	"github.com/Marst/reminder-app/internal/handlers"
	"github.com/Marst/reminder-app/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterReminderRoutes(api *gin.RouterGroup) {
	reminderRouter := api.Group("/reminders")

	reminderRouter.Use(middleware.AuthMiddleware())
	{
		reminderRouter.POST("", handlers.NewReminder)
		reminderRouter.GET("", handlers.GetReminders)
		reminderRouter.PATCH("/:id", handlers.UpdateReminder)
		reminderRouter.DELETE("/:id", handlers.DeleteReminders)
		reminderRouter.PATCH("/:id/toggle-complete", handlers.UpdateCompleteReminder)

	}
}
