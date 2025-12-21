package errors

import (
	"errors"
	"net/http"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name           string
		code           ErrorCode
		message        string
		expectedStatus int
	}{
		{
			name:           "bad request error",
			code:           ErrCodeBadRequest,
			message:        "Invalid input",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthorized error",
			code:           ErrCodeUnauthorized,
			message:        "Authentication required",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "not found error",
			code:           ErrCodeNotFound,
			message:        "Resource not found",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "internal error",
			code:           ErrCodeInternal,
			message:        "Something went wrong",
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := New(tt.code, tt.message)

			if err.Code != tt.code {
				t.Errorf("expected code %s, got %s", tt.code, err.Code)
			}

			if err.Message != tt.message {
				t.Errorf("expected message %s, got %s", tt.message, err.Message)
			}

			if err.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, err.StatusCode)
			}
		})
	}
}

func TestWrap(t *testing.T) {
	originalErr := errors.New("database connection failed")
	appErr := Wrap(originalErr, ErrCodeDatabaseError, "Failed to query database")

	if appErr.Code != ErrCodeDatabaseError {
		t.Errorf("expected code %s, got %s", ErrCodeDatabaseError, appErr.Code)
	}

	if appErr.Err != originalErr {
		t.Errorf("expected wrapped error to be %v, got %v", originalErr, appErr.Err)
	}

	if !errors.Is(appErr, originalErr) {
		t.Error("wrapped error should be unwrappable")
	}
}

func TestWithDetails(t *testing.T) {
	err := BadRequest("Validation failed").
		WithDetails("field", "email").
		WithDetails("reason", "invalid format")

	if len(err.Details) != 2 {
		t.Errorf("expected 2 details, got %d", len(err.Details))
	}

	if err.Details["field"] != "email" {
		t.Errorf("expected field detail to be 'email', got %v", err.Details["field"])
	}

	if err.Details["reason"] != "invalid format" {
		t.Errorf("expected reason detail to be 'invalid format', got %v", err.Details["reason"])
	}
}

func TestIsAppError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "AppError",
			err:      BadRequest("test"),
			expected: true,
		},
		{
			name:     "standard error",
			err:      errors.New("standard error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsAppError(tt.err)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGetAppError(t *testing.T) {
	appErr := NotFound("User not found")

	retrieved := GetAppError(appErr)
	if retrieved == nil {
		t.Fatal("expected to retrieve AppError, got nil")
	}

	if retrieved.Code != ErrCodeNotFound {
		t.Errorf("expected code %s, got %s", ErrCodeNotFound, retrieved.Code)
	}

	standardErr := errors.New("standard error")
	retrieved = GetAppError(standardErr)
	if retrieved != nil {
		t.Errorf("expected nil for standard error, got %v", retrieved)
	}
}

func TestErrorHelpers(t *testing.T) {
	tests := []struct {
		name           string
		createFunc     func() *AppError
		expectedCode   ErrorCode
		expectedStatus int
	}{
		{
			name:           "BadRequest",
			createFunc:     func() *AppError { return BadRequest("test") },
			expectedCode:   ErrCodeBadRequest,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Unauthorized",
			createFunc:     func() *AppError { return Unauthorized("test") },
			expectedCode:   ErrCodeUnauthorized,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Forbidden",
			createFunc:     func() *AppError { return Forbidden("test") },
			expectedCode:   ErrCodeForbidden,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "NotFound",
			createFunc:     func() *AppError { return NotFound("test") },
			expectedCode:   ErrCodeNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Conflict",
			createFunc:     func() *AppError { return Conflict("test") },
			expectedCode:   ErrCodeConflict,
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "ValidationFailed",
			createFunc:     func() *AppError { return ValidationFailed("test") },
			expectedCode:   ErrCodeValidationFailed,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Internal",
			createFunc:     func() *AppError { return Internal("test") },
			expectedCode:   ErrCodeInternal,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.createFunc()

			if err.Code != tt.expectedCode {
				t.Errorf("expected code %s, got %s", tt.expectedCode, err.Code)
			}

			if err.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, err.StatusCode)
			}
		})
	}
}
