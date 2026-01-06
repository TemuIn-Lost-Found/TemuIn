package middleware

import (
	"temuin/config"
	"temuin/models"

	"github.com/gin-gonic/gin"
)

// VisitorTracker tracks all page visits
// VisitorTracker tracks unique daily visits (cookie-based) and excludes admin
func VisitorTracker() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Exclude Admin Routes & Static files
		path := c.Request.URL.Path
		if len(path) >= 6 && path[:6] == "/admin" {
			c.Next()
			return
		}
		if len(path) >= 7 && path[:7] == "/static" {
			c.Next()
			return
		}

		// 2. Check for "temuin_visited" cookie
		_, err := c.Cookie("temuin_visited")
		if err != nil {
			// Cookie not found -> New Unique Visit (for today)
			
			// Record the visit
			visit := models.SiteVisit{
				IPAddress: c.ClientIP(),
				UserAgent: c.Request.UserAgent(),
				Path:      path, // Optional: might just want generic "visit" or keep path
			}

			// Save asynchronously
			go func() {
				config.DB.Create(&visit)
			}()

			// Set cookie for 24 hours (86400 seconds)
			// Name, Value, MaxAge, Path, Domain, Secure, HttpOnly
			c.SetCookie("temuin_visited", "true", 86400, "/", "", false, true)
		}

		c.Next()
	}
}
