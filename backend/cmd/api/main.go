package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"strings"

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

	// 🌐 Serving Frontend in Production Mode
	// Serve the static files from the frontend build (dist folder)
	r.StaticFile("/", "../../frontend/dist/index.html")
	r.Static("/assets", "../../frontend/dist/assets")
	
	// Fallback for SPA routing: All other unknown routes go to index.html
	r.NoRoute(func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.URL.Path, "/api") && !strings.HasPrefix(c.Request.URL.Path, "/uploads") {
			c.File("../../frontend/dist/index.html")
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	})

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

		// Printer Control (Staff Only)
		printer := api.Group("/printer")
		printer.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("admin", "cook", "courier"))
		{
			printer.GET("/test", orderHandler.TestPrinter)
		}

		// WebSocket
		api.GET("/ws", func(c *gin.Context) {
			pk := c.Query("printer_key")
			if pk == "KAFE_PRINTER_SECRET_2026" {
				wsService.HandleConnection(c.Writer, c.Request, 0, "admin")
				return
			}

			middleware.AuthMiddleware()(c)
			if c.IsAborted() {
				return
			}
			
			uID, _ := c.Get("user_id")
			rol, _ := c.Get("role")
			wsService.HandleConnection(c.Writer, c.Request, uID.(int), rol.(string))
		})

		// Internal Notify (For Telegram Bot)
		api.GET("/notify-order/:id", func(c *gin.Context) {
			pk := c.Query("key")
			if pk != "KAFE_PRINTER_SECRET_2026" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid key"})
				return
			}

			id, _ := strconv.Atoi(c.Param("id"))
			order, err := orderHandler.Service().GetOrderByID(id, 0, "admin")
			if err != nil || order == nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
				return
			}

			// Broadcast to all roles including printer
			wsService.BroadcastToRole("admin", map[string]interface{}{"type": "new_order", "order": order})
			wsService.BroadcastToRole("cook", map[string]interface{}{"type": "new_order", "order": order})
			wsService.BroadcastToRole("printer", map[string]interface{}{"type": "new_order", "order": order})

			c.JSON(http.StatusOK, gin.H{"status": "notified"})
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
