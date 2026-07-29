package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/username/kafe-backend/internal/database"
	"github.com/username/kafe-backend/internal/handlers"
	"github.com/username/kafe-backend/internal/middleware"
	"github.com/username/kafe-backend/internal/repository"
	"github.com/username/kafe-backend/internal/service"
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
	auditRepo := repository.NewAuditRepository(database.DB)
	cashierRepo := repository.NewCashierRepository(database.DB)
	debtRepo := repository.NewDebtRepository(database.DB)
	refundRepo := repository.NewRefundRepository(database.DB)
	fiscalRepo := repository.NewFiscalRepository(database.DB)

	// Initialize Services
	authService := service.NewAuthService(userRepo)
	catalogService := service.NewCatalogService(catRepo, prodRepo)
	wsService := service.NewWebsocketService()
	botService := service.NewBotService()
	printerService := service.NewPrinterService()
	orderService := service.NewOrderService(orderRepo, prodRepo, settingsRepo, inventoryRepo, tableRepo, wsService, botService, printerService)
	fiscalService := service.NewFiscalService(fiscalRepo, settingsRepo)

	// Initialize Handlers
	authHandler := handlers.NewAuthHandler(authService, userRepo, auditRepo)
	catalogHandler := handlers.NewCatalogHandler(catalogService, auditRepo)
	orderHandler := handlers.NewOrderHandler(orderService, auditRepo)
	settingsHandler := handlers.NewSettingsHandler(settingsRepo, auditRepo)
	financeHandler := handlers.NewFinanceHandler(financeRepo, wsService, auditRepo)
	inventoryHandler := handlers.NewInventoryHandler(inventoryRepo, auditRepo)
	tableHandler := handlers.NewTableHandler(tableRepo, orderService, auditRepo)
	auditHandler := handlers.NewAuditHandler(auditRepo)
	cashierHandler := handlers.NewCashierHandler(cashierRepo, orderService, auditRepo)
	debtHandler := handlers.NewDebtHandler(debtRepo, auditRepo)
	refundHandler := handlers.NewRefundHandler(refundRepo, auditRepo)
	fiscalHandler := handlers.NewFiscalHandler(fiscalService, auditRepo)

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

		originAllowed := false
		for _, allowed := range strings.Split(allowedOrigins, ",") {
			if strings.TrimSpace(allowed) == origin {
				originAllowed = true
				break
			}
		}

		if origin != "" && originAllowed {
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
				admin.PUT("/staff/:id/default-fee", authHandler.UpdateDefaultServiceFee)
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
				staff.GET("/active-by-table/:tableID", orderHandler.GetActiveOrderByTable)
				staff.PUT("/:id/status", orderHandler.UpdateStatus)
				staff.POST("/:id/assign", orderHandler.AssignCourier)
				staff.POST("/:id/print", orderHandler.ReprintOrder)
				staff.POST("/:id/add-items", orderHandler.AddItemsToOrder)
				staff.PUT("/:id/service-fee", orderHandler.SetServiceFee)
				staff.POST("/:id/items/:item_id/cancel", orderHandler.CancelOrderItem)
				staff.POST("/:id/products/:product_id/cancel", orderHandler.CancelProductFromOrder)
				staff.POST("/:id/bulk-edit", orderHandler.BulkEditOrder)
				staff.POST("/transfer", orderHandler.TransferOrderTable)
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
			tables.PUT("/layout/batch", tableHandler.UpdateLayout)
		}

		// Finance (Admin only)
		finance := api.Group("/finance")
		finance.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("admin"))
		{
			finance.GET("/stats", financeHandler.GetStats)
			finance.GET("/expenses", financeHandler.GetExpenses)
			finance.POST("/expenses", financeHandler.CreateExpense)
			finance.POST("/close-shift", financeHandler.CloseShift)
			finance.POST("/send-real-profit", financeHandler.SendRealProfit)
			finance.GET("/waiter-salaries", financeHandler.GetWaiterSalaries)
		}

		audit := api.Group("/audit")
		audit.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("admin"))
		{
			audit.GET("/logs", auditHandler.GetLogs)
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

		// Cashier POS (Cashier + Admin)
		cashier := api.Group("/cashier")
		cashier.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("admin", "cashier"))
		{
			cashier.POST("/shift/open", cashierHandler.OpenShift)
			cashier.POST("/shift/close", cashierHandler.CloseShift)
			cashier.GET("/shift/current", cashierHandler.GetCurrentShift)
			cashier.POST("/shift/cash-operation", cashierHandler.AddCashOperation)
			cashier.GET("/shift/:id/report", cashierHandler.GetShiftReport)
			cashier.GET("/shifts", cashierHandler.GetAllShifts)
			cashier.POST("/quick-sale", cashierHandler.QuickSale)
		}

		// Debt/Nasiya Management (Admin + Cashier)
		debts := api.Group("/debts")
		debts.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("admin", "cashier", "waiter"))
		{
			debts.GET("/debtors", debtHandler.GetAllDebtors)
			debts.POST("/debtors", debtHandler.CreateDebtor)
			debts.PUT("/debtors/:id", debtHandler.UpdateDebtor)
			debts.GET("/debtors/:id/history", debtHandler.GetDebtorHistory)
			debts.POST("/debtors/:id/pay", debtHandler.PayDebt)
			debts.POST("/record", debtHandler.AddDebtRecord)
			debts.GET("/summary", debtHandler.GetDebtSummary)
		}

		// Refund Management
		refunds := api.Group("/refunds")
		refunds.Use(middleware.AuthMiddleware())
		{
			refunds.POST("/", middleware.RoleMiddleware("admin", "cashier", "waiter"), refundHandler.CreateRefund)
			refunds.GET("/pending", middleware.RoleMiddleware("admin"), refundHandler.GetPendingRefunds)
			refunds.GET("/all", middleware.RoleMiddleware("admin"), refundHandler.GetAllRefunds)
			refunds.GET("/reasons", refundHandler.GetRefundReasons)
			refunds.GET("/pending-count", middleware.RoleMiddleware("admin"), refundHandler.CountPending)
			refunds.PUT("/:id/approve", middleware.RoleMiddleware("admin"), refundHandler.ApproveRefund)
			refunds.PUT("/:id/reject", middleware.RoleMiddleware("admin"), refundHandler.RejectRefund)
			refunds.PUT("/:id/money-returned", middleware.RoleMiddleware("admin"), refundHandler.MarkMoneyReturned)
				refunds.GET("/order/:id", middleware.RoleMiddleware("admin", "cashier", "waiter"), refundHandler.GetRefundsByOrder)
		}

		// Fiscal Receipts (Admin only)
		fiscal := api.Group("/fiscal")
		fiscal.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("admin", "cashier"))
		{
			fiscal.GET("/receipt/:order_id", fiscalHandler.GetReceiptByOrder)
			fiscal.GET("/receipts", fiscalHandler.GetAllReceipts)
			fiscal.GET("/settings", fiscalHandler.GetSettings)
			fiscal.PUT("/settings", fiscalHandler.UpdateSettings)
			fiscal.POST("/receipt/:order_id/resend", fiscalHandler.ResendToOFD)
			fiscal.GET("/stats", fiscalHandler.GetStats)
		}

		// Printer Control (Staff Only)
		printer := api.Group("/printer")
		printer.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("admin", "cook", "courier", "cashier"))
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

		// Public Cafe Config (No Auth Required)
		api.GET("/config", func(c *gin.Context) {
			cafeName := os.Getenv("CAFE_NAME")
			if cafeName == "" {
				cafeName = "Kafe"
			}
			cafeFullName := os.Getenv("CAFE_FULL_NAME")
			if cafeFullName == "" {
				cafeFullName = cafeName
			}
			cafeWebsite := os.Getenv("CAFE_WEBSITE")
			cafeAddress := os.Getenv("CAFE_ADDRESS")
			cafePhone := os.Getenv("CAFE_PHONE")

			c.JSON(http.StatusOK, gin.H{
				"cafe_name":      cafeName,
				"cafe_full_name": cafeFullName,
				"cafe_website":   cafeWebsite,
				"cafe_address":   cafeAddress,
				"cafe_phone":     cafePhone,
			})
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
