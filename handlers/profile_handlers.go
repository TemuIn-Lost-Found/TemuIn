package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"temuin/config"
	"temuin/models"
	"time"

	"github.com/gin-gonic/gin"
)

// UpdateProfile handles profile updates (username and profile picture)
func UpdateProfile(c *gin.Context) {
	user := c.MustGet("user").(*models.User)

	newUsername := c.PostForm("username")
	isJSON := c.GetHeader("Accept") == "application/json"

	// Validate username if provided
	if newUsername != "" && newUsername != user.Username {
		// Check if username is already taken
		var existingUser models.User
		if err := config.DB.Where("username = ? AND id != ?", newUsername, user.ID).First(&existingUser).Error; err == nil {
			if isJSON {
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Username sudah digunakan"})
				return
			}
			c.Redirect(http.StatusFound, "/profile?error=username_taken")
			return
		}

		// Update username
		user.Username = newUsername
	}

	// Handle profile picture deletion
	if c.PostForm("delete_picture") == "true" {
		if user.ProfilePicture != "" {
			profilesDir := "static/images/profiles"
			oldPath := filepath.Join(profilesDir, user.ProfilePicture)
			os.Remove(oldPath)
			user.ProfilePicture = ""
		}
	}

	// Handle profile picture upload
	file, err := c.FormFile("profile_picture")
	if err == nil {
		// Create profiles directory if not exists
		profilesDir := "static/images/profiles"
		os.MkdirAll(profilesDir, os.ModePerm)

		// Generate unique filename
		ext := filepath.Ext(file.Filename)
		filename := fmt.Sprintf("%d_%d%s", user.ID, time.Now().Unix(), ext)
		savePath := filepath.Join(profilesDir, filename)

		// Save file
		if err := c.SaveUploadedFile(file, savePath); err != nil {
			if isJSON {
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Gagal menyimpan foto"})
				return
			}
			c.Redirect(http.StatusFound, "/profile?error=upload_failed")
			return
		}

		// Delete old profile picture if exists AND we haven't already deleted it above
		if user.ProfilePicture != "" && user.ProfilePicture != filename {
			oldPath := filepath.Join(profilesDir, user.ProfilePicture)
			os.Remove(oldPath)
		}

		user.ProfilePicture = filename
	}

	// Save updates to database
	if err := config.DB.Save(user).Error; err != nil {
		if isJSON {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Gagal menyimpan perubahan"})
			return
		}
		c.Redirect(http.StatusFound, "/profile?error=save_failed")
		return
	}

	if isJSON {
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Profile berhasil diperbarui"})
		return
	}

	c.Redirect(http.StatusFound, "/profile?success=true")
}

// GetProfilePicture serves profile pictures
func GetProfilePicture(c *gin.Context) {
	userID := c.Param("user_id")

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	if user.ProfilePicture == "" {
		c.Status(http.StatusNotFound)
		return
	}

	filepath := filepath.Join("static/images/profiles", user.ProfilePicture)
	c.File(filepath)
}
