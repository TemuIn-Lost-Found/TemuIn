package routes

import (
	"temuin/handlers"
	"temuin/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	// Public Routes with Optional Auth (for UI)
	public := r.Group("/")
	public.Use(middleware.AuthOptional())
	{
		public.GET("/", handlers.Home)
		public.GET("/login", handlers.LoginPage)
		public.POST("/login", handlers.Login)
		public.GET("/register", handlers.RegisterPage)
		public.POST("/register", handlers.Register)
		public.GET("/logout", handlers.Logout)

		// Google OAuth routes
		public.GET("/auth/google/login", handlers.GoogleLogin)
		public.GET("/auth/google/callback", handlers.GoogleCallback)

		public.GET("/category/:pk", handlers.CategoryPage)
		public.GET("/subcategory/:pk", handlers.SubCategoryPage)
	}

	// Protected Routes
	authorized := r.Group("/")
	authorized.Use(middleware.AuthRequired())
	{
		authorized.GET("/report", handlers.ReportItemPage)
		authorized.POST("/report", handlers.ReportItem)
		authorized.GET("/item/:pk", handlers.ItemDetail)
		authorized.POST("/item/:pk/comment", handlers.PostComment)

		authorized.GET("/profile", handlers.Profile)
		authorized.POST("/topup", handlers.TopUp)
		authorized.POST("/item/:pk/highlight", handlers.HighlightItem)
		authorized.POST("/item/:pk/found", handlers.MarkAsFound)
		authorized.POST("/item/:pk/select-finder", handlers.SelectFinder) // NEW
		authorized.POST("/item/:pk/return", handlers.ConfirmReturn)
	}
}
