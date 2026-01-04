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
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	config.ConnectDB()
	config.InitGoogleOAuth()
	config.InitMidtrans()

	r := gin.Default()

	r.Static("/static", "./static")
	r.Static("/media", "../media")

	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))

	// OPTIONAL health check (AMAN)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "OK"})
	})

	routes.RegisterRoutes(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Starting server on port", port)
	if err := r.Run("0.0.0.0:" + port); err != nil {
		log.Fatal("❌ Failed to start server:", err)
	}
}