package utils

import (
	"sort"
	"temuin/config"
	"temuin/models"
	"time"

	"github.com/flosch/pongo2/v6"
	"github.com/gin-gonic/gin"
)

func GetGlobalContext(c *gin.Context) pongo2.Context {
    var categories []models.Category
    // Preload SubCategories
    config.DB.Preload("SubCategories").Find(&categories)

    // Sort: "Lainnya" last
    sort.Slice(categories, func(i, j int) bool {
        if categories[i].Name == "Lainnya" {
            return false
        }
        if categories[j].Name == "Lainnya" {
            return true
        }
        return categories[i].Name < categories[j].Name
    })

    // Base context
    ctx := pongo2.Context{
        "sidebar_categories": categories,
        "request":            c.Request, // For request.path checks
        "user":               nil,
        "FormatTime": func(t time.Time, layout string) string {
            return t.Format(layout)
        },
    }

    // Auth context (Check if middleware populated "user")
    if u, exists := c.Get("user"); exists {
        ctx["user"] = u
    }
    
    return ctx
}
