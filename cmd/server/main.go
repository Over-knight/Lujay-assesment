package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Over-knight/Lujay-assesment/internal/auth"
	"github.com/Over-knight/Lujay-assesment/internal/cache"
	"github.com/Over-knight/Lujay-assesment/internal/config"
	"github.com/Over-knight/Lujay-assesment/internal/handlers"
	"github.com/Over-knight/Lujay-assesment/internal/routes"
	"github.com/Over-knight/Lujay-assesment/internal/service"
	"github.com/Over-knight/Lujay-assesment/internal/storage"
	"github.com/Over-knight/Lujay-assesment/internal/upload"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load application configuration
	cfg := config.Load()

	// Parse JWT expiration duration
	jwtExpiration, err := time.ParseDuration(cfg.JWT.Expiration)
	if err != nil {
		log.Fatalf("Invalid JWT expiration format: %v", err)
	}

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(cfg.JWT.Secret, jwtExpiration)

	// Initialize MongoDB connection
	mongoDB, err := storage.NewMongoDB(storage.MongoConfig{
		URI:      cfg.MongoDB.URI,
		Database: cfg.MongoDB.Database,
		Timeout:  10 * time.Second,
	})
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Ensure MongoDB connection is closed on shutdown
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := mongoDB.Close(ctx); err != nil {
			log.Printf("Error closing MongoDB connection: %v", err)
		}
	}()

	// Initialize Redis cache
	redisCache, err := cache.NewRedisCache(cache.RedisConfig{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v. Continuing without cache.", err)
		redisCache = nil // Set to nil to gracefully handle absence
	} else {
		log.Println("Successfully connected to Redis")
		// Ensure Redis connection is closed on shutdown
		defer func() {
			if err := redisCache.Close(); err != nil {
				log.Printf("Error closing Redis connection: %v", err)
			}
		}()
	}

	// Initialize Cloudinary uploader
	var cloudinaryUploader *upload.CloudinaryUploader
	if cfg.Cloudinary.CloudName != "" && cfg.Cloudinary.APIKey != "" && cfg.Cloudinary.APISecret != "" {
		cloudinaryUploader, err = upload.NewCloudinaryUploader(upload.CloudinaryConfig{
			CloudName: cfg.Cloudinary.CloudName,
			APIKey:    cfg.Cloudinary.APIKey,
			APISecret: cfg.Cloudinary.APISecret,
			Folder:    cfg.Cloudinary.Folder,
		})
		if err != nil {
			log.Printf("Warning: Failed to initialize Cloudinary: %v. File upload will not be available.", err)
			cloudinaryUploader = nil
		} else {
			log.Println("Successfully initialized Cloudinary uploader")
		}
	} else {
		log.Println("Cloudinary credentials not configured. File upload will not be available.")
	}

	// Initialize services
	userService := service.NewUserService(mongoDB.Collection("users"), jwtManager)
	vehicleService := service.NewVehicleService(mongoDB.Collection("vehicles"))
	inspectionService := service.NewInspectionService(mongoDB.Database)
	transactionService := service.NewTransactionService(mongoDB.Database)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userService)
	vehicleHandler := handlers.NewVehicleHandler(vehicleService)
	inspectionHandler := handlers.NewInspectionHandler(inspectionService)
	transactionHandler := handlers.NewTransactionHandler(transactionService)
	uploadHandler := handlers.NewUploadHandler(cloudinaryUploader, vehicleService)

	// Initialize Gin router with default middleware (logger and recovery)
	router := gin.Default()

	// Set up routes with Redis cache
	routes.SetupRoutes(router, mongoDB, redisCache, authHandler, vehicleHandler, inspectionHandler, transactionHandler, uploadHandler, jwtManager)

	// Create server with graceful shutdown support
	srv := setupServer(router, cfg.Server.Port)

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// setupServer creates and configures the HTTP server
func setupServer(router *gin.Engine, port string) *http.Server {
	return &http.Server{
		Addr:           ":" + port,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}
}
