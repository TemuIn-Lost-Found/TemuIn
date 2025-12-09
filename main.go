package main

import (
	"temuin/config"
	"temuin/routes"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func main() {
    // 1. Connect DB
    config.ConnectDB()

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

    r.Run(":8080")
}
