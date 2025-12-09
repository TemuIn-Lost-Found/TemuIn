package handlers

import (
	"net/http"
	"temuin/config"
	"temuin/models"
	"temuin/utils"
	"time"

	"github.com/flosch/pongo2/v6"
	"github.com/gin-gonic/gin"
)

func Home(c *gin.Context) {
    // Fetch items

    
    // Check and expire highlights
    // In a real app, this should be a background cron job.
    // For this prototype, we lazy check on Home load for highlighted items.
    var highlightedItems []models.LostItem
    if err := config.DB.Where("is_highlighted = ?", true).Find(&highlightedItems).Error; err == nil {
        now := time.Now()
        for _, item := range highlightedItems {
            if item.HighlightExpiry != nil && now.After(*item.HighlightExpiry) {
                item.IsHighlighted = false
                config.DB.Save(&item)
            }
        }
    }

    var allItems []models.LostItem
    query := config.DB.Preload("User").Order("created_at desc")
    
    // Filtering
    if q := c.Query("q"); q != "" {
        // Search title OR description
        query = query.Where("(title LIKE ? OR description LIKE ?)", "%"+q+"%", "%"+q+"%")
    }
    if status := c.Query("status"); status != "" {
        query = query.Where("status = ?", status)
    }
    if loc := c.Query("location"); loc != "" {
        query = query.Where("location LIKE ?", "%"+loc+"%")
    }
    
    query.Find(&allItems)

    ctx := utils.GetGlobalContext(c)
    ctx["items"] = allItems
    ctx["q"] = c.Query("q")
    ctx["status"] = c.Query("status")
    ctx["location"] = c.Query("location")
    
    // User check for UI
    if u, exists := c.Get("user"); exists {
         ctx["user"] = u
    } else {
        // Try to get from session if not in context (for public home)
        // ... (Middleware might not verify public routes)
    }

    // Render template
    tpl, err := pongo2.FromFile("templates/core/home.html")
    if err != nil {
        println("Pongo2 Load Error:", err.Error())
        c.String(http.StatusInternalServerError, "Pongo2 Load Error: "+err.Error())
        return
    }
    out, err := tpl.Execute(ctx)
    if err != nil {
        // Log the actual error
        println("Template Error:", err.Error())
        c.String(http.StatusInternalServerError, "Template Error: "+err.Error())
        return
    }
    c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(out))
}
