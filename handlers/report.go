package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"temuin/config"
	"temuin/models"
	"temuin/utils"

	"github.com/flosch/pongo2/v6"
	"github.com/gin-gonic/gin"
)

// SubmitReport handles report submission from users
func SubmitReport(c *gin.Context) {
	itemID := c.Param("pk")
	user := c.MustGet("user").(*models.User)

	// Check if user is banned
	if user.IsBanned {
		c.JSON(http.StatusForbidden, gin.H{"error": "Your account has been banned."})
		return
	}

	// Get form data
	reason := c.PostForm("reason")
	description := c.PostForm("description")

	// Validate reason
	validReasons := map[string]bool{
		"fraud":          true,
		"spam":           true,
		"buying_selling": true,
		"inappropriate":  true,
		"other":          true,
	}

	if !validReasons[reason] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid report reason"})
		return
	}

	// Check if item exists
	var item models.LostItem
	iid, _ := strconv.ParseInt(itemID, 10, 64)
	if err := config.DB.Preload("User").First(&item, iid).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	// Prevent user from reporting their own post
	if item.UserID == user.ID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You cannot report your own post"})
		return
	}

	// Check if user already reported this item
	var existingReport models.ItemReport
	if err := config.DB.Where("item_id = ? AND reporter_id = ?", iid, user.ID).First(&existingReport).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You have already reported this post"})
		return
	}

	// Create report
	report := models.ItemReport{
		ItemID:      iid,
		ReporterID:  user.ID,
		Reason:      reason,
		Description: description,
		Status:      "pending",
	}

	if err := config.DB.Create(&report).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit report"})
		return
	}

	// Create notification for all admins
	var admins []models.User
	config.DB.Where("is_superuser = ?", true).Find(&admins)

	reasonText := map[string]string{
		"fraud":          "Penipuan",
		"spam":           "Spam",
		"buying_selling": "Jual Beli",
		"inappropriate":  "Konten Tidak Pantas",
		"other":          "Lainnya",
	}

	for _, admin := range admins {
		notification := models.Notification{
			UserID:          admin.ID,
			Type:            "report",
			Title:           "Laporan Baru dari " + user.Username,
			Message:         fmt.Sprintf("User %s melaporkan postingan '%s' dengan alasan: %s", user.Username, item.Title, reasonText[reason]),
			ReferenceURL:    fmt.Sprintf("/admin/reports?highlight=%d", report.ID),
			RelatedItemID:   &iid,
			RelatedReportID: &report.ID,
		}
		config.DB.Create(&notification)
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Laporan berhasil dikirim"})
}

// AdminReportList displays all reports for admin
func AdminReportList(c *gin.Context) {
	// Get all reports
	var reports []models.ItemReport
	config.DB.Preload("Item.User").Preload("Item").Preload("Reporter").Order("created_at DESC").Find(&reports)

	// Get statistics
	var totalReports int64
	var pendingReports int64
	var resolvedReports int64

	config.DB.Model(&models.ItemReport{}).Count(&totalReports)
	config.DB.Model(&models.ItemReport{}).Where("status = ?", "pending").Count(&pendingReports)
	config.DB.Model(&models.ItemReport{}).Where("status = ?", "resolved").Count(&resolvedReports)

	ctx := utils.GetGlobalContext(c)
	ctx["reports"] = reports
	ctx["total_reports"] = totalReports
	ctx["pending_reports"] = pendingReports
	ctx["resolved_reports"] = resolvedReports

	// Check for highlight parameter
	highlightID := c.Query("highlight")
	if highlightID != "" {
		if hid, err := strconv.ParseInt(highlightID, 10, 64); err == nil {
			ctx["highlight_id"] = hid
		}
	}

	tpl, err := pongo2.FromFile("templates/admin_reports.html")
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

// ResolveReport marks a report as resolved
func ResolveReport(c *gin.Context) {
	reportID := c.Param("id")

	var report models.ItemReport
	if err := config.DB.First(&report, reportID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Report not found"})
		return
	}

	report.Status = "resolved"
	if err := config.DB.Save(&report).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve report"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Report resolved"})
}

// WarnUser sends a warning notification to the post owner
func WarnUser(c *gin.Context) {
	reportID := c.Param("id")

	var report models.ItemReport
	if err := config.DB.Preload("Item").Preload("Item.User").First(&report, reportID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Report not found"})
		return
	}

	// Mark report as reviewed
	report.Status = "reviewed"
	config.DB.Save(&report)

	reasonText := map[string]string{
		"fraud":          "penipuan",
		"spam":           "spam",
		"buying_selling": "jual beli",
		"inappropriate":  "konten tidak pantas",
		"other":          "pelanggaran",
	}

	// Create warning notification for post owner
	notification := models.Notification{
		UserID:          report.Item.UserID,
		Type:            "warning",
		Title:           "⚠️ Teguran untuk Postingan Anda",
		Message:         fmt.Sprintf("Postingan '%s' Anda telah dilaporkan karena %s dan ditinjau oleh admin. Mohon patuhi aturan komunitas kami.", report.Item.Title, reasonText[report.Reason]),
		ReferenceURL:    fmt.Sprintf("/item/%d", report.ItemID),
		RelatedItemID:   &report.ItemID,
		RelatedReportID: &report.ID,
	}
	config.DB.Create(&notification)

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Warning sent to user"})
}
