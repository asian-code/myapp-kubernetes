package interfaces

import (
	"context"
	"net/http"
)

// HTTPHandler defines the interface for HTTP request handlers
type HTTPHandler interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

// Router defines operations for HTTP routing
type Router interface {
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) Router
	Handle(pattern string, handler http.Handler) Router
	Use(middleware func(http.Handler) http.Handler)
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

// Middleware defines the signature for HTTP middleware
type Middleware func(http.Handler) http.Handler

// HealthChecker defines operations for health checking
type HealthChecker interface {
	Check(ctx context.Context) error
	Name() string
}

// Cache defines operations for caching
type Cache interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, ttl int) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Clear(ctx context.Context) error
}

// TokenGenerator defines operations for generating authentication tokens
type TokenGenerator interface {
	GenerateToken(userID string, claims map[string]interface{}) (string, error)
	ValidateToken(token string) (map[string]interface{}, error)
	RefreshToken(oldToken string) (string, error)
}

// PasswordHasher defines operations for password hashing
type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword, password string) error
}

// EventPublisher defines operations for publishing domain events
type EventPublisher interface {
	Publish(ctx context.Context, event *DomainEvent) error
}

// EventSubscriber defines operations for subscribing to domain events
type EventSubscriber interface {
	Subscribe(ctx context.Context, eventType string, handler func(*DomainEvent) error) error
}

// DomainEvent represents an event that occurred in the domain
type DomainEvent struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Timestamp int64                  `json:"timestamp"`
	UserID    string                 `json:"user_id,omitempty"`
	Data      map[string]interface{} `json:"data"`
}
