package middlewares

import (
	"net/http"
	"strings"

	"gin-freemarket/services"

	"github.com/gin-gonic/gin"
)

// Auth Middleware
// check if the user is authenticated based on JWT on Authorization header.
func AuthMiddleware(authService services.IAuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		if !strings.HasPrefix(token, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		token = strings.TrimPrefix(token, "Bearer ")

		user, err := authService.GetUserFromToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		// set user to context for later use.
		c.Set("user", user)
		c.Set("token", token)
		c.Next()
	}
}
