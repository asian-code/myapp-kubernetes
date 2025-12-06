package auth

import (
	"testing"
)

func TestGenerateToken(t *testing.T) {
	token, err := GenerateToken("test-user", "secret-key")
	if err != nil {
		t.Errorf("Failed to generate token: %v", err)
	}

	if token == "" {
		t.Error("Expected token to be generated")
	}
}

func TestValidateToken(t *testing.T) {
	secret := "test-secret"
	userID := "test-user"

	// Generate a token
	token, err := GenerateToken(userID, secret)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Validate the token
	claims, err := ValidateToken(token, secret)
	if err != nil {
		t.Errorf("Failed to validate token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected userID %s, got %s", userID, claims.UserID)
	}
}

func TestValidateTokenInvalid(t *testing.T) {
	_, err := ValidateToken("invalid-token", "secret")
	if err == nil {
		t.Error("Expected error for invalid token")
	}
}
