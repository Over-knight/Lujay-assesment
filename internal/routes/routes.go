package routes

import (
	"time"

	"github.com/Over-knight/Lujay-assesment/internal/auth"
	"github.com/Over-knight/Lujay-assesment/internal/cache"
	"github.com/Over-knight/Lujay-assesment/internal/handlers"
	"github.com/Over-knight/Lujay-assesment/internal/middleware"
	"github.com/Over-knight/Lujay-assesment/internal/storage"
	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all application routes
func SetupRoutes(
	router *gin.Engine,
	db *storage.MongoDB,
	redisCache *cache.RedisCache,
	authHandler *handlers.AuthHandler,
	vehicleHandler *handlers.VehicleHandler,
	inspectionHandler *handlers.InspectionHandler,
	transactionHandler *handlers.TransactionHandler,
	uploadHandler *handlers.UploadHandler,
	jwtManager *auth.JWTManager,
) {
	// Health check endpoint
	router.GET("/health", HealthCheck(db))

	// Apply global cache middleware for Redis-based rate limiting (if Redis is available)
	if redisCache != nil {
		router.Use(middleware.RateLimitMiddleware(redisCache, 100, time.Minute))
	}

	// API v1 group
	v1 := router.Group("/api/v1")
	{
		// Public routes (no authentication required)
		v1.GET("/", WelcomeHandler)

		// Auth routes
		setupAuthRoutes(v1, authHandler, jwtManager)

		// Vehicle routes (pass uploadHandler)
		setupVehicleRoutes(v1, vehicleHandler, inspectionHandler, transactionHandler, uploadHandler, db, redisCache, jwtManager)

		// Inspection routes
		setupInspectionRoutes(v1, inspectionHandler, db, redisCache, jwtManager)

		// Transaction routes
		setupTransactionRoutes(v1, transactionHandler, db, jwtManager)
	}
}

// setupAuthRoutes configures authentication routes
func setupAuthRoutes(v1 *gin.RouterGroup, authHandler *handlers.AuthHandler, jwtManager *auth.JWTManager) {
	authRoutes := v1.Group("/auth")
	{
		authRoutes.POST("/register", authHandler.Register)
		authRoutes.POST("/login", authHandler.Login)

		// Protected routes (authentication required)
		authRoutes.GET("/profile", middleware.AuthMiddleware(jwtManager), authHandler.GetProfile)
	}
}

// setupVehicleRoutes configures vehicle routes
func setupVehicleRoutes(
	v1 *gin.RouterGroup,
	vehicleHandler *handlers.VehicleHandler,
	inspectionHandler *handlers.InspectionHandler,
	transactionHandler *handlers.TransactionHandler,
	uploadHandler *handlers.UploadHandler,
	db *storage.MongoDB,
	redisCache *cache.RedisCache,
	jwtManager *auth.JWTManager,
) {
	vehicleRoutes := v1.Group("/vehicles")
	{
		// Apply cache middleware for public GET routes (5 minute cache)
		if redisCache != nil {
			vehicleRoutes.GET("", middleware.CacheMiddleware(redisCache, 5*time.Minute), vehicleHandler.ListVehicles)
			vehicleRoutes.GET("/:id", middleware.CacheMiddleware(redisCache, 5*time.Minute), vehicleHandler.GetVehicle)
		} else {
			vehicleRoutes.GET("", vehicleHandler.ListVehicles)
			vehicleRoutes.GET("/:id", vehicleHandler.GetVehicle)
		}

		// Protected routes (authentication required)
		vehicleRoutes.POST("", middleware.AuthMiddleware(jwtManager), vehicleHandler.CreateVehicle)
		vehicleRoutes.GET("/my", middleware.AuthMiddleware(jwtManager), vehicleHandler.GetMyVehicles)
		vehicleRoutes.PUT("/:id", middleware.AuthMiddleware(jwtManager), vehicleHandler.UpdateVehicle)
		vehicleRoutes.DELETE("/:id", middleware.AuthMiddleware(jwtManager), middleware.RequireAdminOrDealer(db.Collection("users")), vehicleHandler.DeleteVehicle)

		// Inspection routes for specific vehicle
		vehicleRoutes.GET("/:id/inspections", middleware.AuthMiddleware(jwtManager), inspectionHandler.GetInspectionsByVehicle)

		// Transaction routes for specific vehicle
		vehicleRoutes.GET("/:id/transactions", middleware.AuthMiddleware(jwtManager), transactionHandler.GetTransactionsByVehicle)

		// Image upload routes
		if uploadHandler != nil {
			vehicleRoutes.POST("/:id/images", middleware.AuthMiddleware(jwtManager), uploadHandler.UploadVehicleImages)
			vehicleRoutes.DELETE("/:id/images/:publicId", middleware.AuthMiddleware(jwtManager), uploadHandler.DeleteVehicleImage)
			vehicleRoutes.PUT("/:id/images/:publicId/primary", middleware.AuthMiddleware(jwtManager), uploadHandler.SetPrimaryImage)
		}

		// Apply cache buster for modifying operations
		if redisCache != nil {
			vehicleRoutes.Use(middleware.CacheBuster(redisCache))
		}
	}
}

// setupInspectionRoutes configures inspection routes
func setupInspectionRoutes(v1 *gin.RouterGroup, inspectionHandler *handlers.InspectionHandler, db *storage.MongoDB, redisCache *cache.RedisCache, jwtManager *auth.JWTManager) {
	inspectionRoutes := v1.Group("/inspections")
	{
		// Public routes with cache
		if redisCache != nil {
			inspectionRoutes.GET("", middleware.CacheMiddleware(redisCache, 3*time.Minute), inspectionHandler.ListInspections)
			inspectionRoutes.GET("/:id", middleware.CacheMiddleware(redisCache, 3*time.Minute), inspectionHandler.GetInspection)
		} else {
			inspectionRoutes.GET("", inspectionHandler.ListInspections)
			inspectionRoutes.GET("/:id", inspectionHandler.GetInspection)
		}

		// Protected routes (authentication required)
		inspectionRoutes.POST("", middleware.AuthMiddleware(jwtManager), inspectionHandler.CreateInspection)
		inspectionRoutes.GET("/my", middleware.AuthMiddleware(jwtManager), inspectionHandler.GetMyInspections)
		inspectionRoutes.PUT("/:id", middleware.AuthMiddleware(jwtManager), inspectionHandler.UpdateInspection)
		inspectionRoutes.POST("/:id/complete", middleware.AuthMiddleware(jwtManager), inspectionHandler.CompleteInspection)
		inspectionRoutes.POST("/:id/cancel", middleware.AuthMiddleware(jwtManager), inspectionHandler.CancelInspection)
		inspectionRoutes.DELETE("/:id", middleware.AuthMiddleware(jwtManager), middleware.RequireAdmin(db.Collection("users")), inspectionHandler.DeleteInspection)

		// Apply cache buster for modifying operations
		if redisCache != nil {
			inspectionRoutes.Use(middleware.CacheBuster(redisCache))
		}
	}
}

// setupTransactionRoutes configures transaction routes
func setupTransactionRoutes(v1 *gin.RouterGroup, transactionHandler *handlers.TransactionHandler, db *storage.MongoDB, jwtManager *auth.JWTManager) {
	transactionRoutes := v1.Group("/transactions")
	{
		// Public routes (admin only)
		transactionRoutes.GET("", middleware.AuthMiddleware(jwtManager), middleware.RequireAdmin(db.Collection("users")), transactionHandler.ListTransactions)
		transactionRoutes.GET("/:id", middleware.AuthMiddleware(jwtManager), transactionHandler.GetTransaction)

		// Protected routes (authentication required)
		transactionRoutes.POST("", middleware.AuthMiddleware(jwtManager), transactionHandler.CreateTransaction)
		transactionRoutes.GET("/my", middleware.AuthMiddleware(jwtManager), transactionHandler.GetMyTransactions)
		transactionRoutes.PUT("/:id", middleware.AuthMiddleware(jwtManager), transactionHandler.UpdateTransaction)
		transactionRoutes.POST("/:id/complete", middleware.AuthMiddleware(jwtManager), transactionHandler.CompleteTransaction)
		transactionRoutes.POST("/:id/cancel", middleware.AuthMiddleware(jwtManager), transactionHandler.CancelTransaction)
	}
}
