package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/Over-knight/Lujay-assesment/internal/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// RBACMiddleware creates a middleware that checks user roles
// userCollection: MongoDB users collection to fetch user data
// allowedRoles: List of roles allowed to access the route
// Returns a Gin middleware handler function
func RBACMiddleware(userCollection *mongo.Collection, allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (must be set by AuthMiddleware first)
		userID := GetUserID(c)
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User not authenticated",
			})
			c.Abort()
			return
		}

		// Convert user ID to ObjectID
		objectID, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid user ID",
			})
			c.Abort()
			return
		}

		// Fetch user from database
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		var user models.User
		err = userCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "User not found",
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to verify user",
				})
			}
			c.Abort()
			return
		}

		// Check if user role is in allowed roles
		roleAllowed := false
		for _, role := range allowedRoles {
			if user.Role == role {
				roleAllowed = true
				break
			}
		}

		if !roleAllowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Insufficient permissions",
			})
			c.Abort()
			return
		}

		// Store user role in context for use in handlers
		c.Set("userRole", user.Role)

		// Continue to next handler
		c.Next()
	}
}

// GetUserRole retrieves the user role from the Gin context
// Should be called after RBACMiddleware has run
// Returns the role or an empty string if not found
func GetUserRole(c *gin.Context) string {
	role, exists := c.Get("userRole")
	if !exists {
		return ""
	}
	return role.(string)
}

// RequireAdmin middleware that only allows admin users
func RequireAdmin(userCollection *mongo.Collection) gin.HandlerFunc {
	return RBACMiddleware(userCollection, models.RoleAdmin)
}

// RequireAdminOrDealer middleware that allows admin or dealer users
func RequireAdminOrDealer(userCollection *mongo.Collection) gin.HandlerFunc {
	return RBACMiddleware(userCollection, models.RoleAdmin, models.RoleDealer)
}
