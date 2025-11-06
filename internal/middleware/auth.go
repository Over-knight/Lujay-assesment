package middleware

import (
	"net/http"
	"strings"

	"github.com/Over-knight/Lujay-assesment/internal/auth"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware creates a middleware that validates JWT tokens
// jwtManager: The JWT manager instance to use for token validation
// Returns a Gin middleware handler function
func AuthMiddleware(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// Check for Bearer token format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format. Expected: Bearer <token>",
			})
			c.Abort()
			return
		}

		// Extract token string
		tokenString := parts[1]

		// Validate token
		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Store user information in context for use in handlers
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)

		// Continue to next handler
		c.Next()
	}
}

// GetUserID retrieves the user ID from the Gin context
// Should be called after AuthMiddleware has run
// Returns the user ID or an empty string if not found
func GetUserID(c *gin.Context) string {
	userID, exists := c.Get("userID")
	if !exists {
		return ""
	}
	return userID.(string)
}

// GetUserEmail retrieves the user email from the Gin context
// Should be called after AuthMiddleware has run
// Returns the email or an empty string if not found
func GetUserEmail(c *gin.Context) string {
	email, exists := c.Get("email")
	if !exists {
		return ""
	}
	return email.(string)
}
