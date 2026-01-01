package middleware

import (
	"fmt"
	"net/http"
	"temuin/config"
	"temuin/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("AuthRequired Middleware Triggered")
		session := sessions.Default(c)
		userID := session.Get("user_id")
		fmt.Printf("Session UserID: %v\n", userID)

		if userID == nil {
			c.Redirect(http.StatusFound, "/")
			c.Abort()
			return
		}

		// Fetch user from DB and set in context
		var user models.User
		if err := config.DB.First(&user, userID).Error; err != nil {
			session.Delete("user_id")
			session.Save()
			c.Redirect(http.StatusFound, "/")
			c.Abort()
			return
		}

		// Check if user is banned
		if user.IsBanned {
			session.Delete("user_id")
			session.Save()
			c.Redirect(http.StatusFound, "/login?error=banned")
			c.Abort()
			return
		}

		c.Set("user", &user)
		c.Next()
	}
}

func AuthOptional() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user_id")

		if userID != nil {
			var user models.User
			if err := config.DB.First(&user, userID).Error; err == nil {
				c.Set("user", &user)
			}
		}
		c.Next()
	}
}
