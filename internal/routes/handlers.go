package routes

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Over-knight/Lujay-assesment/internal/storage"
)

// HealthCheck handles health check requests
// Returns a simple status indicating the service is running
// Also checks MongoDB connection status
func HealthCheck(db *storage.MongoDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check MongoDB connection
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		dbStatus := "ok"
		if err := db.Ping(ctx); err != nil {
			dbStatus = "error"
		}

		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"message":  "Service is running",
			"database": dbStatus,
		})
	}
}

// WelcomeHandler handles welcome requests
// Returns a welcome message for the API
func WelcomeHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Welcome to LUJAY Assessment API",
		"version": "1.0.0",
		"endpoints": gin.H{
			"auth":         "/api/v1/auth",
			"vehicles":     "/api/v1/vehicles",
			"inspections":  "/api/v1/inspections",
			"transactions": "/api/v1/transactions",
			"health":       "/health",
		},
	})
}
