package handlers

import (
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
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
	fileHeader, err := c.FormFile("image")
	var imageBytes []byte
	var imageContentType string
	var imagePathForTemplate string // Data URI or existing path

	if err == nil {
		// New file uploaded
		file, _ := fileHeader.Open()
		defer file.Close()
		imageBytes, _ = ioutil.ReadAll(file)

		// Determine content type
		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		if ext == ".png" {
			imageContentType = "image/png"
		} else if ext == ".webp" {
			imageContentType = "image/webp"
		} else if ext == ".gif" {
			imageContentType = "image/gif"
		} else {
			imageContentType = "image/jpeg"
		}

		// Create Data URI for preview
		encoded := base64.StdEncoding.EncodeToString(imageBytes)
		imagePathForTemplate = "data:" + imageContentType + ";base64," + encoded
	} else {
		// Fallback to previous image if exists (Data URI or DB path?)
		// If it comes from "previous_image" hidden field, it's likely a Data URI.
		imagePathForTemplate = c.PostForm("previous_image")
		
		// If we have a previous image string that is a Data URI, decode it to bytes
		if strings.HasPrefix(imagePathForTemplate, "data:") {
			parts := strings.Split(imagePathForTemplate, ",")
			if len(parts) == 2 {
				// extract content type
				mimeParts := strings.Split(parts[0], ";")
				if len(mimeParts) > 0 {
					imageContentType = strings.TrimPrefix(mimeParts[0], "data:")
				}
				
				decoded, err := base64.StdEncoding.DecodeString(parts[1])
				if err == nil {
					imageBytes = decoded
				}
			}
		}
	}

	// Logic: Check coin balance
	if bounty > 0 {
		if user.CoinBalance < bounty {
			// Re-render template with error
			var categories []models.SubCategory
			config.DB.Find(&categories)

			ctx := utils.GetGlobalContext(c)
			ctx["subcategories"] = categories
			ctx["error"] = "Saldo Coins Tidak Cukup!"
			ctx["title"] = title
			ctx["description"] = desc
			ctx["location"] = location
			ctx["bounty_coins"] = bountyStr
			// Persist dropdowns
			ctx["selected_category"] = c.PostForm("category")
			ctx["selected_subcategory"] = subCatIDStr
			// Persist image (pass Data URI)
			ctx["image_path"] = imagePathForTemplate

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

	// Determine generic status for 'Image' field (just a flag now)
	imageFlag := ""
	if len(imageBytes) > 0 {
		imageFlag = "stored"
	}

	item := models.LostItem{
		Title:         title,
		Description:   desc,
		Location:      location,
		BountyCoins:   bounty,
		UserID:        user.ID,
		Image:         imageFlag, // Just a flag, or we can leave it empty and check LostItemImage table. But 'imageFlag' is useful for quick checks.
		Status:        "LOST",
		SubCategoryID: &subCatID,
		CategoryID:    catID,
	}

	config.DB.Create(&item)
	
	// Create Image Record
	if len(imageBytes) > 0 {
		imgRec := models.LostItemImage{
			ItemID:      item.ID,
			ImageData:   imageBytes,
			ContentType: imageContentType,
		}
		config.DB.Create(&imgRec)
	}

	c.Redirect(http.StatusFound, "/dashboard")
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
	fileHeader, err := c.FormFile("image")
	if err == nil {
		file, _ := fileHeader.Open()
		defer file.Close()
		data, _ := ioutil.ReadAll(file)

		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		contentType := "image/jpeg"
		if ext == ".png" {
			contentType = "image/png"
		} else if ext == ".webp" {
			contentType = "image/webp"
		} else if ext == ".gif" {
			contentType = "image/gif"
		}

		var imgRec models.LostItemImage
		if err := config.DB.First(&imgRec, item.ID).Error; err == nil {
			// Update
			imgRec.ImageData = data
			imgRec.ContentType = contentType
			config.DB.Save(&imgRec)
		} else {
			// Create
			imgRec = models.LostItemImage{
				ItemID:      item.ID,
				ImageData:   data,
				ContentType: contentType,
			}
			config.DB.Create(&imgRec)
		}
		item.Image = "stored"
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

	// Use transaction to ensure full cleanup (Manual Cascade)
	tx := config.DB.Begin()

	// 1. Delete Image
	if err := tx.Where("item_id = ?", item.ID).Delete(&models.LostItemImage{}).Error; err != nil {
		tx.Rollback()
		c.String(http.StatusInternalServerError, "Failed to delete item image")
		return
	}

	// 2. Delete Comments
	if err := tx.Where("item_id = ?", item.ID).Delete(&models.Comment{}).Error; err != nil {
		tx.Rollback()
		c.String(http.StatusInternalServerError, "Failed to delete comments")
		return
	}

	// 3. Delete Claims
	if err := tx.Where("item_id = ?", item.ID).Delete(&models.ItemClaim{}).Error; err != nil {
		tx.Rollback()
		c.String(http.StatusInternalServerError, "Failed to delete claims")
		return
	}

	// 4. Delete Reports
	if err := tx.Where("item_id = ?", item.ID).Delete(&models.ItemReport{}).Error; err != nil {
		tx.Rollback()
		c.String(http.StatusInternalServerError, "Failed to delete reports")
		return
	}

	// 5. Delete Notifications linked to this item
	if err := tx.Where("related_item_id = ?", item.ID).Delete(&models.Notification{}).Error; err != nil {
		tx.Rollback()
		c.String(http.StatusInternalServerError, "Failed to delete notifications")
		return
	}

	// 6. Finally, Delete the Item
	if err := tx.Delete(&item).Error; err != nil {
		tx.Rollback()
		c.String(http.StatusInternalServerError, "Failed to delete item")
		return
	}

	tx.Commit()

	c.Redirect(http.StatusFound, "/dashboard")
}

// GetItemImage serves the image from database
func GetItemImage(c *gin.Context) {
	id := c.Param("pk")
	var imgRec models.LostItemImage
	if err := config.DB.First(&imgRec, id).Error; err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	c.Data(http.StatusOK, imgRec.ContentType, imgRec.ImageData)
}
