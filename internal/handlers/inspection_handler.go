package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Over-knight/Lujay-assesment/internal/middleware"
	"github.com/Over-knight/Lujay-assesment/internal/models"
	"github.com/Over-knight/Lujay-assesment/internal/service"
)

// InspectionHandler handles inspection-related HTTP requests
type InspectionHandler struct {
	service *service.InspectionService
}

// NewInspectionHandler creates a new inspection handler
func NewInspectionHandler(service *service.InspectionService) *InspectionHandler {
	return &InspectionHandler{
		service: service,
	}
}

// CreateInspection handles POST /inspections
func (h *InspectionHandler) CreateInspection(c *gin.Context) {
	var req models.CreateInspectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get inspector ID from context
	userIDStr := middleware.GetUserID(c)
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	inspectorID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	inspection, err := h.service.CreateInspection(c.Request.Context(), &req, inspectorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, inspection)
}

// GetInspection handles GET /inspections/:id
func (h *InspectionHandler) GetInspection(c *gin.Context) {
	id := c.Param("id")

	inspection, err := h.service.GetInspectionByID(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "inspection not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, inspection)
}

// GetInspectionsByVehicle handles GET /vehicles/:id/inspections
func (h *InspectionHandler) GetInspectionsByVehicle(c *gin.Context) {
	vehicleID := c.Param("id")

	inspections, err := h.service.GetInspectionsByVehicle(c.Request.Context(), vehicleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"inspections": inspections,
		"count":       len(inspections),
	})
}

// GetMyInspections handles GET /inspections/my
func (h *InspectionHandler) GetMyInspections(c *gin.Context) {
	userIDStr := middleware.GetUserID(c)
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	inspectorID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	inspections, err := h.service.GetInspectionsByInspector(c.Request.Context(), inspectorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"inspections": inspections,
		"count":       len(inspections),
	})
}

// UpdateInspection handles PUT /inspections/:id
func (h *InspectionHandler) UpdateInspection(c *gin.Context) {
	id := c.Param("id")

	var req models.UpdateInspectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	inspection, err := h.service.UpdateInspection(c.Request.Context(), id, &req)
	if err != nil {
		if err.Error() == "inspection not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, inspection)
}

// CompleteInspection handles POST /inspections/:id/complete
func (h *InspectionHandler) CompleteInspection(c *gin.Context) {
	id := c.Param("id")

	var req models.CompleteInspectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	inspection, err := h.service.CompleteInspection(c.Request.Context(), id, &req)
	if err != nil {
		if err.Error() == "inspection not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, inspection)
}

// CancelInspection handles POST /inspections/:id/cancel
func (h *InspectionHandler) CancelInspection(c *gin.Context) {
	id := c.Param("id")

	type CancelRequest struct {
		Notes string `json:"notes"`
	}

	var req CancelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	inspection, err := h.service.CancelInspection(c.Request.Context(), id, req.Notes)
	if err != nil {
		if err.Error() == "inspection not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, inspection)
}

// DeleteInspection handles DELETE /inspections/:id
func (h *InspectionHandler) DeleteInspection(c *gin.Context) {
	id := c.Param("id")

	err := h.service.DeleteInspection(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "inspection not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Inspection deleted successfully"})
}

// ListInspections handles GET /inspections
func (h *InspectionHandler) ListInspections(c *gin.Context) {
	// Parse query parameters
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	inspections, totalCount, err := h.service.ListInspections(c.Request.Context(), status, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalPages := (int(totalCount) + limit - 1) / limit

	c.JSON(http.StatusOK, gin.H{
		"inspections": inspections,
		"totalCount":  totalCount,
		"page":        page,
		"limit":       limit,
		"totalPages":  totalPages,
	})
}
