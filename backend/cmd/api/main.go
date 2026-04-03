package main

import (
	"log"
	"net/http"
	"os"

	"github.com/username/kafe-backend/internal/database"
	"github.com/username/kafe-backend/internal/handlers"
	"github.com/username/kafe-backend/internal/middleware"
	"github.com/username/kafe-backend/internal/repository"
	"github.com/username/kafe-backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	// Initialize Database
	if err := database.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.DB.Close()

	// Initialize Repositories
	userRepo := repository.NewUserRepository(database.DB)
	catRepo := repository.NewCategoryRepository(database.DB)
	prodRepo := repository.NewProductRepository(database.DB)
	orderRepo := repository.NewOrderRepository(database.DB)

	// Initialize Services
	authService := service.NewAuthService(userRepo)
	catalogService := service.NewCatalogService(catRepo, prodRepo)
	wsService := service.NewWebsocketService()
	botService := service.NewBotService()
	printerService := service.NewPrinterService()
	orderService := service.NewOrderService(orderRepo, prodRepo, wsService, botService, printerService)

	// Initialize Handlers
	authHandler := handlers.NewAuthHandler(authService, userRepo)
	catalogHandler := handlers.NewCatalogHandler(catalogService)
	orderHandler := handlers.NewOrderHandler(orderService)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := gin.Default()

	// CORS Middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Static File Serving for Uploads
	r.Static("/uploads", "./uploads")

	// Routes
	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.GET("/me", middleware.AuthMiddleware(), authHandler.Me)
			auth.PUT("/me", middleware.AuthMiddleware(), authHandler.UpdateProfile)
		}

		// Catalog (Public and Admin)
		catalog := api.Group("/catalog")
		{
			// Public
			catalog.GET("/categories", catalogHandler.GetAllCategories)
			catalog.GET("/products", catalogHandler.GetAllProducts)
			catalog.GET("/categories/:cat_id/products", catalogHandler.GetProductsByCategory)

			// Admin Protected
			admin := catalog.Group("/")
			admin.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("admin"))
			{
				admin.POST("/categories", catalogHandler.CreateCategory)
				admin.PUT("/categories/:id", catalogHandler.UpdateCategory)
				admin.DELETE("/categories/:id", catalogHandler.DeleteCategory)

				admin.POST("/products", catalogHandler.CreateProduct)
				admin.PUT("/products/:id", catalogHandler.UpdateProduct)
				admin.DELETE("/products/:id", catalogHandler.DeleteProduct)

				// Staff Management
				admin.POST("/staff", authHandler.Register)
				admin.GET("/staff", authHandler.GetStaff)
				admin.GET("/performance", orderHandler.GetStaffPerformance)

				// Image Upload
				admin.POST("/upload", handlers.UploadImage)
			}
		}

		// Orders
		orders := api.Group("/orders")
		orders.Use(middleware.AuthMiddleware())
		{
			orders.POST("/", orderHandler.CreateOrder)
			orders.GET("/my", orderHandler.GetMyOrders)
			orders.GET("/:id", orderHandler.GetOrderByID)
			orders.GET("/:id/ratings", orderHandler.GetOrderRatings)
			orders.POST("/:id/rate", orderHandler.SubmitRating)

			// Staff/Admin Protected
			staff := orders.Group("/")
			staff.Use(middleware.RoleMiddleware("admin", "cook", "courier"))
			{
				staff.GET("/all", orderHandler.GetAllOrders)
				staff.GET("/active", orderHandler.GetActiveOrders)
				staff.GET("/stats", orderHandler.GetStats)
				staff.PUT("/:id/status", orderHandler.UpdateStatus)
				staff.POST("/:id/assign", orderHandler.AssignCourier)
			}
		}

		// WebSocket
		api.GET("/ws", middleware.AuthMiddleware(), func(c *gin.Context) {
			userID, _ := c.Get("user_id")
			role, _ := c.Get("role")
			wsService.HandleConnection(c.Writer, c.Request, userID.(int), role.(string))
		})

		// Health Check
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status": "up",
			})
		})
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
