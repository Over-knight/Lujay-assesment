package handlers

import (
	"net/http"
	"strconv"

	"github.com/Over-knight/Lujay-assesment/internal/middleware"
	"github.com/Over-knight/Lujay-assesment/internal/models"
	"github.com/Over-knight/Lujay-assesment/internal/service"
	"github.com/gin-gonic/gin"
)

// VehicleHandler handles vehicle-related HTTP requests
type VehicleHandler struct {
	vehicleService *service.VehicleService
}

// NewVehicleHandler creates a new vehicle handler
// vehicleService: The vehicle service for vehicle operations
func NewVehicleHandler(vehicleService *service.VehicleService) *VehicleHandler {
	return &VehicleHandler{
		vehicleService: vehicleService,
	}
}

// CreateVehicle handles vehicle creation requests
// POST /api/v1/vehicles
// Requires authentication
func (h *VehicleHandler) CreateVehicle(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Parse request body
	var req models.CreateVehicleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request payload",
		})
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Create vehicle
	vehicle, err := h.vehicleService.CreateVehicle(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create vehicle",
		})
		return
	}

	// Return success response
	c.JSON(http.StatusCreated, gin.H{
		"message": "Vehicle created successfully",
		"vehicle": vehicle,
	})
}

// ListVehicles handles requests to list vehicles with pagination, filtering, and sorting
// GET /api/v1/vehicles
func (h *VehicleHandler) ListVehicles(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)
	minPrice, _ := strconv.ParseFloat(c.Query("minPrice"), 64)
	maxPrice, _ := strconv.ParseFloat(c.Query("maxPrice"), 64)
	minYear, _ := strconv.Atoi(c.Query("minYear"))
	maxYear, _ := strconv.Atoi(c.Query("maxYear"))

	// Build query
	query := service.VehicleListQuery{
		Page:      page,
		Limit:     limit,
		Make:      c.Query("make"),
		Model:     c.Query("model"),
		MinPrice:  minPrice,
		MaxPrice:  maxPrice,
		MinYear:   minYear,
		MaxYear:   maxYear,
		Status:    c.DefaultQuery("status", "active"),
		SortBy:    c.DefaultQuery("sortBy", "createdAt"),
		SortOrder: c.DefaultQuery("sortOrder", "desc"),
	}

	// Get vehicles from database
	response, err := h.vehicleService.ListVehicles(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve vehicles",
		})
		return
	}

	// Return vehicles
	c.JSON(http.StatusOK, response)
}

// GetVehicle handles requests to get a vehicle by ID
// GET /api/v1/vehicles/:id
func (h *VehicleHandler) GetVehicle(c *gin.Context) {
	// Get vehicle ID from URL parameter
	vehicleID := c.Param("id")
	if vehicleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Vehicle ID is required",
		})
		return
	}

	// Get vehicle from database
	vehicle, err := h.vehicleService.GetVehicleByID(c.Request.Context(), vehicleID)
	if err != nil {
		if err.Error() == "vehicle not found" || err.Error() == "invalid vehicle ID" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Vehicle not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve vehicle",
		})
		return
	}

	// Return vehicle data
	c.JSON(http.StatusOK, gin.H{
		"vehicle": vehicle,
	})
}

// GetMyVehicles handles requests to get all vehicles owned by the authenticated user
// GET /api/v1/vehicles/my
// Requires authentication
func (h *VehicleHandler) GetMyVehicles(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Get vehicles from database
	vehicles, err := h.vehicleService.GetVehiclesByOwner(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve vehicles",
		})
		return
	}

	// Return vehicles
	c.JSON(http.StatusOK, gin.H{
		"vehicles": vehicles,
		"count":    len(vehicles),
	})
}

// UpdateVehicle handles vehicle update requests
// PUT /api/v1/vehicles/:id
// Requires authentication
func (h *VehicleHandler) UpdateVehicle(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Get vehicle ID from URL parameter
	vehicleID := c.Param("id")
	if vehicleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Vehicle ID is required",
		})
		return
	}

	// Parse request body
	var req models.UpdateVehicleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request payload",
		})
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Update vehicle
	vehicle, err := h.vehicleService.UpdateVehicle(c.Request.Context(), vehicleID, userID, req)
	if err != nil {
		if err.Error() == "vehicle not found or unauthorized" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update vehicle",
		})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"message": "Vehicle updated successfully",
		"vehicle": vehicle,
	})
}

// DeleteVehicle handles vehicle deletion requests (soft delete)
// DELETE /api/v1/vehicles/:id
// Requires authentication
func (h *VehicleHandler) DeleteVehicle(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Get vehicle ID from URL parameter
	vehicleID := c.Param("id")
	if vehicleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Vehicle ID is required",
		})
		return
	}

	// Delete vehicle
	err := h.vehicleService.DeleteVehicle(c.Request.Context(), vehicleID, userID)
	if err != nil {
		if err.Error() == "vehicle not found or unauthorized" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete vehicle",
		})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"message": "Vehicle deleted successfully",
	})
}
