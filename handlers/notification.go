package handlers

import (
	"net/http"
	"strconv"
	"temuin/config"
	"temuin/models"
	"temuin/utils"

	"github.com/flosch/pongo2/v6"
	"github.com/gin-gonic/gin"
)

// GetNotifications returns user's notifications with pagination
func GetNotifications(c *gin.Context) {
	user := c.MustGet("user").(*models.User)

	// Get limit from query params (default 20)
	limitStr := c.DefaultQuery("limit", "20")
	limit, _ := strconv.Atoi(limitStr)

	var notifications []models.Notification
	config.DB.Where("user_id = ?", user.ID).
		Order("created_at DESC").
		Limit(limit).
		Find(&notifications)

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"notifications": notifications,
	})
}

// GetUnreadCount returns count of unread notifications
func GetUnreadCount(c *gin.Context) {
	user := c.MustGet("user").(*models.User)

	var count int64
	config.DB.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", user.ID, false).
		Count(&count)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"count":   count,
	})
}

// MarkAsRead marks a notification as read
func MarkAsRead(c *gin.Context) {
	notificationID := c.Param("id")
	user := c.MustGet("user").(*models.User)

	var notification models.Notification
	if err := config.DB.Where("id = ? AND user_id = ?", notificationID, user.ID).First(&notification).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}

	notification.IsRead = true
	if err := config.DB.Save(&notification).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// MarkAllAsRead marks all user notifications as read
func MarkAllAsRead(c *gin.Context) {
	user := c.MustGet("user").(*models.User)

	if err := config.DB.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", user.ID, false).
		Update("is_read", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark all as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// NotificationListPage displays the full notifications page
func NotificationListPage(c *gin.Context) {
	user := c.MustGet("user").(*models.User)

	var notifications []models.Notification
	config.DB.Where("user_id = ?", user.ID).
		Order("created_at DESC").
		Limit(50).
		Find(&notifications)

	ctx := utils.GetGlobalContext(c)
	ctx["notifications"] = notifications

	tpl, err := pongo2.FromFile("templates/core/notifications.html")
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
