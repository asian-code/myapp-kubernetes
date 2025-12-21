package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/asian-code/myapp-kubernetes/services/api-service/internal/repository"
	"github.com/asian-code/myapp-kubernetes/services/pkg/interfaces"
	integration "github.com/asian-code/myapp-kubernetes/services/pkg/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOAuthRepository_SaveAndGet_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()

	// Setup test container
	pgContainer, err := integration.SetupPostgresContainer(ctx)
	require.NoError(t, err, "Failed to start PostgreSQL container")
	defer pgContainer.Close(ctx)

	// Get database connection
	pool, err := pgContainer.GetPool(ctx)
	require.NoError(t, err, "Failed to connect to database")
	defer pool.Close()

	// Run migrations
	err = pgContainer.RunMigrations(ctx, pool)
	require.NoError(t, err, "Failed to run migrations")

	// Create repository
	repo := repository.New(pool, nil)

	// Create a test user first
	userID, err := repo.CreateUser(ctx, "testuser", "test@example.com", "hashedpass")
	require.NoError(t, err)

	t.Run("SaveToken and GetToken", func(t *testing.T) {
		token := &interfaces.OAuthToken{
			UserID:       userID,
			AccessToken:  "test-access-token",
			RefreshToken: "test-refresh-token",
			ExpiresAt:    time.Now().Add(1 * time.Hour),
			Scope:        "daily personal email",
			Provider:     "oura",
		}

		// Save token
		err := repo.SaveToken(ctx, token)
		assert.NoError(t, err)

		// Get token
		retrievedToken, err := repo.GetToken(ctx, userID, "oura")
		assert.NoError(t, err)
		assert.NotNil(t, retrievedToken)
		assert.Equal(t, token.UserID, retrievedToken.UserID)
		assert.Equal(t, token.AccessToken, retrievedToken.AccessToken)
		assert.Equal(t, token.RefreshToken, retrievedToken.RefreshToken)
		assert.Equal(t, token.Scope, retrievedToken.Scope)
		assert.Equal(t, token.Provider, retrievedToken.Provider)
		assert.WithinDuration(t, token.ExpiresAt, retrievedToken.ExpiresAt, 1*time.Second)
	})

	t.Run("SaveToken_UpsertOnConflict", func(t *testing.T) {
		token1 := &interfaces.OAuthToken{
			UserID:       userID,
			AccessToken:  "first-access-token",
			RefreshToken: "first-refresh-token",
			ExpiresAt:    time.Now().Add(1 * time.Hour),
			Scope:        "daily",
			Provider:     "oura",
		}

		// Save initial token
		err := repo.SaveToken(ctx, token1)
		assert.NoError(t, err)

		// Save updated token (should upsert, not fail)
		token2 := &interfaces.OAuthToken{
			UserID:       userID,
			AccessToken:  "second-access-token",
			RefreshToken: "second-refresh-token",
			ExpiresAt:    time.Now().Add(2 * time.Hour),
			Scope:        "daily personal",
			Provider:     "oura",
		}

		err = repo.SaveToken(ctx, token2)
		assert.NoError(t, err)

		// Verify the token was updated
		retrievedToken, err := repo.GetToken(ctx, userID, "oura")
		assert.NoError(t, err)
		assert.Equal(t, "second-access-token", retrievedToken.AccessToken)
		assert.Equal(t, "second-refresh-token", retrievedToken.RefreshToken)
	})

	t.Run("UpdateToken", func(t *testing.T) {
		// Save initial token
		token := &interfaces.OAuthToken{
			UserID:       userID,
			AccessToken:  "initial-token",
			RefreshToken: "initial-refresh",
			ExpiresAt:    time.Now().Add(1 * time.Hour),
			Scope:        "daily",
			Provider:     "oura",
		}
		err := repo.SaveToken(ctx, token)
		require.NoError(t, err)

		// Update token
		updatedToken := &interfaces.OAuthToken{
			UserID:       userID,
			AccessToken:  "updated-token",
			RefreshToken: "updated-refresh",
			ExpiresAt:    time.Now().Add(3 * time.Hour),
			Scope:        "daily personal email",
			Provider:     "oura",
		}
		err = repo.UpdateToken(ctx, updatedToken)
		assert.NoError(t, err)

		// Verify update
		retrievedToken, err := repo.GetToken(ctx, userID, "oura")
		assert.NoError(t, err)
		assert.Equal(t, "updated-token", retrievedToken.AccessToken)
		assert.Equal(t, "updated-refresh", retrievedToken.RefreshToken)
		assert.Equal(t, "daily personal email", retrievedToken.Scope)
	})

	t.Run("DeleteToken", func(t *testing.T) {
		// Save token
		token := &interfaces.OAuthToken{
			UserID:       userID,
			AccessToken:  "token-to-delete",
			RefreshToken: "refresh-to-delete",
			ExpiresAt:    time.Now().Add(1 * time.Hour),
			Scope:        "daily",
			Provider:     "oura",
		}
		err := repo.SaveToken(ctx, token)
		require.NoError(t, err)

		// Verify token exists
		retrievedToken, err := repo.GetToken(ctx, userID, "oura")
		assert.NoError(t, err)
		assert.NotNil(t, retrievedToken)

		// Delete token
		err = repo.DeleteToken(ctx, userID, "oura")
		assert.NoError(t, err)

		// Verify token was deleted
		retrievedToken, err = repo.GetToken(ctx, userID, "oura")
		assert.NoError(t, err)
		assert.Nil(t, retrievedToken)
	})

	t.Run("GetToken_NotFound", func(t *testing.T) {
		// Try to get token for non-existent user
		token, err := repo.GetToken(ctx, "non-existent-user-id", "oura")
		assert.NoError(t, err)
		assert.Nil(t, token)
	})

	t.Run("UpdateToken_NotFound", func(t *testing.T) {
		// Try to update token for non-existent user
		token := &interfaces.OAuthToken{
			UserID:       "non-existent-user",
			AccessToken:  "token",
			RefreshToken: "refresh",
			ExpiresAt:    time.Now().Add(1 * time.Hour),
			Scope:        "daily",
			Provider:     "oura",
		}
		err := repo.UpdateToken(ctx, token)
		assert.Error(t, err)
	})
}
