package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/asian-code/myapp-kubernetes/services/api-service/internal/repository"
	integration "github.com/asian-code/myapp-kubernetes/services/pkg/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_CreateAndGet_Integration(t *testing.T) {
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

	t.Run("CreateUser and GetUserByID", func(t *testing.T) {
		username := "testuser"
		email := "test@example.com"
		passwordHash := "hashed_password"

		// Create user
		userID, err := repo.CreateUser(ctx, username, email, passwordHash)
		assert.NoError(t, err)
		assert.NotEmpty(t, userID)

		// Get user by ID
		user, err := repo.GetUserByID(ctx, userID)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, userID, user.ID)
		assert.Equal(t, username, user.Username)
		assert.Equal(t, email, user.Email)
		assert.Equal(t, passwordHash, user.PasswordHash)
		assert.True(t, user.IsActive)
	})

	t.Run("GetUserByUsername", func(t *testing.T) {
		username := "testuser2"
		email := "test2@example.com"
		passwordHash := "hashed_password2"

		// Create user
		userID, err := repo.CreateUser(ctx, username, email, passwordHash)
		assert.NoError(t, err)

		// Get user by username
		user, err := repo.GetUserByUsername(ctx, username)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, userID, user.ID)
		assert.Equal(t, username, user.Username)
	})

	t.Run("GetUserByEmail", func(t *testing.T) {
		username := "testuser3"
		email := "test3@example.com"
		passwordHash := "hashed_password3"

		// Create user
		userID, err := repo.CreateUser(ctx, username, email, passwordHash)
		assert.NoError(t, err)

		// Get user by email
		user, err := repo.GetUserByEmail(ctx, email)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, userID, user.ID)
		assert.Equal(t, email, user.Email)
	})

	t.Run("UpdateLastLogin", func(t *testing.T) {
		username := "testuser4"
		email := "test4@example.com"
		passwordHash := "hashed_password4"

		// Create user
		userID, err := repo.CreateUser(ctx, username, email, passwordHash)
		assert.NoError(t, err)

		// Get initial user
		user, err := repo.GetUserByID(ctx, userID)
		assert.NoError(t, err)
		assert.Nil(t, user.LastLogin)

		// Update last login
		time.Sleep(100 * time.Millisecond) // Ensure time difference
		loginTime := time.Now()
		err = repo.UpdateLastLogin(ctx, userID, loginTime)
		assert.NoError(t, err)

		// Verify update
		updatedUser, err := repo.GetUserByID(ctx, userID)
		assert.NoError(t, err)
		assert.NotNil(t, updatedUser.LastLogin)
		assert.True(t, updatedUser.LastLogin.After(user.CreatedAt))
	})

	t.Run("UpdateUser", func(t *testing.T) {
		username := "testuser5"
		email := "test5@example.com"
		passwordHash := "hashed_password5"

		// Create user
		userID, err := repo.CreateUser(ctx, username, email, passwordHash)
		assert.NoError(t, err)

		// Get user
		user, err := repo.GetUserByID(ctx, userID)
		assert.NoError(t, err)

		// Update user
		updates := map[string]interface{}{
			"email":     "updated@example.com",
			"is_active": false,
		}
		err = repo.UpdateUser(ctx, userID, updates)
		assert.NoError(t, err)

		// Verify update
		updatedUser, err := repo.GetUserByID(ctx, userID)
		assert.NoError(t, err)
		assert.Equal(t, "updated@example.com", updatedUser.Email)
		assert.False(t, updatedUser.IsActive)
	})

	t.Run("DeleteUser", func(t *testing.T) {
		username := "testuser6"
		email := "test6@example.com"
		passwordHash := "hashed_password6"

		// Create user
		userID, err := repo.CreateUser(ctx, username, email, passwordHash)
		assert.NoError(t, err)

		// Delete user
		err = repo.DeleteUser(ctx, userID)
		assert.NoError(t, err)

		// Verify deletion
		user, err := repo.GetUserByID(ctx, userID)
		assert.NoError(t, err)
		assert.Nil(t, user)
	})

	t.Run("DuplicateUsername", func(t *testing.T) {
		username := "duplicate"
		email := "test7@example.com"
		passwordHash := "hashed_password7"

		// Create first user
		_, err := repo.CreateUser(ctx, username, email, passwordHash)
		assert.NoError(t, err)

		// Try to create with same username
		_, err = repo.CreateUser(ctx, username, "different@example.com", passwordHash)
		assert.Error(t, err, "Should fail with duplicate username")
	})
}
