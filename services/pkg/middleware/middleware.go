package middleware

import (
	"context"
	"net/http"
	"time"

	apperrors "github.com/asian-code/myapp-kubernetes/services/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// ErrorHandler is middleware that handles panics and provides consistent error responses
func ErrorHandler(logger *log.Entry) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.WithField("panic", err).Error("Request panic recovered")
					apperrors.WriteError(w, logger, apperrors.Internal("An unexpected error occurred"))
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// RequestLogger logs all incoming requests with timing
func RequestLogger(logger *log.Entry) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create response writer wrapper to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Add request ID to context
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = generateRequestID()
			}
			ctx := context.WithValue(r.Context(), "request-id", requestID)
			r = r.WithContext(ctx)

			// Log request
			logger.WithFields(log.Fields{
				"request_id": requestID,
				"method":     r.Method,
				"path":       r.URL.Path,
				"remote":     r.RemoteAddr,
			}).Info("Request started")

			next.ServeHTTP(wrapped, r)

			// Log response
			duration := time.Since(start)
			logger.WithFields(log.Fields{
				"request_id": requestID,
				"method":     r.Method,
				"path":       r.URL.Path,
				"status":     wrapped.statusCode,
				"duration":   duration.Milliseconds(),
			}).Info("Request completed")
		})
	}
}

// RequestID middleware adds a request ID to the response headers
func RequestID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = generateRequestID()
			}
			w.Header().Set("X-Request-ID", requestID)
			next.ServeHTTP(w, r)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// generateRequestID creates a simple request ID
// In production, consider using UUID or similar
func generateRequestID() string {
	return time.Now().Format("20060102150405.000000")
}
