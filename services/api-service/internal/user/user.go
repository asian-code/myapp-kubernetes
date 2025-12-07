package user

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/asian-code/myapp-kubernetes/services/api-service/internal/auth"
	"github.com/jackc/pgx/v5/pgxpool"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	db        *pgxpool.Pool
	jwtSecret string
	logger    *log.Entry
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token  string `json:"token"`
	UserID string `json:"user_id"`
}

func NewHandler(db *pgxpool.Pool, jwtSecret string, logger *log.Entry) *Handler {
	return &Handler{
		db:        db,
		jwtSecret: jwtSecret,
		logger:    logger,
	}
}

// Register creates a new user account
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to decode request")
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Username == "" || req.Email == "" || req.Password == "" {
		http.Error(w, "Username, email, and password are required", http.StatusBadRequest)
		return
	}

	if len(req.Password) < 8 {
		http.Error(w, "Password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.logger.WithError(err).Error("Failed to hash password")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Insert user into database
	ctx := context.Background()
	var userID string
	query := `
		INSERT INTO users (username, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	err = h.db.QueryRow(ctx, query, req.Username, req.Email, string(hashedPassword)).Scan(&userID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create user")
		// Check for duplicate username/email
		http.Error(w, "Username or email already exists", http.StatusConflict)
		return
	}

	// Generate JWT token
	token, err := auth.GenerateToken(userID, h.jwtSecret)
	if err != nil {
		h.logger.WithError(err).Error("Failed to generate token")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.WithField("user_id", userID).Info("User registered successfully")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(AuthResponse{
		Token:  token,
		UserID: userID,
	})
}

// Login authenticates a user and returns a JWT token
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to decode request")
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Username == "" || req.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	// Get user from database
	ctx := context.Background()
	var userID, passwordHash string
	var isActive bool
	query := `SELECT id, password_hash, is_active FROM users WHERE username = $1`
	err := h.db.QueryRow(ctx, query, req.Username).Scan(&userID, &passwordHash, &isActive)
	if err != nil {
		h.logger.WithError(err).Error("User not found")
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Check if user is active
	if !isActive {
		http.Error(w, "Account is disabled", http.StatusForbidden)
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		h.logger.WithError(err).Error("Invalid password")
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Update last login time
	updateQuery := `UPDATE users SET last_login = $1 WHERE id = $2`
	_, err = h.db.Exec(ctx, updateQuery, time.Now(), userID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update last login")
		// Non-critical error, continue
	}

	// Generate JWT token
	token, err := auth.GenerateToken(userID, h.jwtSecret)
	if err != nil {
		h.logger.WithError(err).Error("Failed to generate token")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.WithField("user_id", userID).Info("User logged in successfully")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthResponse{
		Token:  token,
		UserID: userID,
	})
}

// Me returns the current user's information
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	// Get user ID from JWT context (set by auth middleware)
	userID := r.Context().Value("user_id").(string)

	// Get user details from database
	ctx := context.Background()
	var username, email string
	var createdAt, lastLogin time.Time
	query := `SELECT username, email, created_at, last_login FROM users WHERE id = $1`
	err := h.db.QueryRow(ctx, query, userID).Scan(&username, &email, &createdAt, &lastLogin)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get user")
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id":    userID,
		"username":   username,
		"email":      email,
		"created_at": createdAt,
		"last_login": lastLogin,
	})
}
