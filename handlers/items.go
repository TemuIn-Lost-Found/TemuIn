package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"temuin/config"
	"temuin/models"
	"temuin/utils"
	"time"

	"github.com/flosch/pongo2/v6"
	"github.com/gin-gonic/gin"
)

func ReportItemPage(c *gin.Context) {
	ctx := utils.GetGlobalContext(c)

	// Fetch categories for dropdown
	var categories []models.SubCategory
	config.DB.Find(&categories)
	ctx["subcategories"] = categories

	tpl, err := pongo2.FromFile("templates/core/report_item.html")
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

func ReportItem(c *gin.Context) {
	user := c.MustGet("user").(*models.User)

	// Check if user is banned
	if user.IsBanned {
		c.String(http.StatusForbidden, "Your account has been banned. You cannot create posts.")
		return
	}

	title := c.PostForm("title")
	desc := c.PostForm("description")
	location := c.PostForm("location")
	bountyStr := c.PostForm("bounty_coins")
	subCatIDStr := c.PostForm("subcategory") // Assuming form sends ID

	bounty, _ := strconv.Atoi(bountyStr)
	subCatID, _ := strconv.ParseInt(subCatIDStr, 10, 64)

	// Handle Image Upload
	file, err := c.FormFile("image")
	var imagePath string
	if err == nil {
		filename := fmt.Sprintf("%d_%s", time.Now().Unix(), filepath.Base(file.Filename))
		// Save to static/images for simplicity in Go app structure
		dst := "static/images/" + filename
		if err := c.SaveUploadedFile(file, dst); err == nil {
			imagePath = "images/" + filename
		}
	}

	// Logic: Check coin balance
	if bounty > 0 {
		if user.CoinBalance < bounty {
			c.String(http.StatusBadRequest, "Insufficient coins")
			return
		}
		user.CoinBalance -= bounty
		config.DB.Save(user)
	}

	// Fetch Category ID from SubCategory
	var subCat models.SubCategory
	var catID int64
	if config.DB.First(&subCat, subCatID).Error == nil {
		catID = subCat.CategoryID
	}

	item := models.LostItem{
		Title:         title,
		Description:   desc,
		Location:      location,
		BountyCoins:   bounty,
		UserID:        user.ID,
		Image:         imagePath,
		Status:        "LOST",
		SubCategoryID: &subCatID,
		CategoryID:    catID,
	}

	config.DB.Create(&item)
	c.Redirect(http.StatusFound, "/")
}

func ItemDetail(c *gin.Context) {
	id := c.Param("pk")
	var item models.LostItem

	if err := config.DB.Preload("User").Preload("Finder").Preload("Comments").Preload("Comments.User").First(&item, id).Error; err != nil {
		c.String(http.StatusNotFound, "Item not found")
		return
	}

	ctx := utils.GetGlobalContext(c)
	ctx["item"] = item
	ctx["comments"] = item.Comments

	// User for template logic
	if u, exists := c.Get("user"); exists {
		ctx["user"] = u
		// Check if user is owner
		user := u.(*models.User)
		ctx["is_owner"] = (item.UserID == user.ID)
		ctx["is_finder"] = (item.FinderID != nil && *item.FinderID == user.ID)

		// Logic: If Owner, fetch claims
		if item.UserID == user.ID {
			var claims []models.ItemClaim
			config.DB.Preload("User").Where("item_id = ?", item.ID).Find(&claims)
			ctx["claims"] = claims
		}

		// Logic: Check if current user has claimed
		var count int64
		config.DB.Model(&models.ItemClaim{}).Where("item_id = ? AND user_id = ?", item.ID, user.ID).Count(&count)
		ctx["has_claimed"] = (count > 0)
	}

	tpl, err := pongo2.FromFile("templates/core/item_detail.html")
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

func PostComment(c *gin.Context) {
	itemID := c.Param("pk")
	content := c.PostForm("content")
	user := c.MustGet("user").(*models.User)

	// Check if user is banned
	if user.IsBanned {
		c.String(http.StatusForbidden, "Your account has been banned. You cannot comment.")
		return
	}

	iid, _ := strconv.ParseInt(itemID, 10, 64)

	comment := models.Comment{
		Content: content,
		UserID:  user.ID,
		ItemID:  iid,
	}

	config.DB.Create(&comment)
	c.Redirect(http.StatusFound, "/item/"+itemID)
}

// MarkAsFound handles logic when someone claims they found it or owner marks it found
func MarkAsFound(c *gin.Context) {
	itemID := c.Param("pk")
	user := c.MustGet("user").(*models.User)

	var item models.LostItem
	if err := config.DB.First(&item, itemID).Error; err != nil {
		c.String(http.StatusNotFound, "Item not found")
		return
	}

	// Logic: User claims they found it.
	// Creates an ItemClaim record.
	if item.Status == "LOST" {
		// Prevent duplicate claims
		var count int64
		config.DB.Model(&models.ItemClaim{}).Where("item_id = ? AND user_id = ?", item.ID, user.ID).Count(&count)
		if count == 0 {
			claim := models.ItemClaim{
				ItemID: item.ID,
				UserID: user.ID,
			}
			config.DB.Create(&claim)
		}
	}

	c.Redirect(http.StatusFound, "/item/"+itemID)
}

func SelectFinder(c *gin.Context) {
	itemID := c.Param("pk")
	targetUserID := c.PostForm("candidate_id")
	user := c.MustGet("user").(*models.User)

	var item models.LostItem
	if err := config.DB.First(&item, itemID).Error; err != nil {
		c.String(http.StatusNotFound, "Item not found")
		return
	}

	// Only Owner can select
	if item.UserID != user.ID {
		c.String(http.StatusForbidden, "Not authorized")
		return
	}

	uid, _ := strconv.ParseInt(targetUserID, 10, 64)
	item.FinderID = &uid
	item.FinderConfirmed = false // Reset confirmation to force mutual check
	item.OwnerConfirmed = false
	config.DB.Save(&item)

	c.Redirect(http.StatusFound, "/item/"+itemID)
}

// ConfirmReturn handles mutual confirmation. Both Owner and Finder must confirm.
func ConfirmReturn(c *gin.Context) {
	itemID := c.Param("pk")
	user := c.MustGet("user").(*models.User)

	var item models.LostItem
	if err := config.DB.Preload("Finder").First(&item, itemID).Error; err != nil {
		c.String(http.StatusNotFound, "Item not found")
		return
	}

	// Check authorization (Owner or Finder)
	isOwner := (item.UserID == user.ID)
	isFinder := (item.FinderID != nil && *item.FinderID == user.ID)

	if !isOwner && !isFinder {
		c.String(http.StatusForbidden, "Not authorized")
		return
	}

	// Update confirmation status
	if isOwner {
		item.OwnerConfirmed = true
		// User Logic: "Harusnya penemu tidak perlu konfirmasi lagi ketika pemilik sudah konfirmasi"
		// Meaning: Owner's confirmation is the Final Say (since they pay).
		// So if Owner accepts, we auto-confirm the Finder side (assuming they gave it).
		item.FinderConfirmed = true
	} else if isFinder {
		item.FinderConfirmed = true
	}
	config.DB.Save(&item)

	// Check if BOTH confirmed (Now Owner creates this state immediately)
	// So, if Owner is confirming now, AND FinderConfirmed is true -> DEAL DONE.
	if item.FinderConfirmed && item.OwnerConfirmed && item.Status == "LOST" {
		item.Status = "FOUND"      // Terminal state
		item.IsHighlighted = false // Remove highlight
		config.DB.Save(&item)

		// Transfer Bounty
		if item.BountyCoins > 0 && item.Finder != nil {
			item.Finder.CoinBalance += item.BountyCoins
			config.DB.Save(item.Finder)
		}
	}

	c.Redirect(http.StatusFound, "/item/"+itemID)
}

func HighlightItem(c *gin.Context) {
	itemID := c.Param("pk")
	user := c.MustGet("user").(*models.User)

	var item models.LostItem
	if err := config.DB.First(&item, itemID).Error; err != nil {
		c.String(http.StatusNotFound, "Item not found")
		return
	}

	// Check ownership
	if item.UserID != user.ID {
		c.String(http.StatusForbidden, "Not authorized")
		return
	}

	// Cost: 50 coins
	cost := 50
	if user.CoinBalance < cost {
		// Should show error, but for now redirect with internal error or just ignore
		c.Redirect(http.StatusFound, "/profile")
		return
	}

	user.CoinBalance -= cost
	item.IsHighlighted = true

	// Expiry: 24h
	expiry := time.Now().Add(24 * time.Hour)
	item.HighlightExpiry = &expiry

	config.DB.Save(user)
	config.DB.Save(&item)

	c.Redirect(http.StatusFound, "/profile")
}

// EditItemPage displays the edit form for a post
func EditItemPage(c *gin.Context) {
	itemID := c.Param("pk")
	user := c.MustGet("user").(*models.User)

	var item models.LostItem
	if err := config.DB.Preload("User").First(&item, itemID).Error; err != nil {
		c.String(http.StatusNotFound, "Item not found")
		return
	}

	// Check ownership
	if item.UserID != user.ID {
		c.String(http.StatusForbidden, "Not authorized to edit this item")
		return
	}

	// Fetch categories for dropdown
	var subcategories []models.SubCategory
	config.DB.Find(&subcategories)

	ctx := utils.GetGlobalContext(c)
	ctx["item"] = item
	ctx["subcategories"] = subcategories

	tpl, err := pongo2.FromFile("templates/core/edit_item.html")
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

// EditItem handles the edit form submission
func EditItem(c *gin.Context) {
	itemID := c.Param("pk")
	user := c.MustGet("user").(*models.User)

	// Check if user is banned
	if user.IsBanned {
		c.String(http.StatusForbidden, "Your account has been banned. You cannot edit posts.")
		return
	}

	var item models.LostItem
	if err := config.DB.First(&item, itemID).Error; err != nil {
		c.String(http.StatusNotFound, "Item not found")
		return
	}

	// Check ownership
	if item.UserID != user.ID {
		c.String(http.StatusForbidden, "Not authorized to edit this item")
		return
	}

	// Get form data
	title := c.PostForm("title")
	desc := c.PostForm("description")
	location := c.PostForm("location")
	bountyStr := c.PostForm("bounty_coins")
	subCatIDStr := c.PostForm("subcategory")

	bounty, _ := strconv.Atoi(bountyStr)
	subCatID, _ := strconv.ParseInt(subCatIDStr, 10, 64)

	// Handle Image Upload (if new image provided)
	file, err := c.FormFile("image")
	if err == nil {
		filename := fmt.Sprintf("%d_%s", time.Now().Unix(), filepath.Base(file.Filename))
		dst := "static/images/" + filename
		if err := c.SaveUploadedFile(file, dst); err == nil {
			item.Image = "images/" + filename
		}
	}

	// Fetch Category ID from SubCategory
	var subCat models.SubCategory
	var catID int64
	if config.DB.First(&subCat, subCatID).Error == nil {
		catID = subCat.CategoryID
	}

	// Update item fields
	item.Title = title
	item.Description = desc
	item.Location = location
	item.BountyCoins = bounty
	item.SubCategoryID = &subCatID
	item.CategoryID = catID

	config.DB.Save(&item)
	c.Redirect(http.StatusFound, "/item/"+itemID)
}

// DeleteItem allows user to delete their own post
func DeleteItem(c *gin.Context) {
	itemID := c.Param("pk")
	user := c.MustGet("user").(*models.User)

	var item models.LostItem
	if err := config.DB.First(&item, itemID).Error; err != nil {
		c.String(http.StatusNotFound, "Item not found")
		return
	}

	// Check ownership
	if item.UserID != user.ID {
		c.String(http.StatusForbidden, "Not authorized to delete this item")
		return
	}

	// Delete the item
	if err := config.DB.Delete(&item).Error; err != nil {
		c.String(http.StatusInternalServerError, "Failed to delete item")
		return
	}

	c.Redirect(http.StatusFound, "/")
}
