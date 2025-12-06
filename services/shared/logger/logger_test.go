package logger

import (
	"testing"
)

func TestInit(t *testing.T) {
	logger := Init("test-service")

	if logger == nil {
		t.Error("Expected logger to be initialized")
	}

	// Check if the service field is set
	if logger.Data["service"] != "test-service" {
		t.Errorf("Expected service field to be 'test-service', got %v", logger.Data["service"])
	}
}
