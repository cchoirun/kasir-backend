package main

import (
	"kasir-backend/config"
	"kasir-backend/controllers"
	"kasir-backend/middleware"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Jalankan Gin dalam mode Release untuk performa produksi yang optimal
	gin.SetMode(gin.ReleaseMode)

	config.ConnectDB()
	controllers.SeedOwner()

	r := gin.Default()

	// Konfigurasi CORS (Nanti AllowOrigins bisa diubah ke domain Vercel-mu)
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Ubah nanti menjadi ["https://namadomain.vercel.app"] setelah frontend rilis
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Ping route untuk health check
	r.GET("/api/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong", "status": "online"})
	})

	// Auth Routes
	authRoutes := r.Group("/api/auth")
	{
		authRoutes.POST("/login", controllers.Login)
		authRoutes.PUT("/profile", middleware.AuthMiddleware(""), controllers.UpdateProfile)
		authRoutes.PUT("/change-password", middleware.AuthMiddleware(""), controllers.ChangePassword)
	}

	// Product Routes
	productRoutes := r.Group("/api/products")
	productRoutes.Use(middleware.AuthMiddleware(""))
	{
		productRoutes.GET("", controllers.GetProducts)
		productRoutes.POST("", middleware.AuthMiddleware("owner"), controllers.CreateProduct)
		productRoutes.PUT("/:id", middleware.AuthMiddleware("owner"), controllers.UpdateProduct)
		productRoutes.DELETE("/:id", middleware.AuthMiddleware("owner"), controllers.DeleteProduct)
	}

	// Transaction Routes
	trxRoutes := r.Group("/api/transactions")
	trxRoutes.Use(middleware.AuthMiddleware(""))
	{
		trxRoutes.POST("", controllers.CreateTransaction)
		trxRoutes.GET("/:id/qris", controllers.GetQrisSimulation)
		trxRoutes.GET("", controllers.GetTransactions)
	}

	// Webhook & Analytics Routes
	r.POST("/api/webhooks/payment", controllers.WebhookPayment)
	r.GET("/api/analytics/dashboard", middleware.AuthMiddleware("owner"), controllers.GetDashboardAnalytics)

	// Pegawai / Users Routes
	userRoutes := r.Group("/api/users")
	userRoutes.Use(middleware.AuthMiddleware("owner"))
	{
		userRoutes.GET("", controllers.GetPegawai)
		userRoutes.POST("", controllers.CreatePegawai)
		userRoutes.DELETE("/:id", controllers.DeletePegawai)
	}

	r.Run(":8080")
}
