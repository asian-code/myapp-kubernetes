# Go Coding Standards & Quick Start

This guide outlines the coding standards and quick start steps for developers working on the myHealth microservices.

## 1. Environment Setup

```bash
cd services

# Update all modules
cd pkg && go mod tidy
cd ../api-service && go mod tidy
cd ../data-processor && go mod tidy
cd ../oura-collector && go mod tidy
cd ../migrator && go mod tidy
```

### 2. Using the New Error Handling

**Import the package:**
```go
import apperrors "github.com/asian-code/myapp-kubernetes/services/pkg/errors"
```

**Common patterns:**
```go
// In your HTTP handler
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
    userID := mux.Vars(r)["id"]
    
    user, err := h.repo.GetUser(ctx, userID)
    if err != nil {
        // Database error with context
        apperrors.WriteError(w, h.logger, 
            apperrors.Database(err, "Failed to fetch user"))
        return
    }
    
    if user == nil {
        // Not found
        apperrors.WriteError(w, h.logger, 
            apperrors.NotFound("User not found"))
        return
    }
    
    // Success response
    apperrors.WriteSuccess(w, user, http.StatusOK)
}
```

**Error types available:**
- `apperrors.BadRequest(msg)` - 400
- `apperrors.Unauthorized(msg)` - 401
- `apperrors.Forbidden(msg)` - 403
- `apperrors.NotFound(msg)` - 404
- `apperrors.Conflict(msg)` - 409
- `apperrors.ValidationFailed(msg)` - 400
- `apperrors.Internal(msg)` - 500
- `apperrors.Database(err, msg)` - 500 (wraps error)
- `apperrors.ExternalService(err, msg)` - 500 (wraps error)

**Adding details:**
```go
err := apperrors.ValidationFailed("Invalid input").
    WithDetails("field", "email").
    WithDetails("reason", "format is invalid")
apperrors.WriteError(w, logger, err)
```

### 3. Creating New Configuration

**Add validation tags:**
```go
type MyConfig struct {
    APIKey    string `validate:"required,min=10"`
    Timeout   int    `validate:"required,min=1,max=300"`
    LogLevel  string `validate:"required,oneof=debug info warn error"`
    Endpoint  string `validate:"required,url"`
}

func Load() *MyConfig {
    cfg := &MyConfig{
        APIKey:   os.Getenv("API_KEY"),
        Timeout:  parseInt(os.Getenv("TIMEOUT")),
        LogLevel: getEnv("LOG_LEVEL", "info"),
        Endpoint: os.Getenv("ENDPOINT"),
    }
    
    // This will panic if validation fails (good for startup)
    validation.MustValidate(cfg)
    
    return cfg
}
```

**Available validation tags:**
- `required` - Field must be set
- `min=N`, `max=N` - Numeric/string length limits
- `oneof=a b c` - Must be one of the values
- `url` - Must be valid URL
- `email` - Must be valid email
- `ip` - Must be valid IP address

See https://pkg.go.dev/github.com/go-playground/validator/v10 for more.

### 4. Database Migrations (Never Use InitSchema!)

**Creating a new migration:**

```bash
# Choose next version number (check existing files)
cd services/migrations

# Create both up and down files
touch 000006_add_user_preferences.up.sql
touch 000006_add_user_preferences.down.sql
```

**000006_add_user_preferences.up.sql:**
```sql
-- Add new columns
ALTER TABLE users ADD COLUMN preferences JSONB DEFAULT '{}';
CREATE INDEX idx_users_preferences ON users USING GIN (preferences);
```

**000006_add_user_preferences.down.sql:**
```sql
-- Rollback changes
DROP INDEX IF EXISTS idx_users_preferences;
ALTER TABLE users DROP COLUMN IF EXISTS preferences;
```

**Testing migrations locally:**
```bash
cd services/migrator

export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=myhealth_user
export DB_PASSWORD=your_password
export DB_NAME=myhealth
export DB_SSLMODE=disable
export MIGRATIONS_PATH=file://../migrations

go run main.go
```

### 5. Adding Middleware to Your Service

