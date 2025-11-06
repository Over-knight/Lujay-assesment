package middleware

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	apperrors "github.com/Over-knight/Lujay-assesment/internal/errors"
)

// SanitizationMiddleware provides protection against injection attacks
func SanitizationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Sanitize query parameters
		for key, values := range c.Request.URL.Query() {
			for i, value := range values {
				values[i] = sanitizeInput(value)
			}
			c.Request.URL.Query()[key] = values
		}

		c.Next()
	}
}

// sanitizeInput removes potentially dangerous characters
func sanitizeInput(input string) string {
	// Remove script tags and common injection patterns
	input = removeScriptTags(input)
	input = removeNoSQLInjection(input)
	return strings.TrimSpace(input)
}

// removeScriptTags removes HTML/JavaScript tags
func removeScriptTags(input string) string {
	scriptRegex := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	input = scriptRegex.ReplaceAllString(input, "")
	
	tagRegex := regexp.MustCompile(`<[^>]*>`)
	input = tagRegex.ReplaceAllString(input, "")
	
	return input
}

// removeNoSQLInjection removes common NoSQL injection patterns
func removeNoSQLInjection(input string) string {
	// Remove $-prefixed operators
	operatorRegex := regexp.MustCompile(`\$[a-zA-Z]+`)
	input = operatorRegex.ReplaceAllString(input, "")
	
	// Remove potential MongoDB operators in object notation
	input = strings.ReplaceAll(input, "{", "")
	input = strings.ReplaceAll(input, "}", "")
	input = strings.ReplaceAll(input, "[", "")
	input = strings.ReplaceAll(input, "]", "")
	
	return input
}

// ValidateQueryParams validates and limits allowed query parameters
func ValidateQueryParams(allowedParams []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		queryParams := c.Request.URL.Query()
		
		for param := range queryParams {
			if !contains(allowedParams, param) {
				apperrors.RespondWithError(c, apperrors.NewBadRequestError("Invalid query parameter: "+param))
				c.Abort()
				return
			}
		}
		
		c.Next()
	}
}

// ValidatePaginationParams validates pagination parameters
func ValidatePaginationParams() gin.HandlerFunc {
	return func(c *gin.Context) {
		if pageStr := c.Query("page"); pageStr != "" {
			page, err := strconv.Atoi(pageStr)
			if err != nil || page < 1 {
				apperrors.RespondWithError(c, apperrors.NewBadRequestError("page must be a positive integer"))
				c.Abort()
				return
			}
		}

		if limitStr := c.Query("limit"); limitStr != "" {
			limit, err := strconv.Atoi(limitStr)
			if err != nil || limit < 1 {
				apperrors.RespondWithError(c, apperrors.NewBadRequestError("limit must be a positive integer"))
				c.Abort()
				return
			}
			if limit > 100 {
				apperrors.RespondWithError(c, apperrors.NewBadRequestError("limit cannot exceed 100"))
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// ValidateVehicleQueryParams validates vehicle-specific query parameters
func ValidateVehicleQueryParams() gin.HandlerFunc {
	allowedParams := []string{
		"page", "limit", "make", "model", "status",
		"minPrice", "maxPrice", "minYear", "maxYear",
		"sortBy", "sortOrder",
	}
	
	return func(c *gin.Context) {
		ValidateQueryParams(allowedParams)(c)
		if c.IsAborted() {
			return
		}

		// Validate status if provided
		if status := c.Query("status"); status != "" {
			validStatuses := []string{"active", "sold", "archived"}
			if !contains(validStatuses, status) {
				apperrors.RespondWithError(c, apperrors.NewBadRequestError("invalid status value"))
				c.Abort()
				return
			}
		}

		// Validate sortBy if provided
		if sortBy := c.Query("sortBy"); sortBy != "" {
			validSortFields := []string{"price", "year", "mileage", "createdAt"}
			if !contains(validSortFields, sortBy) {
				apperrors.RespondWithError(c, apperrors.NewBadRequestError("invalid sortBy field"))
				c.Abort()
				return
			}
		}

		// Validate sortOrder if provided
		if sortOrder := c.Query("sortOrder"); sortOrder != "" {
			validSortOrders := []string{"asc", "desc"}
			if !contains(validSortOrders, sortOrder) {
				apperrors.RespondWithError(c, apperrors.NewBadRequestError("sortOrder must be asc or desc"))
				c.Abort()
				return
			}
		}

		// Validate price range
		if minPriceStr := c.Query("minPrice"); minPriceStr != "" {
			minPrice, err := strconv.ParseFloat(minPriceStr, 64)
			if err != nil || minPrice < 0 {
				apperrors.RespondWithError(c, apperrors.NewBadRequestError("minPrice must be a non-negative number"))
				c.Abort()
				return
			}
		}

		if maxPriceStr := c.Query("maxPrice"); maxPriceStr != "" {
			maxPrice, err := strconv.ParseFloat(maxPriceStr, 64)
			if err != nil || maxPrice < 0 {
				apperrors.RespondWithError(c, apperrors.NewBadRequestError("maxPrice must be a non-negative number"))
				c.Abort()
				return
			}
		}

		// Validate year range
		if minYearStr := c.Query("minYear"); minYearStr != "" {
			minYear, err := strconv.Atoi(minYearStr)
			if err != nil || minYear < 1900 {
				apperrors.RespondWithError(c, apperrors.NewBadRequestError("minYear must be a valid year"))
				c.Abort()
				return
			}
		}

		if maxYearStr := c.Query("maxYear"); maxYearStr != "" {
			maxYear, err := strconv.Atoi(maxYearStr)
			if err != nil || maxYear < 1900 {
				apperrors.RespondWithError(c, apperrors.NewBadRequestError("maxYear must be a valid year"))
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// ValidateInspectionQueryParams validates inspection-specific query parameters
func ValidateInspectionQueryParams() gin.HandlerFunc {
	allowedParams := []string{"page", "limit", "status"}
	
	return func(c *gin.Context) {
		ValidateQueryParams(allowedParams)(c)
		if c.IsAborted() {
			return
		}

		// Validate status if provided
		if status := c.Query("status"); status != "" {
			validStatuses := []string{"pending", "scheduled", "completed", "cancelled"}
			if !contains(validStatuses, status) {
				apperrors.RespondWithError(c, apperrors.NewBadRequestError("invalid status value"))
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// ValidateTransactionQueryParams validates transaction-specific query parameters
func ValidateTransactionQueryParams() gin.HandlerFunc {
	allowedParams := []string{"page", "limit", "status"}
	
	return func(c *gin.Context) {
		ValidateQueryParams(allowedParams)(c)
		if c.IsAborted() {
			return
		}

		// Validate status if provided
		if status := c.Query("status"); status != "" {
			validStatuses := []string{"pending", "completed", "cancelled", "failed"}
			if !contains(validStatuses, status) {
				apperrors.RespondWithError(c, apperrors.NewBadRequestError("invalid status value"))
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// BasicRateLimitMiddleware provides basic rate limiting (placeholder - use Redis-based one)
// Deprecated: Use RateLimitMiddleware from cache.go with Redis for production
func BasicRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add rate limiting headers
		c.Header("X-RateLimit-Limit", "1000")
		c.Header("X-RateLimit-Remaining", "999")
		
		// In production, implement proper rate limiting with Redis or similar
		c.Next()
	}
}

// SecurityHeadersMiddleware adds security headers to responses
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		
		c.Next()
	}
}

// RequestSizeLimitMiddleware limits the size of request bodies
func RequestSizeLimitMiddleware(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
		c.Next()
	}
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
