package handlers

import (
	"fmt"
	"net/http"
	"temuin/config"
	"temuin/models"
	"temuin/utils"
	"time"

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

// AdminDashboard displays admin control panel
func AdminDashboard(c *gin.Context) {
	// Fetch statistics
	var totalPosts int64
	var totalUsers int64
	var bannedUsersCount int64
	var totalVisitors int64

	config.DB.Model(&models.LostItem{}).Count(&totalPosts)
	config.DB.Model(&models.User{}).Count(&totalUsers)
	config.DB.Model(&models.User{}).Where("is_banned = ?", true).Count(&bannedUsersCount)
	config.DB.Model(&models.SiteVisit{}).Count(&totalVisitors)

	// Fetch recent posts
	var recentPosts []models.LostItem
	config.DB.Preload("User").Order("created_at DESC").Limit(20).Find(&recentPosts)

	// Fetch banned users
	var bannedUsers []models.User
	config.DB.Where("is_banned = ?", true).Find(&bannedUsers)

	// Fetch all users with subscription status
	var allUsers []models.User
	config.DB.Order("date_joined DESC").Find(&allUsers)

	// Calculate subscription status for each user
	type UserWithSubscription struct {
		models.User
		IsSubscribed bool
		TotalTopUp   int64
	}

	var usersWithSubscription []UserWithSubscription
	for _, user := range allUsers {
		var successfulTopupCount int64
		config.DB.Model(&models.TopUpTransaction{}).
			Where("user_id = ? AND status = ?", user.ID, "success").
			Count(&successfulTopupCount)

		usersWithSubscription = append(usersWithSubscription, UserWithSubscription{
			User:         user,
			IsSubscribed: successfulTopupCount > 0,
			TotalTopUp:   successfulTopupCount,
		})
	}

	// Add global context utilities
	ctx := utils.GetGlobalContext(c)
	ctx["total_posts"] = totalPosts
	ctx["total_users"] = totalUsers
	ctx["banned_users_count"] = bannedUsersCount
	ctx["total_visitors"] = totalVisitors
	ctx["recent_posts"] = recentPosts
	ctx["banned_users"] = bannedUsers
	ctx["all_users"] = usersWithSubscription

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

// AdminGetVisitorStats returns visitor statistics for charts
func AdminGetVisitorStats(c *gin.Context) {
	period := c.DefaultQuery("period", "daily")

	var visits []models.SiteVisit
	var labels []string
	var data []int

	now := time.Now()

	switch period {
	case "daily":
		// Last 24 hours, grouped by hour
		startTime := now.Add(-24 * time.Hour)
		config.DB.Where("visited_at >= ?", startTime).Find(&visits)

		hourCounts := make(map[int]int)
		for _, visit := range visits {
			hour := visit.VisitedAt.Hour()
			hourCounts[hour]++
		}

		for i := 0; i < 24; i++ {
			hour := (now.Hour() - 23 + i + 24) % 24
			labels = append(labels, fmt.Sprintf("%02d:00", hour))
			data = append(data, hourCounts[hour])
		}

	case "weekly":
		// Last 7 days
		startTime := now.AddDate(0, 0, -7)
		config.DB.Where("visited_at >= ?", startTime).Find(&visits)

		dayCounts := make(map[string]int)
		for _, visit := range visits {
			day := visit.VisitedAt.Format("2006-01-02")
			dayCounts[day]++
		}

		for i := 6; i >= 0; i-- {
			day := now.AddDate(0, 0, -i)
			dayStr := day.Format("2006-01-02")
			labels = append(labels, day.Format("Jan 02"))
			data = append(data, dayCounts[dayStr])
		}

	case "monthly":
		// Last 30 days
		startTime := now.AddDate(0, 0, -30)
		config.DB.Where("visited_at >= ?", startTime).Find(&visits)

		dayCounts := make(map[string]int)
		for _, visit := range visits {
			day := visit.VisitedAt.Format("2006-01-02")
			dayCounts[day]++
		}

		for i := 29; i >= 0; i-- {
			day := now.AddDate(0, 0, -i)
			dayStr := day.Format("2006-01-02")
			labels = append(labels, day.Format("Jan 02"))
			data = append(data, dayCounts[dayStr])
		}

	case "yearly":
		// Last 12 months
		startTime := now.AddDate(-1, 0, 0)
		config.DB.Where("visited_at >= ?", startTime).Find(&visits)

		monthCounts := make(map[string]int)
		for _, visit := range visits {
			month := visit.VisitedAt.Format("2006-01")
			monthCounts[month]++
		}

		for i := 11; i >= 0; i-- {
			month := now.AddDate(0, -i, 0)
			monthStr := month.Format("2006-01")
			labels = append(labels, month.Format("Jan 2006"))
			data = append(data, monthCounts[monthStr])
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"labels": labels,
		"data":   data,
	})
}
