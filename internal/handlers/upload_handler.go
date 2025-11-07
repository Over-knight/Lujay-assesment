package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Over-knight/Lujay-assesment/internal/errors"
	"github.com/Over-knight/Lujay-assesment/internal/middleware"
	"github.com/Over-knight/Lujay-assesment/internal/models"
	"github.com/Over-knight/Lujay-assesment/internal/service"
	"github.com/Over-knight/Lujay-assesment/internal/upload"
)

// UploadHandler handles file upload operations
type UploadHandler struct {
	uploader       *upload.CloudinaryUploader
	vehicleService *service.VehicleService
}

// NewUploadHandler creates a new upload handler
func NewUploadHandler(uploader *upload.CloudinaryUploader, vehicleService *service.VehicleService) *UploadHandler {
	return &UploadHandler{
		uploader:       uploader,
		vehicleService: vehicleService,
	}
}

// UploadVehicleImagesRequest represents the upload request
type UploadVehicleImagesRequest struct {
	IsPrimary bool `form:"is_primary"`
}

// UploadVehicleImages handles uploading images for a vehicle
// @Summary Upload vehicle images
// @Description Upload one or multiple images for a vehicle (max 10 images per vehicle, 10MB per file)
// @Tags uploads
// @Accept multipart/form-data
// @Produce json
// @Param vehicleId path string true "Vehicle ID"
// @Param images formData file true "Image files (jpg, jpeg, png, gif, webp)"
// @Param is_primary formData bool false "Set first image as primary"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/vehicles/{vehicleId}/images [post]
func (h *UploadHandler) UploadVehicleImages(c *gin.Context) {
	// Get vehicle ID
	vehicleID := c.Param("id")
	_, err := primitive.ObjectIDFromHex(vehicleID)
	if err != nil {
		errors.HandleError(c, errors.NewValidationError("invalid vehicle ID"))
		return
	}

	// Get user ID from context
	userIDStr := middleware.GetUserID(c)
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		errors.HandleError(c, errors.ErrUnauthorized)
		return
	}

	// Get vehicle to verify ownership
	vehicle, err := h.vehicleService.GetVehicleByID(c.Request.Context(), vehicleID)
	if err != nil {
		errors.HandleError(c, errors.NewNotFoundError("vehicle not found"))
		return
	}

	// Verify user is the owner
	if vehicle.OwnerID != userID {
		errors.HandleError(c, errors.ErrInsufficientPermissions)
		return
	}

	// Parse multipart form
	form, err := c.MultipartForm()
	if err != nil {
		errors.HandleError(c, errors.NewValidationError("failed to parse form"))
		return
	}

	// Get uploaded files
	files := form.File["images"]
	if len(files) == 0 {
		errors.HandleError(c, errors.NewValidationError("no images provided"))
		return
	}

	// Check max images limit (10 images per vehicle)
	maxImages := 10
	currentImagesCount := len(vehicle.Images)
	if currentImagesCount+len(files) > maxImages {
		errors.HandleError(c, errors.NewValidationError(
			fmt.Sprintf("maximum %d images allowed per vehicle (current: %d)", maxImages, currentImagesCount),
		))
		return
	}

	// Validate and upload each file
	uploadedImages := make([]models.VehicleImage, 0, len(files))
	uploadedPublicIDs := make([]string, 0, len(files))

	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	for i, fileHeader := range files {
		// Open file
		file, err := fileHeader.Open()
		if err != nil {
			// Rollback: delete previously uploaded images
			h.rollbackUploads(ctx, uploadedPublicIDs)
			errors.HandleError(c, errors.NewValidationError(fmt.Sprintf("failed to open file: %s", fileHeader.Filename)))
			return
		}
		defer file.Close()

		// Validate file
		if err := upload.ValidateImageFile(file, fileHeader); err != nil {
			h.rollbackUploads(ctx, uploadedPublicIDs)
			errors.HandleError(c, errors.NewValidationError(fmt.Sprintf("%s: %s", fileHeader.Filename, err.Error())))
			return
		}

		// Upload to Cloudinary
		result, err := h.uploader.UploadImage(ctx, file, fileHeader.Filename, fmt.Sprintf("vehicle_%s", vehicleID))
		if err != nil {
			h.rollbackUploads(ctx, uploadedPublicIDs)
			errors.HandleError(c, errors.NewDatabaseError("failed to upload image"))
			return
		}

		// Create vehicle image
		isPrimary := i == 0 && len(vehicle.Images) == 0 // First image is primary if no images exist
		vehicleImage := models.VehicleImage{
			URL:       result.URL,
			PublicID:  result.PublicID,
			IsPrimary: isPrimary,
		}

		uploadedImages = append(uploadedImages, vehicleImage)
		uploadedPublicIDs = append(uploadedPublicIDs, result.PublicID)
	}

	// Update vehicle with new images
	vehicle.Images = append(vehicle.Images, uploadedImages...)

	updateReq := models.UpdateVehicleRequest{
		Images: vehicle.Images,
	}

	updatedVehicle, err := h.vehicleService.UpdateVehicle(c.Request.Context(), vehicleID, userIDStr, updateReq)
	if err != nil {
		// Rollback uploads
		h.rollbackUploads(ctx, uploadedPublicIDs)
		errors.HandleError(c, errors.NewDatabaseError("failed to update vehicle"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "images uploaded successfully",
		"images_added":    len(uploadedImages),
		"total_images":    len(updatedVehicle.Images),
		"uploaded_images": uploadedImages,
	})
}

