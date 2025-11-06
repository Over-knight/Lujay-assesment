package upload

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

// CloudinaryUploader handles file uploads to Cloudinary
type CloudinaryUploader struct {
	cld    *cloudinary.Cloudinary
	folder string
}

// CloudinaryConfig holds Cloudinary configuration
type CloudinaryConfig struct {
	CloudName string
	APIKey    string
	APISecret string
	Folder    string // Base folder for uploads
}

// UploadResult contains information about uploaded file
type UploadResult struct {
	URL      string `json:"url"`
	PublicID string `json:"public_id"`
	Format   string `json:"format"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	Bytes    int    `json:"bytes"`
}

// NewCloudinaryUploader creates a new Cloudinary uploader
func NewCloudinaryUploader(config CloudinaryConfig) (*CloudinaryUploader, error) {
	cld, err := cloudinary.NewFromParams(config.CloudName, config.APIKey, config.APISecret)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Cloudinary: %w", err)
	}

	return &CloudinaryUploader{
		cld:    cld,
		folder: config.Folder,
	}, nil
}

// UploadImage uploads an image file to Cloudinary
func (u *CloudinaryUploader) UploadImage(ctx context.Context, file multipart.File, filename, folder string) (*UploadResult, error) {
	// Validate file extension
	if !isValidImageFormat(filename) {
		return nil, fmt.Errorf("invalid file format. Allowed: jpg, jpeg, png, gif, webp")
	}

	// Prepare upload folder
	uploadFolder := u.folder
	if folder != "" {
		uploadFolder = filepath.Join(u.folder, folder)
	}

	// Generate unique public ID
	publicID := generatePublicID(filename)

	// Helper for bool pointer
	overwriteFalse := false

	// Upload to Cloudinary
	uploadParams := uploader.UploadParams{
		Folder:         uploadFolder,
		PublicID:       publicID,
		Overwrite:      &overwriteFalse,
		ResourceType:   "image",
		Transformation: "q_auto,f_auto", // Auto quality and format
	}

	result, err := u.cld.Upload.Upload(ctx, file, uploadParams)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	return &UploadResult{
		URL:      result.SecureURL,
		PublicID: result.PublicID,
		Format:   result.Format,
		Width:    result.Width,
		Height:   result.Height,
		Bytes:    result.Bytes,
	}, nil
}

// DeleteImage deletes an image from Cloudinary
func (u *CloudinaryUploader) DeleteImage(ctx context.Context, publicID string) error {
	_, err := u.cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID:     publicID,
		ResourceType: "image",
	})
	if err != nil {
		return fmt.Errorf("failed to delete image: %w", err)
	}
	return nil
}

// UploadMultipleImages uploads multiple images
func (u *CloudinaryUploader) UploadMultipleImages(ctx context.Context, files []multipart.File, filenames []string, folder string) ([]*UploadResult, error) {
	if len(files) != len(filenames) {
		return nil, fmt.Errorf("files and filenames length mismatch")
	}

	results := make([]*UploadResult, 0, len(files))
	uploadedPublicIDs := make([]string, 0, len(files))

	for i, file := range files {
		result, err := u.UploadImage(ctx, file, filenames[i], folder)
		if err != nil {
			// Rollback: delete previously uploaded images
			for _, publicID := range uploadedPublicIDs {
				_ = u.DeleteImage(ctx, publicID)
			}
			return nil, fmt.Errorf("failed to upload file %s: %w", filenames[i], err)
		}
		results = append(results, result)
		uploadedPublicIDs = append(uploadedPublicIDs, result.PublicID)
	}

	return results, nil
}

// GetImageURL generates a URL for an image with transformations
func (u *CloudinaryUploader) GetImageURL(publicID string, width, height int, quality string) (string, error) {
	if quality == "" {
		quality = "auto"
	}

	// Build transformation string
	transformation := fmt.Sprintf("w_%d,h_%d,c_fill,q_%s,f_auto", width, height, quality)

	// Get asset
	asset, err := u.cld.Image(publicID)
	if err != nil {
		return "", fmt.Errorf("failed to generate URL: %w", err)
	}

	// Get string representation
	baseURL, err := asset.String()
	if err != nil {
		return "", fmt.Errorf("failed to get asset URL: %w", err)
	}

	return baseURL + "?" + transformation, nil
}

// GenerateThumbnail creates a thumbnail URL
func (u *CloudinaryUploader) GenerateThumbnail(publicID string) (string, error) {
	return u.GetImageURL(publicID, 300, 300, "auto")
}

// isValidImageFormat checks if the file has a valid image extension
func isValidImageFormat(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	validFormats := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}

	for _, format := range validFormats {
		if ext == format {
			return true
		}
	}
	return false
}

// generatePublicID generates a unique public ID for the file
func generatePublicID(filename string) string {
	timestamp := time.Now().UnixNano()
	ext := filepath.Ext(filename)
	name := strings.TrimSuffix(filename, ext)

	// Sanitize filename
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ToLower(name)

	return fmt.Sprintf("%s_%d", name, timestamp)
}

// ValidateImageFile validates file size and format
func ValidateImageFile(file multipart.File, header *multipart.FileHeader) error {
	// Check file size (max 10MB)
	maxSize := int64(10 * 1024 * 1024) // 10MB
	if header.Size > maxSize {
		return fmt.Errorf("file size exceeds maximum limit of 10MB")
	}

	// Check file format
	if !isValidImageFormat(header.Filename) {
		return fmt.Errorf("invalid file format. Allowed: jpg, jpeg, png, gif, webp")
	}

	// Read first 512 bytes to detect content type
	buffer := make([]byte, 512)
	_, err := file.Read(buffer)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Reset file pointer
	_, err = file.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("failed to reset file pointer: %w", err)
	}

	// Validate content type
	contentType := header.Header.Get("Content-Type")
	validTypes := []string{"image/jpeg", "image/jpg", "image/png", "image/gif", "image/webp"}

	isValid := false
	for _, validType := range validTypes {
		if contentType == validType {
			isValid = true
			break
		}
	}

	if !isValid {
		return fmt.Errorf("invalid content type: %s", contentType)
	}

	return nil
}