```go
import "github.com/asian-code/myapp-kubernetes/services/pkg/middleware"

func main() {
    // ... setup code ...
    
    router := mux.NewRouter()
    
    // Apply middleware (order matters!)
    router.Use(middleware.RequestID())              // Add request ID
    router.Use(middleware.RequestLogger(logger))    // Log requests
    router.Use(middleware.ErrorHandler(logger))     // Catch panics
    
    // Your routes
    router.HandleFunc("/api/endpoint", handler).Methods("GET")
    
    // ... rest of code ...
}
```

### 6. Writing Tests

**Config test example:**
```go
func TestLoad_Success(t *testing.T) {
    os.Setenv("REQUIRED_VAR", "value")
    defer os.Unsetenv("REQUIRED_VAR")
    
    cfg := Load()
    
    if cfg.RequiredVar != "value" {
        t.Errorf("expected %s, got %s", "value", cfg.RequiredVar)
    }
}
```

**Handler test example (coming in Phase 2 - with mocks):**
```go
// Future: Will use interfaces and mocks
// For now, test with real database using testcontainers
```

### 7. Running Tests

```bash
# Test a specific package
cd services/pkg/errors
go test -v

# Test with coverage
go test -cover ./...

# Test all services
cd services/api-service && go test ./...
cd services/data-processor && go test ./...
```

### 8. Common Mistakes to Avoid

❌ **Don't:**
```go
// Manual validation
if cfg.DBPassword == "" {
    log.Fatal("DB_PASSWORD required")
}

// Creating tables in code
db.Exec("CREATE TABLE IF NOT EXISTS ...")

// Generic error responses
http.Error(w, "error", 500)

// Ignoring context
func MyFunction() error {
    // Missing context parameter
}
```

✅ **Do:**
```go
// Use automatic validation
cfg := config.Load() // Panics if invalid

// Use migrations
// See migrations/ directory

// Use structured errors
apperrors.WriteError(w, logger, 
    apperrors.Database(err, "Failed to query"))

// Pass context everywhere
func MyFunction(ctx context.Context) error {
    // Can be cancelled, has timeouts
}
```

### 9. Debugging Tips

**Check config validation:**
```bash
# Service will exit immediately with clear error
kubectl logs pod/api-service-xxx

# Look for: "Configuration validation failed: DB_PASSWORD is required"
```

**Check migrations:**
```bash
# View migration job logs
kubectl logs job/myhealth-migrator

# Check if migrations ran
kubectl get jobs -n myhealth
```

**Trace requests across services:**
```bash
# All logs include request_id
kubectl logs pod/api-service-xxx | grep "request_id=20231221120000"
```

### 10. IDE Setup Recommendations

**VS Code extensions:**
- Go (official)
- Go Test Explorer
- Error Lens (shows validation errors inline)

**Settings:**
```json
{
    "go.testFlags": ["-v"],
    "go.coverOnSave": true,
    "go.lintTool": "golangci-lint"
}
```

---

## Quick Reference Card

### Error Handling
```go
import apperrors "github.com/asian-code/myapp-kubernetes/services/pkg/errors"
apperrors.WriteError(w, logger, apperrors.NotFound("Not found"))
```

### Config Validation
```go
import "github.com/asian-code/myapp-kubernetes/services/pkg/validation"
validation.MustValidate(cfg) // In Load() function
```

### Middleware
```go
import "github.com/asian-code/myapp-kubernetes/services/pkg/middleware"
router.Use(middleware.ErrorHandler(logger))
```

### Migrations
```bash
# Create: services/migrations/000XXX_description.{up,down}.sql
# Test: cd services/migrator && go run main.go
```

---

## Getting Help

1. **Check tests:** Most packages have `*_test.go` showing usage
2. **Read docs:** See `docs/PHASE_1_REFACTORING.md`
3. **View examples:** Look at existing handlers in `api-service`
4. **Ask team:** Share learnings in team chat

---

## Next Steps for New Features

When adding a new feature:

1. ✅ Add config fields with validation tags
2. ✅ Use apperrors for all error responses
3. ✅ Create migrations for schema changes
4. ✅ Write tests for business logic
5. ✅ Apply middleware to routes
6. ✅ Use context everywhere
7. ✅ Log with structured fields

---

**Happy coding! 🚀**