// DeleteVehicleImage handles deleting a vehicle image
// @Summary Delete vehicle image
// @Description Delete a specific image from a vehicle
// @Tags uploads
// @Produce json
// @Param vehicleId path string true "Vehicle ID"
// @Param publicId path string true "Image Public ID"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/vehicles/{vehicleId}/images/{publicId} [delete]
func (h *UploadHandler) DeleteVehicleImage(c *gin.Context) {
	// Get vehicle ID
	vehicleID := c.Param("id")
	_, err := primitive.ObjectIDFromHex(vehicleID)
	if err != nil {
		errors.HandleError(c, errors.NewValidationError("invalid vehicle ID"))
		return
	}

	// Get public ID
	publicID := c.Param("publicId")
	if publicID == "" {
		errors.HandleError(c, errors.NewValidationError("public ID is required"))
		return
	}

	// Get user ID from context
	userIDStr := middleware.GetUserID(c)
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		errors.HandleError(c, errors.ErrUnauthorized)
		return
	}

	// Get vehicle to verify ownership
	vehicle, err := h.vehicleService.GetVehicleByID(c.Request.Context(), vehicleID)
	if err != nil {
		errors.HandleError(c, errors.NewNotFoundError("vehicle not found"))
		return
	}

	// Verify user is the owner
	if vehicle.OwnerID != userID {
		errors.HandleError(c, errors.ErrInsufficientPermissions)
		return
	}

	// Find the image in vehicle images
	imageIndex := -1
	for i, img := range vehicle.Images {
		if img.PublicID == publicID {
			imageIndex = i
			break
		}
	}

	if imageIndex == -1 {
		errors.HandleError(c, errors.NewNotFoundError("image not found"))
		return
	}

	// Delete from Cloudinary
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	if err := h.uploader.DeleteImage(ctx, publicID); err != nil {
		errors.HandleError(c, errors.NewDatabaseError("failed to delete image from storage"))
		return
	}

	// Remove image from vehicle
	vehicle.Images = append(vehicle.Images[:imageIndex], vehicle.Images[imageIndex+1:]...)

	// If deleted image was primary and there are other images, make the first one primary
	if len(vehicle.Images) > 0 {
		hasPrimary := false
		for _, img := range vehicle.Images {
			if img.IsPrimary {
				hasPrimary = true
				break
			}
		}
		if !hasPrimary {
			vehicle.Images[0].IsPrimary = true
		}
	}

	// Update vehicle
	updateReq := models.UpdateVehicleRequest{
		Images: vehicle.Images,
	}

	_, err = h.vehicleService.UpdateVehicle(c.Request.Context(), vehicleID, userIDStr, updateReq)
	if err != nil {
		errors.HandleError(c, errors.NewDatabaseError("failed to update vehicle"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":          "image deleted successfully",
		"remaining_images": len(vehicle.Images),
	})
}

// SetPrimaryImage sets an image as the primary image for a vehicle
// @Summary Set primary vehicle image
// @Description Set a specific image as the primary image for display
// @Tags uploads
// @Produce json
// @Param vehicleId path string true "Vehicle ID"
// @Param publicId path string true "Image Public ID"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/vehicles/{vehicleId}/images/{publicId}/primary [put]
func (h *UploadHandler) SetPrimaryImage(c *gin.Context) {
	// Get vehicle ID
	vehicleID := c.Param("id")
	_, err := primitive.ObjectIDFromHex(vehicleID)
	if err != nil {
		errors.HandleError(c, errors.NewValidationError("invalid vehicle ID"))
		return
	}

	// Get public ID
	publicID := c.Param("publicId")
	if publicID == "" {
		errors.HandleError(c, errors.NewValidationError("public ID is required"))
		return
	}

	// Get user ID from context
	userIDStr := middleware.GetUserID(c)
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		errors.HandleError(c, errors.ErrUnauthorized)
		return
	}

	// Get vehicle to verify ownership
	vehicle, err := h.vehicleService.GetVehicleByID(c.Request.Context(), vehicleID)
	if err != nil {
		errors.HandleError(c, errors.NewNotFoundError("vehicle not found"))
		return
	}

	// Verify user is the owner
	if vehicle.OwnerID != userID {
		errors.HandleError(c, errors.ErrInsufficientPermissions)
		return
	}

	// Find the image and update primary status
	found := false
	for i := range vehicle.Images {
		if vehicle.Images[i].PublicID == publicID {
			vehicle.Images[i].IsPrimary = true
			found = true
		} else {
			vehicle.Images[i].IsPrimary = false
		}
	}

	if !found {
		errors.HandleError(c, errors.NewNotFoundError("image not found"))
		return
	}

	// Update vehicle
	updateReq := models.UpdateVehicleRequest{
		Images: vehicle.Images,
	}

	_, err = h.vehicleService.UpdateVehicle(c.Request.Context(), vehicleID, userIDStr, updateReq)
	if err != nil {
		errors.HandleError(c, errors.NewDatabaseError("failed to update vehicle"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "primary image updated successfully",
	})
}

// rollbackUploads deletes uploaded images from Cloudinary
func (h *UploadHandler) rollbackUploads(ctx context.Context, publicIDs []string) {
	for _, publicID := range publicIDs {
		_ = h.uploader.DeleteImage(ctx, publicID)
	}
}
