package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Over-knight/Lujay-assesment/internal/middleware"
	"github.com/Over-knight/Lujay-assesment/internal/models"
	"github.com/Over-knight/Lujay-assesment/internal/service"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	userService *service.UserService
}

// NewAuthHandler creates a new authentication handler
// userService: The user service for authentication operations
func NewAuthHandler(userService *service.UserService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
	}
}

// Register handles user registration requests
// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	// Parse request body
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request payload",
		})
		return
	}

	// Register user
	authResponse, err := h.userService.Register(c.Request.Context(), req)
	if err != nil {
		// Check for specific error types
		if err.Error() == "user with this email already exists" {
			c.JSON(http.StatusConflict, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to register user",
		})
		return
	}

	// Return success response
	c.JSON(http.StatusCreated, authResponse)
}

// Login handles user login requests
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	// Parse request body
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request payload",
		})
		return
	}

	// Authenticate user
	authResponse, err := h.userService.Login(c.Request.Context(), req)
	if err != nil {
		// Return generic error for security
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, authResponse)
}

// GetProfile handles requests to get the current user's profile
// GET /api/v1/auth/profile
// Requires authentication middleware
func (h *AuthHandler) GetProfile(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Get user from database
	user, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	// Return user data
	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}
