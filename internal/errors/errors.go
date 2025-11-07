package errors

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AppError represents a standardized application error
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	return e.Message
}

// ErrorResponse represents the JSON error response structure
type ErrorResponse struct {
	Error AppError `json:"error"`
}

// ValidationError represents a validation error with field details
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrorResponse represents the JSON validation error response
type ValidationErrorResponse struct {
	Error struct {
		Code    string            `json:"code"`
		Message string            `json:"message"`
		Details []ValidationError `json:"details,omitempty"`
	} `json:"error"`
}

// Error codes
const (
	ErrCodeBadRequest              = "BAD_REQUEST"
	ErrCodeUnauthorized            = "UNAUTHORIZED"
	ErrCodeForbidden               = "FORBIDDEN"
	ErrCodeNotFound                = "NOT_FOUND"
	ErrCodeConflict                = "CONFLICT"
	ErrCodeValidation              = "VALIDATION_ERROR"
	ErrCodeInternalServer          = "INTERNAL_SERVER_ERROR"
	ErrCodeDatabaseError           = "DATABASE_ERROR"
	ErrCodeInvalidToken            = "INVALID_TOKEN"
	ErrCodeExpiredToken            = "EXPIRED_TOKEN"
	ErrCodeInvalidCredentials      = "INVALID_CREDENTIALS"
	ErrCodeResourceExists          = "RESOURCE_EXISTS"
	ErrCodeInsufficientPermissions = "INSUFFICIENT_PERMISSIONS"
)

// Predefined errors
var (
	ErrBadRequest = &AppError{
		Code:       ErrCodeBadRequest,
		Message:    "Invalid request",
		StatusCode: http.StatusBadRequest,
	}

	ErrUnauthorized = &AppError{
		Code:       ErrCodeUnauthorized,
		Message:    "Unauthorized",
		StatusCode: http.StatusUnauthorized,
	}

	ErrForbidden = &AppError{
		Code:       ErrCodeForbidden,
		Message:    "Access forbidden",
		StatusCode: http.StatusForbidden,
	}

	ErrNotFound = &AppError{
		Code:       ErrCodeNotFound,
		Message:    "Resource not found",
		StatusCode: http.StatusNotFound,
	}

	ErrInternalServer = &AppError{
		Code:       ErrCodeInternalServer,
		Message:    "Internal server error",
		StatusCode: http.StatusInternalServerError,
	}

	ErrInvalidToken = &AppError{
		Code:       ErrCodeInvalidToken,
		Message:    "Invalid or expired token",
		StatusCode: http.StatusUnauthorized,
	}

	ErrInvalidCredentials = &AppError{
		Code:       ErrCodeInvalidCredentials,
		Message:    "Invalid email or password",
		StatusCode: http.StatusUnauthorized,
	}

	ErrInsufficientPermissions = &AppError{
		Code:       ErrCodeInsufficientPermissions,
		Message:    "Insufficient permissions to perform this action",
		StatusCode: http.StatusForbidden,
	}
)

// NewAppError creates a new application error
func NewAppError(code string, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// NewBadRequestError creates a bad request error with custom message
func NewBadRequestError(message string) *AppError {
	return &AppError{
		Code:       ErrCodeBadRequest,
		Message:    message,
		StatusCode: http.StatusBadRequest,
	}
}

// NewNotFoundError creates a not found error with custom message
func NewNotFoundError(message string) *AppError {
	return &AppError{
		Code:       ErrCodeNotFound,
		Message:    message,
		StatusCode: http.StatusNotFound,
	}
}

// NewValidationError creates a validation error with custom message
func NewValidationError(message string) *AppError {
	return &AppError{
		Code:       ErrCodeValidation,
		Message:    message,
		StatusCode: http.StatusBadRequest,
	}
}

// NewConflictError creates a conflict error with custom message
func NewConflictError(message string) *AppError {
	return &AppError{
		Code:       ErrCodeConflict,
		Message:    message,
		StatusCode: http.StatusConflict,
	}
}

// NewDatabaseError creates a database error
func NewDatabaseError(message string) *AppError {
	return &AppError{
		Code:       ErrCodeDatabaseError,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
	}
}

// RespondWithError sends a standardized error response
func RespondWithError(c *gin.Context, err *AppError) {
	c.JSON(err.StatusCode, ErrorResponse{
		Error: *err,
	})
}

// RespondWithValidationError sends a validation error response with field details
func RespondWithValidationError(c *gin.Context, message string, validationErrors []ValidationError) {
	response := ValidationErrorResponse{}
	response.Error.Code = ErrCodeValidation
	response.Error.Message = message
	response.Error.Details = validationErrors

	c.JSON(http.StatusBadRequest, response)
}

// HandleError handles errors and sends appropriate response
// Converts common error strings to AppErrors
func HandleError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	// Check if it's already an AppError
	if appErr, ok := err.(*AppError); ok {
		RespondWithError(c, appErr)
		return
	}

	// Map common error messages to appropriate errors
	errMsg := err.Error()
	switch errMsg {
	case "mongo: no documents in result":
		RespondWithError(c, NewNotFoundError("Resource not found"))
	case "invalid vehicleId", "invalid buyerId", "invalid inspectionId":
		RespondWithError(c, NewBadRequestError(errMsg))
	case "vehicle not found", "inspection not found", "transaction not found", "user not found":
		RespondWithError(c, NewNotFoundError(errMsg))
	case "you are not the owner of this vehicle", "you are not authorized to update this transaction", "you are not authorized to cancel this transaction", "only the seller can complete this transaction":
		RespondWithError(c, ErrInsufficientPermissions)
	case "vehicle is not available for sale", "cannot create transaction with yourself", "transaction is not pending", "only pending transactions can be cancelled":
		RespondWithError(c, NewBadRequestError(errMsg))
	case "email already exists":
		RespondWithError(c, NewConflictError(errMsg))
	default:
		// For unknown errors, return internal server error
		RespondWithError(c, ErrInternalServer)
	}
}

// Success response helpers

// SuccessResponse represents a standardized success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// RespondWithSuccess sends a standardized success response
func RespondWithSuccess(c *gin.Context, statusCode int, message string, data interface{}) {
	response := SuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
	c.JSON(statusCode, response)
}

// RespondWithData sends a success response with just data
func RespondWithData(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, data)
}

// RespondCreated sends a 201 Created response
func RespondCreated(c *gin.Context, message string, data interface{}) {
	RespondWithSuccess(c, http.StatusCreated, message, data)
}

// RespondOK sends a 200 OK response
func RespondOK(c *gin.Context, message string, data interface{}) {
	RespondWithSuccess(c, http.StatusOK, message, data)
}

// RespondNoContent sends a 204 No Content response
func RespondNoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}
