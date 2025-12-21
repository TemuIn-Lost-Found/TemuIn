package handlers

import (
	"net/http"
	"temuin/config"
	"temuin/models"
	"temuin/utils"

	"github.com/flosch/pongo2/v6"
	"github.com/gin-gonic/gin"
)

// AdminDeleteItem allows admin to delete any post
func AdminDeleteItem(c *gin.Context) {
	itemID := c.Param("pk")

	var item models.LostItem
	if err := config.DB.First(&item, itemID).Error; err != nil {
		c.String(http.StatusNotFound, "Item not found")
		return
	}

	// Admin can delete any post - no ownership check needed
	if err := config.DB.Delete(&item).Error; err != nil {
		c.String(http.StatusInternalServerError, "Failed to delete item")
		return
	}

	c.Redirect(http.StatusFound, "/")
}

// BanUser allows admin to ban a user account
func BanUser(c *gin.Context) {
	userID := c.Param("id")

	var targetUser models.User
	if err := config.DB.First(&targetUser, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Prevent admin from banning themselves
	currentUser := c.MustGet("user").(*models.User)
	if targetUser.ID == currentUser.ID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot ban yourself"})
		return
	}

	// Set user as banned
	targetUser.IsBanned = true
	if err := config.DB.Save(&targetUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to ban user"})
		return
	}

	// Return success - can be JSON or redirect depending on frontend
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "User banned successfully"})
}

// UnbanUser allows admin to unban a user account
func UnbanUser(c *gin.Context) {
	userID := c.Param("id")

	var targetUser models.User
	if err := config.DB.First(&targetUser, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Set user as unbanned
	targetUser.IsBanned = false
	if err := config.DB.Save(&targetUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unban user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "User unbanned successfully"})
}

// AdminDashboard displays admin control panel (optional feature)
func AdminDashboard(c *gin.Context) {
	// Fetch statistics
	var totalPosts int64
	var totalUsers int64
	var bannedUsersCount int64

	config.DB.Model(&models.LostItem{}).Count(&totalPosts)
	config.DB.Model(&models.User{}).Count(&totalUsers)
	config.DB.Model(&models.User{}).Where("is_banned = ?", true).Count(&bannedUsersCount)

	// Fetch recent posts
	var recentPosts []models.LostItem
	config.DB.Preload("User").Order("created_at DESC").Limit(20).Find(&recentPosts)

	// Fetch banned users
	var bannedUsers []models.User
	config.DB.Where("is_banned = ?", true).Find(&bannedUsers)

	// Add global context utilities
	ctx := utils.GetGlobalContext(c)
	ctx["total_posts"] = totalPosts
	ctx["total_users"] = totalUsers
	ctx["banned_users_count"] = bannedUsersCount
	ctx["recent_posts"] = recentPosts
	ctx["banned_users"] = bannedUsers

	tpl, err := pongo2.FromFile("templates/admin_dashboard.html")
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

// AdminWithdrawalsPage displays all withdrawal requests
func AdminWithdrawalsPage(c *gin.Context) {
	var withdrawals []models.WithdrawalRequest
	config.DB.Preload("User").Order("created_at desc").Find(&withdrawals)

	ctx := utils.GetGlobalContext(c)
	ctx["withdrawals"] = withdrawals

	tpl, err := pongo2.FromFile("templates/admin_withdrawals.html")
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

// AdminApproveWithdrawal approves a pending withdrawal
func AdminApproveWithdrawal(c *gin.Context) {
	id := c.Param("id")

	var wr models.WithdrawalRequest
	if err := config.DB.Preload("User").First(&wr, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Withdrawal request not found"})
		return
	}

	if wr.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request is not pending"})
		return
	}

	tx := config.DB.Begin()

	// Update status
	wr.Status = "approved"
	now := utils.GetCurrentTime()
	wr.ProcessedAt = &now

	if err := tx.Save(&wr).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status"})
		return
	}

	// Create notification
	notification := models.Notification{
		UserID:  wr.UserID,
		Type:    "system_update",
		Title:   "Withdrawal Approved",
		Message: "Your withdrawal request for " + utils.FormatRupiah(wr.Amount) + " has been approved and is being processed.",
	}
	if err := tx.Create(&notification).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// AdminRejectWithdrawal rejects and refunds a withdrawal
func AdminRejectWithdrawal(c *gin.Context) {
	id := c.Param("id")

	var wr models.WithdrawalRequest
	if err := config.DB.Preload("User").First(&wr, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Withdrawal request not found"})
		return
	}

	if wr.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request is not pending"})
		return
	}

	tx := config.DB.Begin()

	// 1. Update status
	wr.Status = "rejected"
	now := utils.GetCurrentTime()
	wr.ProcessedAt = &now

	if err := tx.Save(&wr).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status"})
		return
	}

	// 2. Refund coins to user
	var user models.User
	if err := tx.First(&user, wr.UserID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	user.CoinBalance += wr.Coins
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refund coins"})
		return
	}

	// 3. Create Refund Transaction Record
	transaction := models.CoinTransaction{
		UserID:          user.ID,
		Amount:          wr.Coins,
		TransactionType: "withdraw_refund",
	}
	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create transaction record"})
		return
	}

	// 4. Create notification
	notification := models.Notification{
		UserID:  wr.UserID,
		Type:    "warning",
		Title:   "Withdrawal Rejected",
		Message: "Your withdrawal request for " + utils.FormatRupiah(wr.Amount) + " has been rejected. The coins have been refunded to your balance.",
	}
	if err := tx.Create(&notification).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"success": true})
}
