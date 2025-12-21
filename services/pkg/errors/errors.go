package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorCode represents application-specific error codes
type ErrorCode string

const (
	// Client errors (4xx)
	ErrCodeBadRequest         ErrorCode = "BAD_REQUEST"
	ErrCodeUnauthorized       ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden          ErrorCode = "FORBIDDEN"
	ErrCodeNotFound           ErrorCode = "NOT_FOUND"
	ErrCodeConflict           ErrorCode = "CONFLICT"
	ErrCodeValidationFailed   ErrorCode = "VALIDATION_FAILED"
	ErrCodeInvalidCredentials ErrorCode = "INVALID_CREDENTIALS"
	ErrCodeTokenExpired       ErrorCode = "TOKEN_EXPIRED"
	ErrCodeInvalidToken       ErrorCode = "INVALID_TOKEN"

	// Server errors (5xx)
	ErrCodeInternal           ErrorCode = "INTERNAL_ERROR"
	ErrCodeDatabaseError      ErrorCode = "DATABASE_ERROR"
	ErrCodeExternalServiceErr ErrorCode = "EXTERNAL_SERVICE_ERROR"
	ErrCodeConfigError        ErrorCode = "CONFIG_ERROR"
)

// AppError represents a structured application error
type AppError struct {
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	StatusCode int                    `json:"-"`
	Err        error                  `json:"-"`
	Details    map[string]interface{} `json:"details,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a new AppError
func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: getStatusCode(code),
	}
}

// Wrap wraps an existing error with application context
func Wrap(err error, code ErrorCode, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: getStatusCode(code),
		Err:        err,
	}
}

// WithDetails adds additional context to an error
func (e *AppError) WithDetails(key string, value interface{}) *AppError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// getStatusCode maps error codes to HTTP status codes
func getStatusCode(code ErrorCode) int {
	switch code {
	case ErrCodeBadRequest, ErrCodeValidationFailed:
		return http.StatusBadRequest
	case ErrCodeUnauthorized, ErrCodeInvalidCredentials, ErrCodeTokenExpired, ErrCodeInvalidToken:
		return http.StatusUnauthorized
	case ErrCodeForbidden:
		return http.StatusForbidden
	case ErrCodeNotFound:
		return http.StatusNotFound
	case ErrCodeConflict:
		return http.StatusConflict
	case ErrCodeDatabaseError, ErrCodeExternalServiceErr, ErrCodeInternal, ErrCodeConfigError:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// GetAppError extracts AppError from error chain
func GetAppError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return nil
}

// Common error constructors
func BadRequest(message string) *AppError {
	return New(ErrCodeBadRequest, message)
}

func Unauthorized(message string) *AppError {
	return New(ErrCodeUnauthorized, message)
}

func Forbidden(message string) *AppError {
	return New(ErrCodeForbidden, message)
}

func NotFound(message string) *AppError {
	return New(ErrCodeNotFound, message)
}

func Conflict(message string) *AppError {
	return New(ErrCodeConflict, message)
}

func ValidationFailed(message string) *AppError {
	return New(ErrCodeValidationFailed, message)
}

func InvalidCredentials(message string) *AppError {
	return New(ErrCodeInvalidCredentials, message)
}

func Internal(message string) *AppError {
	return New(ErrCodeInternal, message)
}

func Database(err error, message string) *AppError {
	return Wrap(err, ErrCodeDatabaseError, message)
}

func ExternalService(err error, message string) *AppError {
	return Wrap(err, ErrCodeExternalServiceErr, message)
}
