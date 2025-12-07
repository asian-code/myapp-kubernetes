package oauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	log "github.com/sirupsen/logrus"
)

const (
	OuraAuthURL  = "https://cloud.ouraring.com/oauth/authorize"
	OuraTokenURL = "https://api.ouraring.com/oauth/token"
)

type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

type Handler struct {
	db     *pgxpool.Pool
	config Config
	logger *log.Entry
}

type OuraTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

func NewHandler(db *pgxpool.Pool, config Config, logger *log.Entry) *Handler {
	return &Handler{
		db:     db,
		config: config,
		logger: logger,
	}
}

// GenerateState creates a random state token for CSRF protection
func GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// Authorize initiates the OAuth2 flow
func (h *Handler) Authorize(w http.ResponseWriter, r *http.Request) {
	state, err := GenerateState()
	if err != nil {
		h.logger.WithError(err).Error("Failed to generate state")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Store state in session/cookie for verification (simplified - in production use proper session management)
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600, // 10 minutes
	})

	authURL := fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&response_type=code&state=%s&scope=daily",
		OuraAuthURL,
		url.QueryEscape(h.config.ClientID),
		url.QueryEscape(h.config.RedirectURI),
		url.QueryEscape(state),
	)

	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// Callback handles the OAuth2 callback from Oura
func (h *Handler) Callback(w http.ResponseWriter, r *http.Request) {
	// Verify state to prevent CSRF
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil {
		h.logger.WithError(err).Error("State cookie not found")
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	state := r.URL.Query().Get("state")
	if state != stateCookie.Value {
		h.logger.Error("State mismatch")
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	// Exchange authorization code for tokens
	code := r.URL.Query().Get("code")
	if code == "" {
		h.logger.Error("No authorization code received")
		http.Error(w, "No authorization code", http.StatusBadRequest)
		return
	}

	tokens, err := h.exchangeCode(code)
	if err != nil {
		h.logger.WithError(err).Error("Failed to exchange code for tokens")
		http.Error(w, "Failed to get tokens", http.StatusInternalServerError)
		return
	}

	// Get user ID from session/JWT (simplified - assumes user is authenticated)
	// In production, get this from the authenticated session
	userID := r.URL.Query().Get("user_id") // Temporary - should come from JWT
	if userID == "" {
		h.logger.Error("No user_id provided")
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Store tokens in database
	ctx := context.Background()
	expiresAt := time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second)

	query := `
		INSERT INTO oauth_tokens (user_id, provider, access_token, refresh_token, token_type, expires_at, scope)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id, provider) 
		DO UPDATE SET 
			access_token = EXCLUDED.access_token,
			refresh_token = EXCLUDED.refresh_token,
			expires_at = EXCLUDED.expires_at,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err = h.db.Exec(ctx, query, userID, "oura", tokens.AccessToken, tokens.RefreshToken, tokens.TokenType, expiresAt, tokens.Scope)
	if err != nil {
		h.logger.WithError(err).Error("Failed to store tokens")
		http.Error(w, "Failed to store tokens", http.StatusInternalServerError)
		return
	}

	h.logger.WithField("user_id", userID).Info("OAuth tokens stored successfully")

	// Redirect to success page
	http.Redirect(w, r, "/oauth/success", http.StatusTemporaryRedirect)
}

// exchangeCode exchanges authorization code for access and refresh tokens
func (h *Handler) exchangeCode(code string) (*OuraTokenResponse, error) {
	data := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {h.config.RedirectURI},
		"client_id":     {h.config.ClientID},
		"client_secret": {h.config.ClientSecret},
	}

	resp, err := http.PostForm(OuraTokenURL, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token exchange failed: %d - %s", resp.StatusCode, string(body))
	}

	var tokens OuraTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		return nil, err
	}

	return &tokens, nil
}

// RefreshToken refreshes an expired access token
func (h *Handler) RefreshToken(ctx context.Context, userID string) error {
	// Get current refresh token
	var refreshToken string
	query := `SELECT refresh_token FROM oauth_tokens WHERE user_id = $1 AND provider = 'oura'`
	err := h.db.QueryRow(ctx, query, userID).Scan(&refreshToken)
	if err != nil {
		return fmt.Errorf("failed to get refresh token: %w", err)
	}

	// Exchange refresh token for new tokens
	data := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
		"client_id":     {h.config.ClientID},
		"client_secret": {h.config.ClientSecret},
	}

	resp, err := http.PostForm(OuraTokenURL, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("token refresh failed: %d - %s", resp.StatusCode, string(body))
	}

	var tokens OuraTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		return err
	}

	// Update tokens in database
	expiresAt := time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second)
	updateQuery := `
		UPDATE oauth_tokens 
		SET access_token = $1, refresh_token = $2, expires_at = $3, updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $4 AND provider = 'oura'
	`
	_, err = h.db.Exec(ctx, updateQuery, tokens.AccessToken, tokens.RefreshToken, expiresAt, userID)
	if err != nil {
		return fmt.Errorf("failed to update tokens: %w", err)
	}

	h.logger.WithField("user_id", userID).Info("OAuth tokens refreshed successfully")
	return nil
}

// GetValidToken returns a valid access token, refreshing if necessary
func (h *Handler) GetValidToken(ctx context.Context, userID string) (string, error) {
	var accessToken string
	var expiresAt time.Time

	query := `SELECT access_token, expires_at FROM oauth_tokens WHERE user_id = $1 AND provider = 'oura'`
	err := h.db.QueryRow(ctx, query, userID).Scan(&accessToken, &expiresAt)
	if err != nil {
		return "", fmt.Errorf("failed to get token: %w", err)
	}

	// Check if token is expired or will expire in the next 5 minutes
	if time.Now().Add(5 * time.Minute).After(expiresAt) {
		h.logger.WithField("user_id", userID).Info("Token expired or expiring soon, refreshing")
		if err := h.RefreshToken(ctx, userID); err != nil {
			return "", err
		}

		// Fetch the new token
		err = h.db.QueryRow(ctx, query, userID).Scan(&accessToken, &expiresAt)
		if err != nil {
			return "", fmt.Errorf("failed to get refreshed token: %w", err)
		}
	}

	return accessToken, nil
}
