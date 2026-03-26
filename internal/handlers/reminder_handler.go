package handlers

import (
	"net/http"
	"strconv"

	"github.com/Marst/reminder-app/internal/models"
	"github.com/Marst/reminder-app/internal/services"
	"github.com/Marst/reminder-app/internal/utils"
	"github.com/gin-gonic/gin"
)

func GetReminders(c *gin.Context) {
	userId, exists := c.Get("user_id")
	if !exists {
		utils.JSONError(c, http.StatusNotFound, "UserID is not found!")
		return
	}

	limit := c.Query("limit")
	offset := c.Query("offset")

	if limit == "" {
		limit = "10"
	}

	if offset == "" {
		offset = "0"
	}

	limitInt, err := strconv.ParseInt(limit, 10, 32)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	offsetInt, err := strconv.ParseInt(offset, 10, 32)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	reminders, total, active, err := services.GetReminders(c.Request.Context(), userId.(int), int(limitInt), int(offsetInt))

	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Successfully get all reminders!", models.GetReminderResponse{
		Total:     total,
		Active:    active,
		Reminders: reminders,
		Results:   len(*reminders),
		Pagination: models.Page{
			Limit:  int(limitInt),
			Offset: int(offsetInt),
		},
	})
}

func NewReminder(c *gin.Context) {
	var req *models.Reminder
	userID, exists := c.Get("user_id")

	if !exists {
		utils.JSONError(c, http.StatusNotFound, "UserID is not found!")
		return
	}

	err := c.ShouldBindJSON(&req)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Failed to parse struct")
		return
	}

	if req.Title == "" || req.Date.IsZero() || req.Time.IsZero() {
		utils.JSONError(c, http.StatusBadRequest, "Title, date, and time cannot be empty!")
		return
	}

	reminder, err := services.NewReminder(c.Request.Context(), req, userID.(int))

	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, http.StatusCreated, "Successfully created new reminder!", reminder)
}

func UpdateReminder(c *gin.Context) {
	var req *models.Reminder

	id, exists := c.Params.Get("id")

	if !exists {
		utils.JSONError(c, http.StatusBadRequest, "ID is required")
		return
	}

	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "Internal server error!")
		return
	}

	err = c.ShouldBindJSON(&req)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Failed to parse struct")
		return
	}

	updatedReminder, err := services.UpdateReminder(c.Request.Context(),
		idInt, req)

	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Reminder updated successfully!", updatedReminder)
}

func DeleteReminders(c *gin.Context) {
	id, exists := c.Params.Get("id")
	if !exists {
		utils.JSONError(c, http.StatusBadRequest, "ID is required")
		return
	}

	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "Internal server error!")
		return
	}

	deletedReminder, err := services.DeleteReminder(c.Request.Context(), idInt)

	utils.JSONSuccess(c, http.StatusNoContent, "Successfully delete reminder!", deletedReminder)
}

func UpdateCompleteReminder(c *gin.Context) {
	var req models.ReminderToggleComplete
	idReminder, exists := c.Params.Get("id")

	err := c.ShouldBindJSON(&req)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Failed to parse struct")
		return
	}

	if !exists {
		utils.JSONError(c, http.StatusNotFound, "Reminder is not found")
		return
	}

	idInt, err := strconv.ParseInt(idReminder, 10, 64)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "Failed to parse int!")
		return
	}

	reminder, err := services.UpdateCompleteReminder(c.Request.Context(), idInt, req.IsCompleted)

	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "Internal server error !")
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Successfully update complated reminder!", reminder)
}
