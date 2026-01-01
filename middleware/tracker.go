package middleware

import (
	"temuin/config"
	"temuin/models"

	"github.com/gin-gonic/gin"
)

// VisitorTracker tracks all page visits
func VisitorTracker() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Record the visit
		visit := models.SiteVisit{
			IPAddress: c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
			Path:      c.Request.URL.Path,
		}

		// Save asynchronously to not block requests
		go func() {
			config.DB.Create(&visit)
		}()

		c.Next()
	}
}
