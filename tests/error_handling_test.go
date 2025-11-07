package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Over-knight/Lujay-assesment/internal/errors"
)

func TestAppError(t *testing.T) {
	err := errors.NewBadRequestError("Invalid input")
	assert.Equal(t, "Invalid input", err.Error())
	assert.Equal(t, errors.ErrCodeBadRequest, err.Code)
	assert.Equal(t, 400, err.StatusCode)
}

func TestErrorResponses(t *testing.T) {
	tests := []struct {
		name       string
		errorFunc  func() *errors.AppError
		wantCode   string
		wantStatus int
	}{
		{
			name:       "Bad Request",
			errorFunc:  func() *errors.AppError { return errors.NewBadRequestError("bad request") },
			wantCode:   errors.ErrCodeBadRequest,
			wantStatus: 400,
		},
		{
			name:       "Not Found",
			errorFunc:  func() *errors.AppError { return errors.NewNotFoundError("not found") },
			wantCode:   errors.ErrCodeNotFound,
			wantStatus: 404,
		},
		{
			name:       "Validation Error",
			errorFunc:  func() *errors.AppError { return errors.NewValidationError("validation failed") },
			wantCode:   errors.ErrCodeValidation,
			wantStatus: 400,
		},
		{
			name:       "Conflict",
			errorFunc:  func() *errors.AppError { return errors.NewConflictError("conflict") },
			wantCode:   errors.ErrCodeConflict,
			wantStatus: 409,
		},
		{
			name:       "Database Error",
			errorFunc:  func() *errors.AppError { return errors.NewDatabaseError("db error") },
			wantCode:   errors.ErrCodeDatabaseError,
			wantStatus: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.errorFunc()
			assert.Equal(t, tt.wantCode, err.Code)
			assert.Equal(t, tt.wantStatus, err.StatusCode)
		})
	}
}

func TestPredefinedErrors(t *testing.T) {
	assert.Equal(t, errors.ErrCodeUnauthorized, errors.ErrUnauthorized.Code)
	assert.Equal(t, 401, errors.ErrUnauthorized.StatusCode)

	assert.Equal(t, errors.ErrCodeForbidden, errors.ErrForbidden.Code)
	assert.Equal(t, 403, errors.ErrForbidden.StatusCode)

	assert.Equal(t, errors.ErrCodeInternalServer, errors.ErrInternalServer.Code)
	assert.Equal(t, 500, errors.ErrInternalServer.StatusCode)
}
