package main

import (
	"log"
	"os"
	"temuin/config"
	"temuin/routes"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 0. Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// 1. Connect DB
	config.ConnectDB()

	// 1.5. Initialize Google OAuth
	config.InitGoogleOAuth()

	// 1.6. Initialize Midtrans
	config.InitMidtrans()

	// 2. Setup Gin
	r := gin.Default()

	// 3. Static Files
	r.Static("/static", "./static")
	// Media files currently served from django's media root?
	// For now let's serve media from parent if needed, or just assume same relative path
	r.Static("/media", "../media")

	// 4. Session Store
	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))

	// 5. Routes
	routes.RegisterRoutes(r)

	port := os.Getenv("WEBSITES_PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)

}
