package user

import (
	"context"
	"testing"
	"time"

	"github.com/asian-code/myapp-kubernetes/services/pkg/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, username, email, passwordHash string) (string, error) {
	args := m.Called(ctx, username, email, passwordHash)
	return args.String(0), args.Error(1)
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, userID string) (*interfaces.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*interfaces.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByUsername(ctx context.Context, username string) (*interfaces.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*interfaces.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*interfaces.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*interfaces.User), args.Error(1)
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateUser(ctx context.Context, user *interfaces.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) DeleteUser(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// MockLogger is a mock implementation of Logger
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(msg string, fields map[string]interface{}) {
	m.Called(msg, fields)
}

func (m *MockLogger) Info(msg string, fields map[string]interface{}) {
	m.Called(msg, fields)
}

func (m *MockLogger) Warn(msg string, fields map[string]interface{}) {
	m.Called(msg, fields)
}

func (m *MockLogger) Error(msg string, err error, fields map[string]interface{}) {
	m.Called(msg, err, fields)
}

func (m *MockLogger) Fatal(msg string, err error, fields map[string]interface{}) {
	m.Called(msg, err, fields)
}

func (m *MockLogger) WithFields(fields map[string]interface{}) interfaces.Logger {
	args := m.Called(fields)
	return args.Get(0).(interfaces.Logger)
}

func (m *MockLogger) WithContext(ctx context.Context) interfaces.Logger {
	args := m.Called(ctx)
	return args.Get(0).(interfaces.Logger)
}

func TestRegister_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockLogger := new(MockLogger)
	service := NewService(mockRepo, "test-secret", mockLogger)

	ctx := context.Background()
	username := "testuser"
	email := "test@example.com"
	password := "password123"

	// Setup expectations
	mockRepo.On("GetUserByUsername", ctx, username).Return(nil, nil)
	mockRepo.On("GetUserByEmail", ctx, email).Return(nil, nil)
	mockRepo.On("CreateUser", ctx, username, email, mock.AnythingOfType("string")).Return("user-123", nil)
	mockRepo.On("GetUserByID", ctx, "user-123").Return(&interfaces.User{
		ID:        "user-123",
		Username:  username,
		Email:     email,
		CreatedAt: time.Now(),
		IsActive:  true,
	}, nil)
	mockLogger.On("Info", "User registered successfully", mock.Anything).Return()

	// Execute
	result, err := service.Register(ctx, username, email, password)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "user-123", result.ID)
	assert.Equal(t, username, result.Username)
	assert.Equal(t, email, result.Email)
	assert.True(t, result.IsActive)

	mockRepo.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestRegister_DuplicateUsername(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockLogger := new(MockLogger)
	service := NewService(mockRepo, "test-secret", mockLogger)

	ctx := context.Background()
	username := "existinguser"
	email := "test@example.com"
	password := "password123"

	// Setup expectations - user already exists
	mockRepo.On("GetUserByUsername", ctx, username).Return(&interfaces.User{
		ID:       "existing-123",
		Username: username,
	}, nil)

	// Execute
	result, err := service.Register(ctx, username, email, password)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Username already taken")

	mockRepo.AssertExpectations(t)
}

func TestRegister_ValidationError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockLogger := new(MockLogger)
	service := NewService(mockRepo, "test-secret", mockLogger)

	ctx := context.Background()

	tests := []struct {
		name     string
		username string
		email    string
		password string
		errMsg   string
	}{
		{
			name:     "empty username",
			username: "",
			email:    "test@example.com",
			password: "password123",
			errMsg:   "Username, email, and password are required",
		},
		{
			name:     "short password",
			username: "testuser",
			email:    "test@example.com",
			password: "short",
			errMsg:   "Password must be at least 8 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.Register(ctx, tt.username, tt.email, tt.password)
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}
