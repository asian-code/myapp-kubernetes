package user

import (
	"context"
	"time"

	"github.com/asian-code/myapp-kubernetes/services/pkg/errors"
	"github.com/asian-code/myapp-kubernetes/services/pkg/interfaces"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Service implements the UserService interface
type Service struct {
	repo      interfaces.UserRepository
	jwtSecret string
	logger    interfaces.Logger
}

// NewService creates a new user service
func NewService(repo interfaces.UserRepository, jwtSecret string, logger interfaces.Logger) *Service {
	return &Service{
		repo:      repo,
		jwtSecret: jwtSecret,
		logger:    logger,
	}
}

// Register creates a new user account
func (s *Service) Register(ctx context.Context, username, email, password string) (*interfaces.UserDTO, error) {
	// Validate input
	if username == "" || email == "" || password == "" {
		return nil, errors.ValidationFailed("Username, email, and password are required")
	}

	if len(password) < 8 {
		return nil, errors.ValidationFailed("Password must be at least 8 characters")
	}

	// Check if user already exists
	existingUser, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, errors.Database(err, "Failed to check existing user")
	}
	if existingUser != nil {
		return nil, errors.Conflict("Username already taken")
	}

	existingEmail, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, errors.Database(err, "Failed to check existing email")
	}
	if existingEmail != nil {
		return nil, errors.Conflict("Email already registered")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.Internal("Failed to hash password")
	}

	// Create user
	userID, err := s.repo.CreateUser(ctx, username, email, string(hashedPassword))
	if err != nil {
		return nil, errors.Database(err, "Failed to create user")
	}

	// Fetch created user
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errors.Database(err, "Failed to fetch created user")
	}

	s.logger.Info("User registered successfully", map[string]interface{}{
		"user_id":  userID,
		"username": username,
	})

	return toUserDTO(user), nil
}

// Login authenticates a user and returns a JWT token
func (s *Service) Login(ctx context.Context, username, password string) (string, error) {
	// Validate input
	if username == "" || password == "" {
		return "", errors.ValidationFailed("Username and password are required")
	}

	// Fetch user
	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return "", errors.Database(err, "Failed to fetch user")
	}
	if user == nil {
		return "", errors.InvalidCredentials("Invalid username or password")
	}

	// Check if user is active
	if !user.IsActive {
		return "", errors.Forbidden("Account is disabled")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.InvalidCredentials("Invalid username or password")
	}

	// Update last login
	if err := s.repo.UpdateLastLogin(ctx, user.ID); err != nil {
		// Log but don't fail the login
		s.logger.Warn("Failed to update last login", map[string]interface{}{
			"user_id": user.ID,
			"error":   err.Error(),
		})
	}

	// Generate JWT token
	token, err := s.generateToken(user.ID, username)
	if err != nil {
		return "", errors.Internal("Failed to generate token")
	}

	s.logger.Info("User logged in successfully", map[string]interface{}{
		"user_id":  user.ID,
		"username": username,
	})

	return token, nil
}

// GetProfile retrieves a user's profile
func (s *Service) GetProfile(ctx context.Context, userID string) (*interfaces.UserDTO, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errors.Database(err, "Failed to fetch user profile")
	}
	if user == nil {
		return nil, errors.NotFound("User not found")
	}

	return toUserDTO(user), nil
}

// UpdateProfile updates a user's profile
func (s *Service) UpdateProfile(ctx context.Context, userID string, updates *interfaces.UserUpdateDTO) error {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return errors.Database(err, "Failed to fetch user")
	}
	if user == nil {
		return errors.NotFound("User not found")
	}

	// Update email if provided
	if updates.Email != nil {
		// Check if email is already taken
		existingUser, err := s.repo.GetUserByEmail(ctx, *updates.Email)
		if err != nil {
			return errors.Database(err, "Failed to check email availability")
		}
		if existingUser != nil && existingUser.ID != userID {
			return errors.Conflict("Email already in use")
		}
		user.Email = *updates.Email
	}

	// Update password if provided
	if updates.Password != nil {
		if len(*updates.Password) < 8 {
			return errors.ValidationFailed("Password must be at least 8 characters")
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*updates.Password), bcrypt.DefaultCost)
		if err != nil {
			return errors.Internal("Failed to hash password")
		}
		user.PasswordHash = string(hashedPassword)
	}

	user.UpdatedAt = time.Now()

	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return errors.Database(err, "Failed to update user")
	}

	s.logger.Info("User profile updated", map[string]interface{}{
		"user_id": userID,
	})

	return nil
}

// DeleteAccount deletes a user account
func (s *Service) DeleteAccount(ctx context.Context, userID string) error {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return errors.Database(err, "Failed to fetch user")
	}
	if user == nil {
		return errors.NotFound("User not found")
	}

	if err := s.repo.DeleteUser(ctx, userID); err != nil {
		return errors.Database(err, "Failed to delete user")
	}

	s.logger.Info("User account deleted", map[string]interface{}{
		"user_id": userID,
	})

	return nil
}

// generateToken creates a JWT token for the user
func (s *Service) generateToken(userID, username string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// toUserDTO converts a User entity to a UserDTO
func toUserDTO(user *interfaces.User) *interfaces.UserDTO {
	return &interfaces.UserDTO{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		LastLogin: user.LastLogin,
		IsActive:  user.IsActive,
	}
}
