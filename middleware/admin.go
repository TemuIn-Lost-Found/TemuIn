package middleware

import (
	"net/http"
	"temuin/models"

	"github.com/gin-gonic/gin"
)

// AdminRequired checks if the user is an admin (IsSuperuser = true)
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		userInterface, exists := c.Get("user")
		if !exists {
			c.String(http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}

		user := userInterface.(*models.User)

		if !user.IsSuperuser {
			c.String(http.StatusForbidden, "Admin access required")
			c.Abort()
			return
		}

		c.Next()
	}
}
