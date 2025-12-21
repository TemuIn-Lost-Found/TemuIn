package main

import (
	"fmt"
	"log"
	"temuin/config"
	"temuin/models"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	config.ConnectDB()

	var items []models.LostItem
	// Check for any images that might be pointing to media folder
	config.DB.Where("image LIKE ?", "%media%").Or("image LIKE ?", "%lost_items%").Find(&items)

	if len(items) > 0 {
		fmt.Println("WARNING: Found items referencing media/ folder:")
		for _, item := range items {
			fmt.Printf("ID: %d, Image: %s\n", item.ID, item.Image)
		}
	} else {
		fmt.Println("No items found referencing media/ or lost_items/ paths.")
	}
}
