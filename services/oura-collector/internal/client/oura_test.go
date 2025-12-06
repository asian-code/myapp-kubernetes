package client

import (
	"context"
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestOuraClientNew(t *testing.T) {
	logger := log.NewEntry(log.New())
	client := New("test-api-key", logger)

	if client == nil {
		t.Error("Expected client to be created")
	}

	if client.apiKey != "test-api-key" {
		t.Errorf("Expected apiKey to be 'test-api-key', got %s", client.apiKey)
	}
}

func TestOuraClientGetSleepData(t *testing.T) {
	logger := log.NewEntry(log.New())
	client := New("invalid-key", logger)

	ctx := context.Background()

	// This will fail because we don't have a valid API key
	// Just testing the error handling
	_, err := client.GetSleepData(ctx, "2024-01-01")
	if err == nil {
		t.Error("Expected error with invalid API key")
	}
}
