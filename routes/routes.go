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

		// Highlights pages
		public.GET("/highlights", handlers.AllHighlightsPage)
		public.GET("/highlights/category/:pk", handlers.CategoryHighlightsPage)

		public.GET("/category/:pk", handlers.CategoryPage)
		public.GET("/subcategory/:pk", handlers.SubCategoryPage)

		// Midtrans payment notification callback (public, no auth)
		public.POST("/topup/notification", handlers.MidtransNotification)
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

		// TopUp routes
		authorized.POST("/topup/initiate", handlers.InitiateTopUp)
		authorized.POST("/topup/confirm", handlers.ConfirmTopUp)
		authorized.GET("/topup/history", handlers.GetTopUpHistory)
		authorized.GET("/topup/status/:order_id", handlers.CheckTopUpStatus)

		authorized.POST("/item/:pk/highlight", handlers.HighlightItem)
		authorized.POST("/item/:pk/found", handlers.MarkAsFound)
		authorized.POST("/item/:pk/select-finder", handlers.SelectFinder) // NEW
		authorized.POST("/item/:pk/return", handlers.ConfirmReturn)

		// User post management
		authorized.GET("/item/:pk/edit", handlers.EditItemPage)
		authorized.POST("/item/:pk/edit", handlers.EditItem)
		authorized.POST("/item/:pk/delete", handlers.DeleteItem)

		// Report routes
		authorized.POST("/item/:pk/report", handlers.SubmitReport)

		// Notification routes
		authorized.GET("/notifications", handlers.NotificationListPage)
		authorized.GET("/api/notifications", handlers.GetNotifications)
		authorized.GET("/api/notifications/count", handlers.GetUnreadCount)
		authorized.POST("/api/notifications/:id/read", handlers.MarkAsRead)
		authorized.POST("/api/notifications/read-all", handlers.MarkAllAsRead)

		// witdhrawal
		authorized.POST("/withdraw", handlers.RequestWithdrawal)
		authorized.GET("/withdrawals", handlers.GetWithdrawalHistory)
	}

	// Admin Routes
	admin := r.Group("/admin")
	admin.Use(middleware.AuthRequired(), middleware.AdminRequired())
	{
		admin.GET("/dashboard", handlers.AdminDashboard)
		admin.POST("/item/:pk/delete", handlers.AdminDeleteItem)
		admin.POST("/user/:id/ban", handlers.BanUser)
		admin.POST("/user/:id/unban", handlers.UnbanUser)

		// Report management
		admin.GET("/reports", handlers.AdminReportList)
		admin.POST("/report/:id/resolve", handlers.ResolveReport)
		admin.POST("/report/:id/warn", handlers.WarnUser)
	}
}
