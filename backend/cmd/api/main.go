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
	settingsRepo := repository.NewSettingsRepository(database.DB)
	financeRepo := repository.NewFinanceRepository(database.DB)
	inventoryRepo := repository.NewInventoryRepository(database.DB)
	tableRepo := repository.NewTableRepository(database.DB)

	// Initialize Services
	authService := service.NewAuthService(userRepo)
	catalogService := service.NewCatalogService(catRepo, prodRepo)
	wsService := service.NewWebsocketService()
	botService := service.NewBotService()
	printerService := service.NewPrinterService()
	orderService := service.NewOrderService(orderRepo, prodRepo, settingsRepo, inventoryRepo, wsService, botService, printerService)

	// Initialize Handlers
	authHandler := handlers.NewAuthHandler(authService, userRepo)
	catalogHandler := handlers.NewCatalogHandler(catalogService)
	orderHandler := handlers.NewOrderHandler(orderService)
	settingsHandler := handlers.NewSettingsHandler(settingsRepo)
	financeHandler := handlers.NewFinanceHandler(financeRepo)
	inventoryHandler := handlers.NewInventoryHandler(inventoryRepo)
	tableHandler := handlers.NewTableHandler(tableRepo, orderService)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := gin.Default()

	// CORS Middleware
	r.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		allowedOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
		if allowedOrigins == "" {
			allowedOrigins = "http://localhost:3000,http://localhost:5173" // Default safe local origins
		}

		if origin != "" && strings.Contains(allowedOrigins, origin) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else if origin == "" {
			// Allow non-browser clients (like bots/mobile) if no origin is set
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}
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
				admin.DELETE("/staff/:id", authHandler.DeleteStaff)
				admin.GET("/performance", orderHandler.GetStaffPerformance)

				// Image Upload
				admin.POST("/upload", handlers.UploadImage)

				// Global Settings
				admin.PUT("/settings", settingsHandler.UpdateSettings)
			}
			catalog.GET("/settings", settingsHandler.GetSettings)
		}
		orders := api.Group("/orders")
		orders.Use(middleware.AuthMiddleware())
		{
			orders.POST("/", orderHandler.CreateOrder)
			orders.GET("/my", orderHandler.GetMyOrders)

			// Staff/Admin Protected
			staff := orders.Group("/")
			staff.Use(middleware.RoleMiddleware("admin", "cook", "courier", "waiter"))
			{
				staff.GET("/all", orderHandler.GetAllOrders)
				staff.GET("/active", orderHandler.GetActiveOrders)
				staff.GET("/stats", orderHandler.GetStats)
				staff.GET("/waiter-history", orderHandler.GetWaiterHistory)
				staff.GET("/waiter-active/:waiterID", orderHandler.GetActiveOrdersByWaiter)
				staff.GET("/waiter-hist/:waiterID", orderHandler.GetOrderHistoryByWaiter)
				staff.PUT("/:id/status", orderHandler.UpdateStatus)
				staff.POST("/:id/assign", orderHandler.AssignCourier)
				staff.POST("/:id/print", orderHandler.ReprintOrder)
				staff.PUT("/:id/service-fee", orderHandler.SetServiceFee)
			}

			// Parameterized routes must be at the end to avoid shadowing
			orders.GET("/:id", orderHandler.GetOrderByID)
			orders.GET("/:id/ratings", orderHandler.GetOrderRatings)
			orders.POST("/:id/rate", orderHandler.SubmitRating)
		}

		// Tables (Admin and Waiter)
		tables := api.Group("/tables")
		tables.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("admin", "waiter"))
		{
			tables.GET("/", tableHandler.GetAll)
			tables.POST("/", tableHandler.Create)
			tables.PUT("/:id", tableHandler.Update)
			tables.DELETE("/:id", tableHandler.Delete)
		}
		
		// Finance (Admin only)
		finance := api.Group("/finance")
		finance.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("admin"))
		{
			finance.GET("/stats", financeHandler.GetStats)
			finance.GET("/expenses", financeHandler.GetExpenses)
			finance.POST("/expenses", financeHandler.CreateExpense)
		}

		// Inventory (Admin only)
		inventory := api.Group("/inventory")
		inventory.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("admin"))
		{
			inventory.GET("/ingredients", inventoryHandler.GetIngredients)
			inventory.POST("/ingredients", inventoryHandler.CreateIngredient)
			inventory.PUT("/ingredients/:id", inventoryHandler.UpdateIngredient)
			inventory.DELETE("/ingredients/:id", inventoryHandler.DeleteIngredient)
			inventory.POST("/ingredients/:id/restock", inventoryHandler.RestockIngredient)

			inventory.GET("/recipes/:product_id", inventoryHandler.GetProductIngredients)
			inventory.POST("/recipes", inventoryHandler.AddProductIngredient)
			inventory.DELETE("/recipes/:id", inventoryHandler.DeleteProductIngredient)
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
			expectedPK := os.Getenv("PRINTER_SECRET")
			if pk != "" && expectedPK != "" && pk == expectedPK {
				wsService.HandleConnection(c.Writer, c.Request, 0, "printer")
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
			expectedPK := os.Getenv("PRINTER_SECRET")
			if expectedPK == "" || pk != expectedPK {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid key"})
				return
			}

			id, _ := strconv.Atoi(c.Param("id"))
			order, err := orderHandler.Service().GetOrderWithItems(id)
			if err != nil || order == nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
				return
			}

			// Broadcast to all roles including printer
			wsService.BroadcastToRole("admin", map[string]interface{}{"type": "new_order", "order": order})
			wsService.BroadcastToRole("cook", map[string]interface{}{"type": "new_order", "order": order})
			wsService.BroadcastToRole("waiter", map[string]interface{}{"type": "new_order", "order": order})
			wsService.BroadcastToRole("printer", map[string]interface{}{"type": "new_order", "order": order})

			c.JSON(http.StatusOK, gin.H{"status": "notified"})
		})

		// Health Check
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status": "up",
			})
		})

		// WebSocket Connectivity Test
		api.GET("/ws-test", func(c *gin.Context) {
			log.Printf("✅ TEST: Bridge reached /api/ws-test from %s", c.ClientIP())
			c.String(http.StatusOK, "WS_TEST_OK")
		})
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
