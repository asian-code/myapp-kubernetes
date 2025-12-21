package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/asian-code/myapp-kubernetes/services/pkg/errors"
	"github.com/asian-code/myapp-kubernetes/services/pkg/interfaces"
)

type service struct {
	repo         interfaces.OAuthRepository
	clientID     string
	clientSecret string
	redirectURI  string
	logger       interfaces.Logger
}

func NewService(repo interfaces.OAuthRepository, clientID, clientSecret, redirectURI string, logger interfaces.Logger) interfaces.OAuthService {
	return &service{
		repo:         repo,
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
		logger:       logger,
	}
}

// GenerateAuthURL generates the OAuth authorization URL for the user to visit
func (s *service) GenerateAuthURL(ctx context.Context, userID string) (string, error) {
	if userID == "" {
		return "", errors.BadRequest("user ID is required")
	}

	// Oura OAuth 2.0 authorization endpoint
	authURL := "https://cloud.ouraring.com/oauth/authorize"
	
	params := url.Values{}
	params.Add("response_type", "code")
	params.Add("client_id", s.clientID)
	params.Add("redirect_uri", s.redirectURI)
	params.Add("scope", "daily personal email")
	params.Add("state", userID) // Use userID as state for CSRF protection

	fullURL := authURL + "?" + params.Encode()

	s.logger.Info("Generated OAuth authorization URL", map[string]interface{}{
		"user_id": userID,
	})

	return fullURL, nil
}

// HandleCallback processes the OAuth callback and exchanges the authorization code for tokens
func (s *service) HandleCallback(ctx context.Context, code, state string) (*interfaces.OAuthResult, error) {
	if code == "" {
		return nil, errors.BadRequest("authorization code is required")
	}
	if state == "" {
		return nil, errors.BadRequest("state is required")
	}

	userID := state // state contains the userID

	// Exchange authorization code for access token
	tokenResp, err := s.exchangeCodeForToken(ctx, code)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to exchange code for token")
	}

	// Save token to database
	token := &interfaces.OAuthToken{
		UserID:       userID,
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
		Scope:        tokenResp.Scope,
		Provider:     "oura",
	}

	err = s.repo.SaveToken(ctx, token)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to save token")
	}

	s.logger.Info("Successfully saved OAuth token", map[string]interface{}{
		"user_id": userID,
		"expires_at": token.ExpiresAt,
	})

	return &interfaces.OAuthResult{
		UserID:      userID,
		AccessToken: token.AccessToken,
		ExpiresAt:   token.ExpiresAt,
	}, nil
}

// RefreshAccessToken refreshes an expired access token using the refresh token
func (s *service) RefreshAccessToken(ctx context.Context, userID, provider string) error {
	if userID == "" {
		return errors.BadRequest("user ID is required")
	}
	if provider == "" {
		return errors.BadRequest("provider is required")
	}

	// Get existing token from database
	existingToken, err := s.repo.GetToken(ctx, userID, provider)
	if err != nil {
		return errors.Wrap(err, errors.ErrCodeInternal, "failed to get token")
	}
	if existingToken == nil {
		return errors.NotFound("no OAuth token found for user")
	}

	// Check if token is still valid (not expired)
	if time.Now().Before(existingToken.ExpiresAt) {
		s.logger.Info("Token still valid, no refresh needed", map[string]interface{}{
			"user_id": userID,
			"expires_at": existingToken.ExpiresAt,
		})
		return nil
	}

	// Token is expired, refresh it
	tokenResp, err := s.refreshToken(ctx, existingToken.RefreshToken)
	if err != nil {
		return errors.Wrap(err, errors.ErrCodeInternal, "failed to refresh token")
	}

	// Update token in database using RefreshToken method
	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	err = s.repo.RefreshToken(ctx, userID, provider, tokenResp.AccessToken, tokenResp.RefreshToken, expiresAt)
	if err != nil {
		return errors.Wrap(err, errors.ErrCodeInternal, "failed to update token")
	}

	s.logger.Info("Successfully refreshed OAuth token", map[string]interface{}{
		"user_id": userID,
		"expires_at": expiresAt,
	})

	return nil
}

// RevokeToken revokes the OAuth token and removes it from the database
func (s *service) RevokeToken(ctx context.Context, userID, provider string) error {
	if userID == "" {
		return errors.BadRequest("user ID is required")
	}
	if provider == "" {
		return errors.BadRequest("provider is required")
	}

	// Get token from database
	token, err := s.repo.GetToken(ctx, userID, provider)
	if err != nil {
		return errors.Wrap(err, errors.ErrCodeInternal, "failed to get token")
	}
	if token == nil {
		return errors.NotFound("no OAuth token found for user")
	}

	// Revoke token with OAuth provider
	err = s.revokeTokenWithProvider(ctx, token.AccessToken)
	if err != nil {
		s.logger.Error("failed to revoke token with provider", err, map[string]interface{}{
			"user_id": userID,
		})
		// Continue to delete from database even if provider revocation fails
	}

	// Delete token from database
	err = s.repo.DeleteToken(ctx, userID, provider)
	if err != nil {
		return errors.Wrap(err, errors.ErrCodeInternal, "failed to delete token")
	}

	s.logger.Info("Successfully revoked OAuth token", map[string]interface{}{
		"user_id": userID,
	})

	return nil
}

// exchangeCodeForToken exchanges an authorization code for an access token
func (s *service) exchangeCodeForToken(ctx context.Context, code string) (*tokenResponse, error) {
	tokenURL := "https://api.ouraring.com/oauth/token"

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", s.redirectURI)
	data.Set("client_id", s.clientID)
	data.Set("client_secret", s.clientSecret)

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp tokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &tokenResp, nil
}

// refreshToken refreshes an access token using a refresh token
func (s *service) refreshToken(ctx context.Context, refreshToken string) (*tokenResponse, error) {
	tokenURL := "https://api.ouraring.com/oauth/token"

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	data.Set("client_id", s.clientID)
	data.Set("client_secret", s.clientSecret)

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token refresh failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp tokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &tokenResp, nil
}

// revokeTokenWithProvider revokes a token with the OAuth provider
func (s *service) revokeTokenWithProvider(ctx context.Context, token string) error {
	// Note: Oura API may not have a dedicated revoke endpoint
	// This is a placeholder implementation
	revokeURL := "https://api.ouraring.com/oauth/revoke"

	data := url.Values{}
	data.Set("token", token)
	data.Set("client_id", s.clientID)
	data.Set("client_secret", s.clientSecret)

	req, err := http.NewRequestWithContext(ctx, "POST", revokeURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("token revocation failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// tokenResponse represents the response from the OAuth token endpoint
type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
}
