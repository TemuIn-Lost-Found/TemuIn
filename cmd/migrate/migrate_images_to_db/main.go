package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings" // Added missing import
	"temuin/config"
	"temuin/models"

	"github.com/joho/godotenv"
)

func main() {
	// load env
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	config.ConnectDB()

	// AutoMigrate the new table
	fmt.Println("Migrating database schema...")
	if err := config.DB.AutoMigrate(&models.LostItemImage{}); err != nil {
		log.Fatal("Failed to migrate:", err)
	}

	// Fetch all LostItems with images
	var items []models.LostItem
	if err := config.DB.Where("image IS NOT NULL AND image != ''").Find(&items).Error; err != nil {
		log.Fatal("Failed to fetch items:", err)
	}

	fmt.Printf("Found %d items with images to migrate.\n", len(items))

	successCount := 0
	skipCount := 0
	failCount := 0

	for _, item := range items {
		// Image path string in DB is like "images/17123123_foo.jpg" or "17123123_foo.jpg" depending on how it was saved.
		// Handlers save it to "static/images/" + filename
		// And DB stores "images/" + filename
		
		// If stored as "images/filename.jpg", the physical path is "static/images/filename.jpg"
		// Wait, handler said:
		// dst := "static/images/" + filename
        // imagePath = "images/" + filename
		
		// So physical path is "static/" + item.Image
		
		// Construct physical path
		// Remove "images/" prefix if present to be safe, or just prepend generic static path logic
		
		// Let's assume item.Image is "images/foo.jpg"
		
		physicalPath := filepath.Join("static", item.Image)
		
		// Check if file exists
		if _, err := os.Stat(physicalPath); os.IsNotExist(err) {
			// Try without 'images/' prefix just in case
			altPath := filepath.Join("static", "images", filepath.Base(item.Image))
			if _, err2 := os.Stat(altPath); os.IsNotExist(err2) {
				fmt.Printf("File not found for item %d: %s (checked %s and %s)\n", item.ID, item.Image, physicalPath, altPath)
				failCount++
				continue
			}
			physicalPath = altPath
		}

		// Read file
		data, err := ioutil.ReadFile(physicalPath)
		if err != nil {
			fmt.Printf("Failed to read file for item %d: %v\n", item.ID, err)
			failCount++
			continue
		}

		// Determine content type
		ext := strings.ToLower(filepath.Ext(physicalPath))
		contentType := "image/jpeg" // default
		if ext == ".png" {
			contentType = "image/png"
		} else if ext == ".webp" {
			contentType = "image/webp"
		} else if ext == ".gif" {
			contentType = "image/gif"
		}

		// Check if it already exists in DB
		var existing models.LostItemImage
		if err := config.DB.First(&existing, item.ID).Error; err == nil {
			// Already exists
			// Skip or update? Skip strictly for now to avoid overwriting
			// fmt.Printf("Image for item %d already in DB, skipping.\n", item.ID)
			skipCount++
			continue
		}

		// Save to DB
		imgEntry := models.LostItemImage{
			ItemID:      item.ID,
			ImageData:   data,
			ContentType: contentType,
		}

		if err := config.DB.Create(&imgEntry).Error; err != nil {
			fmt.Printf("Failed to save DB entry for item %d: %v\n", item.ID, err)
			failCount++
		} else {
			fmt.Printf("Migrated image for item %d (%d bytes)\n", item.ID, len(data))
			successCount++
		}
	}

	fmt.Println("------------------------------------------------")
	fmt.Printf("Migration Complete.\nSuccess: %d\nSkipped: %d\nFailed: %d\n", successCount, skipCount, failCount)
}

