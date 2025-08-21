package middlewares

import (
	"log"
	"net/http"

	"gin-freemarket/services"
	"gin-freemarket/utils/sessions"

	"github.com/gin-gonic/gin"
)

func SessionMiddleware(authService services.IAuthService) gin.HandlerFunc {
	sessionManager, err := sessions.GetSessionManager()
	if err != nil {
		log.Fatal(err)
	}
	return func(c *gin.Context) {
		token, exists := c.Get("token")
		if !exists || token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Issue on getting token"})
			return
		}

		exists, err := sessionManager.SessionExists(token.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Issue on checking session"})
			return
		}

		if !exists {
			// need to register new session
			ok, err := sessionManager.RegisterSession(token.(string))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Issue on registering session"})
				return
			}
			if !ok {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Session limit reached"})
				return
			}
		}

		c.Next()
	}
}
