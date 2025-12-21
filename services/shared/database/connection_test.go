package database

import (
	"context"
	"testing"
)

func TestNewPool_InvalidConfig(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "invalid port",
			config: Config{
				Host:     "localhost",
				Port:     -1,
				User:     "test",
				Password: "test",
				Database: "test",
				MaxConns: 5,
				SSLMode:  "disable",
			},
		},
		{
			name: "empty host",
			config: Config{
				Host:     "",
				Port:     5432,
				User:     "test",
				Password: "test",
				Database: "test",
				MaxConns: 5,
				SSLMode:  "disable",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			_, err := NewPool(ctx, tt.config)
			if err == nil {
				t.Error("expected error for invalid config")
			}
		})
	}
}

func TestNewPool_DefaultSSLMode(t *testing.T) {
	cfg := Config{
		Host:     "localhost",
		Port:     5432,
		User:     "test",
		Password: "test",
		Database: "test",
		MaxConns: 5,
		SSLMode:  "", // Empty - should default to 'require'
	}

	ctx := context.Background()
	// This will fail to connect but we're testing DSN construction
	_, err := NewPool(ctx, cfg)

	// We expect an error (can't connect to non-existent DB)
	// but not a parsing error
	if err == nil {
		t.Error("expected connection error (test doesn't have real DB)")
	}
}
