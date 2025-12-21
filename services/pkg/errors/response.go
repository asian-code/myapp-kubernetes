package errors

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// ErrorResponse represents the JSON structure for error responses
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error information
type ErrorDetail struct {
	Code    ErrorCode              `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// WriteError writes an error response to the HTTP response writer
func WriteError(w http.ResponseWriter, logger *log.Entry, err error) {
	var appErr *AppError

	// Check if it's an AppError, otherwise create a generic internal error
	if !IsAppError(err) {
		appErr = Internal("An unexpected error occurred")
		appErr.Err = err
	} else {
		appErr = GetAppError(err)
	}

	// Log error with appropriate level
	logEntry := logger.WithFields(log.Fields{
		"error_code": appErr.Code,
		"status":     appErr.StatusCode,
	})

	if appErr.Err != nil {
		logEntry = logEntry.WithError(appErr.Err)
	}

	if appErr.Details != nil {
		logEntry = logEntry.WithField("details", appErr.Details)
	}

	// Log 5xx errors as errors, 4xx as warnings
	if appErr.StatusCode >= 500 {
		logEntry.Error(appErr.Message)
	} else {
		logEntry.Warn(appErr.Message)
	}

	// Write response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.StatusCode)

	response := ErrorResponse{
		Error: ErrorDetail{
			Code:    appErr.Code,
			Message: appErr.Message,
			Details: appErr.Details,
		},
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.WithError(err).Error("Failed to encode error response")
	}
}

// WriteSuccess writes a successful JSON response
func WriteSuccess(w http.ResponseWriter, data interface{}, statusCode int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(data)
}
