package handlers

import (
	"net/http"
	"temuin/config"
	"temuin/models"
	"temuin/utils"

	"github.com/flosch/pongo2/v6"
	"github.com/gin-gonic/gin"
)

func Home(c *gin.Context) {
	// Fetch items

	// Expire old highlights using utility function
	utils.ExpireHighlights(config.DB)

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

	// Fetch highlighted items (max 3)
	var highlightedItems []models.LostItem
	config.DB.Preload("User").
		Where("is_highlighted = ?", true).
		Order("highlight_expiry DESC").
		Limit(3).
		Find(&highlightedItems)

	// Check if there are more highlights
	var totalHighlights int64
	config.DB.Model(&models.LostItem{}).
		Where("is_highlighted = ?", true).
		Count(&totalHighlights)
	hasMoreHighlights := totalHighlights > 3

	ctx := utils.GetGlobalContext(c)
	ctx["items"] = allItems
	ctx["pinned_items"] = highlightedItems
	ctx["has_more_highlights"] = hasMoreHighlights
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

func LandingPage(c *gin.Context) {
	// Expire old highlights using utility function
	utils.ExpireHighlights(config.DB)

	ctx := utils.GetGlobalContext(c)

	// Fetch items from database with filtering
	var items []models.LostItem
	query := config.DB.Preload("User").Order("created_at desc")

	// Apply filters (same logic as Home handler)
	if q := c.Query("q"); q != "" {
		// Search in title OR description
		query = query.Where("(title LIKE ? OR description LIKE ?)", "%"+q+"%", "%"+q+"%")
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if loc := c.Query("location"); loc != "" {
		query = query.Where("location LIKE ?", "%"+loc+"%")
	}

	// Limit items for landing page (9 items)
	query.Limit(9).Find(&items)

	// DEBUG: Log the number of items fetched
	println("DEBUG LandingPage: Fetched", len(items), "items from database")

	// Pass data to template
	ctx["items"] = items
	ctx["q"] = c.Query("q")
	ctx["status"] = c.Query("status")
	ctx["location"] = c.Query("location")

	// Render template
	tpl, err := pongo2.FromFile("templates/core/landing.html")
	if err != nil {
		c.String(http.StatusInternalServerError, "Template Error: "+err.Error())
		return
	}
	out, err := tpl.Execute(ctx)
	if err != nil {
		c.String(http.StatusInternalServerError, "Render Error: "+err.Error())
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(out))
}
