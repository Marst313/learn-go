package middleware

import (
	"net/http"
	"strings"

	"github.com/Marst/reminder-app/internal/utils"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		// ! 1. GET HEADERS AUTHORIZATION
		bearerToken := c.Request.Header.Get("Authorization")

		if bearerToken == "" {
			utils.JSONError(c, http.StatusUnauthorized, "Token is missing!")
			return
		}

		// ! 2. SPLIT HEADERS BEARER
		token := (strings.Split(bearerToken, " ")[1])

		// ! 3. VALIDATE JWT
		claims, err := utils.ValidateJWT(token)

		if err != nil {
			utils.JSONError(c, http.StatusUnauthorized, err.Error())
			return
		}

		userID := int(claims["user_id"].(float64))
		email := claims["email"].(string)

		c.Set("user_id", userID)
		c.Set("email", email)

		c.Next()
	}

}
