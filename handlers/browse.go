package handlers

import (
	"net/http"
	"temuin/config"
	"temuin/models"
	"temuin/utils"

	"github.com/flosch/pongo2/v6"
	"github.com/gin-gonic/gin"
)

func CategoryPage(c *gin.Context) {
    id := c.Param("pk")
    ctx := utils.GetGlobalContext(c)
    
    var category models.Category
    if err := config.DB.Preload("SubCategories").First(&category, id).Error; err != nil {
         c.String(http.StatusNotFound, "Category not found")
         return
    }
    
    // items
    var items []models.LostItem
    config.DB.Preload("User").Where("category_id = ?", id).Find(&items)
    
    ctx["active_category"] = category
    ctx["items"] = items
    ctx["header_title"] = category.Name
    
    tpl := pongo2.Must(pongo2.FromFile("templates/core/home.html")) // Reuse home template? Or landing?
    // Originally loop uses 'items', so home.html works if we pass items.
    out, _ := tpl.Execute(ctx)
    c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(out))
}

func SubCategoryPage(c *gin.Context) {
    id := c.Param("pk")
    ctx := utils.GetGlobalContext(c)
    
    var sub models.SubCategory
    if err := config.DB.First(&sub, id).Error; err != nil {
         c.String(http.StatusNotFound, "SubCategory not found")
         return
    }
    
    // Fetch parent for active state
    var category models.Category
    config.DB.Preload("SubCategories").First(&category, sub.CategoryID)

    // items
    var items []models.LostItem
    var pinnedItems []models.LostItem
    var normalItems []models.LostItem
    
    // Fetch all for this subcat
    config.DB.Preload("User").Where("subcategory_id = ?", id).Find(&items)
    
    for _, item := range items {
         if item.IsHighlighted {
             pinnedItems = append(pinnedItems, item)
         } else {
             normalItems = append(normalItems, item)
         }
    }
    
    ctx["active_category"] = category
    ctx["active_subcategory"] = sub
    ctx["items"] = normalItems
    ctx["pinned_items"] = pinnedItems
    ctx["header_title"] = sub.Name
    
    tpl := pongo2.Must(pongo2.FromFile("templates/core/home.html"))
    out, _ := tpl.Execute(ctx)
    c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(out))
}
