package handlers

import (
	"net/http"

	"github.com/Marst/reminder-app/internal/models"
	"github.com/Marst/reminder-app/internal/services"
	"github.com/Marst/reminder-app/internal/utils"
	"github.com/gin-gonic/gin"
)

func Register(c *gin.Context) {
	var req models.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Failed to parse struct")
		return
	}

	user, err := services.Register(&req)

	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSONSuccess(c, http.StatusCreated, "Successfully register new user!", user)
}

func Login(c *gin.Context) {
	var req models.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Failed to parse struct")
		return
	}

	resp, err := services.Login(&req)

	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Successfully login!", resp)
}

func Profile(c *gin.Context) {
	userID, exists := c.Get("user_id")

	if !exists {
		utils.JSONError(c, http.StatusBadRequest, "Token is required")
		return
	}

	user, err := services.Profile(c.Request.Context(), userID.(int))

	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Profile successfully fetched!", user)
}

func RefreshCookies(c *gin.Context) {
	userID, exists := c.Get("user_id")

	if !exists {
		utils.JSONError(c, http.StatusBadRequest, "Token is required")
		return
	}

	user, err := services.RefreshCookies(c.Request.Context(), userID.(int))

	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Refresh login are successfully!", user)
}

func Logout(c *gin.Context) {
	data := ""
	utils.JSONSuccess(c, http.StatusOK, "Successfully logged out!", data)
}
