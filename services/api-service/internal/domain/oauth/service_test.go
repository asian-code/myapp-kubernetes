package oauth

import (
	"context"
	"testing"
	"time"

	"github.com/asian-code/myapp-kubernetes/services/pkg/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockOAuthRepository is a mock implementation of interfaces.OAuthRepository
type MockOAuthRepository struct {
	mock.Mock
}

func (m *MockOAuthRepository) SaveToken(ctx context.Context, token *interfaces.OAuthToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockOAuthRepository) GetToken(ctx context.Context, userID, provider string) (*interfaces.OAuthToken, error) {
	args := m.Called(ctx, userID, provider)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*interfaces.OAuthToken), args.Error(1)
}

func (m *MockOAuthRepository) UpdateToken(ctx context.Context, token *interfaces.OAuthToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockOAuthRepository) DeleteToken(ctx context.Context, userID, provider string) error {
	args := m.Called(ctx, userID, provider)
	return args.Error(0)
}

func (m *MockOAuthRepository) RefreshToken(ctx context.Context, userID, provider string, newAccessToken, newRefreshToken string, expiresAt time.Time) error {
	args := m.Called(ctx, userID, provider, newAccessToken, newRefreshToken, expiresAt)
	return args.Error(0)
}

// MockLogger is a simple mock implementation of interfaces.Logger
type MockLogger struct{}

func (m *MockLogger) Debug(msg string, fields map[string]interface{})                  {}
func (m *MockLogger) Info(msg string, fields map[string]interface{})                   {}
func (m *MockLogger) Warn(msg string, fields map[string]interface{})                   {}
func (m *MockLogger) Error(msg string, err error, fields map[string]interface{})       {}
func (m *MockLogger) Fatal(msg string, err error, fields map[string]interface{})       {}
func (m *MockLogger) WithFields(fields map[string]interface{}) interfaces.Logger       { return m }
func (m *MockLogger) WithContext(ctx context.Context) interfaces.Logger                { return m }

func TestOAuthService_GenerateAuthURL_Success(t *testing.T) {
	mockRepo := new(MockOAuthRepository)
	mockLogger := &MockLogger{}
	
	service := NewService(mockRepo, "test-client-id", "test-secret", "http://localhost/callback", mockLogger)
	
	authURL, err := service.GenerateAuthURL(context.Background(), "user-123")
	
	assert.NoError(t, err)
	assert.NotEmpty(t, authURL)
	assert.Contains(t, authURL, "https://cloud.ouraring.com/oauth/authorize")
	assert.Contains(t, authURL, "client_id=test-client-id")
	assert.Contains(t, authURL, "redirect_uri=http")
	assert.Contains(t, authURL, "state=user-123")
	assert.Contains(t, authURL, "scope=daily+personal+email")
}

func TestOAuthService_GenerateAuthURL_MissingUserID(t *testing.T) {
	mockRepo := new(MockOAuthRepository)
	mockLogger := &MockLogger{}
	
	service := NewService(mockRepo, "test-client-id", "test-secret", "http://localhost/callback", mockLogger)
	
	_, err := service.GenerateAuthURL(context.Background(), "")
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user ID is required")
}

func TestOAuthService_RefreshAccessToken_TokenStillValid(t *testing.T) {
	mockRepo := new(MockOAuthRepository)
	mockLogger := &MockLogger{}
	
	// Token that expires in the future
	futureExpiry := time.Now().Add(1 * time.Hour)
	existingToken := &interfaces.OAuthToken{
		UserID:       "user-123",
		AccessToken:  "existing-access-token",
		RefreshToken: "existing-refresh-token",
		ExpiresAt:    futureExpiry,
		Scope:        "daily personal",
		Provider:     "oura",
	}
	
	mockRepo.On("GetToken", mock.Anything, "user-123", "oura").Return(existingToken, nil)
	
	service := NewService(mockRepo, "test-client-id", "test-secret", "http://localhost/callback", mockLogger)
	
	err := service.RefreshAccessToken(context.Background(), "user-123", "oura")
	
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestOAuthService_RefreshAccessToken_NoTokenFound(t *testing.T) {
	mockRepo := new(MockOAuthRepository)
	mockLogger := &MockLogger{}
	
	mockRepo.On("GetToken", mock.Anything, "user-123", "oura").Return(nil, nil)
	
	service := NewService(mockRepo, "test-client-id", "test-secret", "http://localhost/callback", mockLogger)
	
	err := service.RefreshAccessToken(context.Background(), "user-123", "oura")
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no OAuth token found")
	mockRepo.AssertExpectations(t)
}

func TestOAuthService_RefreshAccessToken_MissingUserID(t *testing.T) {
	mockRepo := new(MockOAuthRepository)
	mockLogger := &MockLogger{}
	
	service := NewService(mockRepo, "test-client-id", "test-secret", "http://localhost/callback", mockLogger)
	
	err := service.RefreshAccessToken(context.Background(), "", "oura")
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user ID is required")
}

func TestOAuthService_RevokeToken_Success(t *testing.T) {
	mockRepo := new(MockOAuthRepository)
	mockLogger := &MockLogger{}
	
	existingToken := &interfaces.OAuthToken{
		UserID:       "user-123",
		AccessToken:  "access-token-to-revoke",
		RefreshToken: "refresh-token",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		Scope:        "daily",
		Provider:     "oura",
	}
	
	mockRepo.On("GetToken", mock.Anything, "user-123", "oura").Return(existingToken, nil)
	mockRepo.On("DeleteToken", mock.Anything, "user-123", "oura").Return(nil)
	
	service := NewService(mockRepo, "test-client-id", "test-secret", "http://localhost/callback", mockLogger)
	
	err := service.RevokeToken(context.Background(), "user-123", "oura")
	
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestOAuthService_RevokeToken_NoTokenFound(t *testing.T) {
	mockRepo := new(MockOAuthRepository)
	mockLogger := &MockLogger{}
	
	mockRepo.On("GetToken", mock.Anything, "user-123", "oura").Return(nil, nil)
	
	service := NewService(mockRepo, "test-client-id", "test-secret", "http://localhost/callback", mockLogger)
	
	err := service.RevokeToken(context.Background(), "user-123", "oura")
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no OAuth token found")
	mockRepo.AssertExpectations(t)
}

func TestOAuthService_RevokeToken_MissingUserID(t *testing.T) {
	mockRepo := new(MockOAuthRepository)
	mockLogger := &MockLogger{}
	
	service := NewService(mockRepo, "test-client-id", "test-secret", "http://localhost/callback", mockLogger)
	
	err := service.RevokeToken(context.Background(), "", "oura")
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user ID is required")
}
