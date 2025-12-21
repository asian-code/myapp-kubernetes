package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// PostgresContainer wraps a testcontainers PostgreSQL instance
type PostgresContainer struct {
	Container testcontainers.Container
	ConnString string
}

// SetupPostgresContainer creates a PostgreSQL container for testing
func SetupPostgresContainer(ctx context.Context) (*PostgresContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("database system is ready to accept connections"),
			wait.ForListeningPort("5432/tcp"),
		).WithDeadline(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get container host: %w", err)
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return nil, fmt.Errorf("failed to get container port: %w", err)
	}

	connString := fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable", host, port.Port())

	return &PostgresContainer{
		Container:  container,
		ConnString: connString,
	}, nil
}

// Close terminates the container
func (pc *PostgresContainer) Close(ctx context.Context) error {
	return pc.Container.Terminate(ctx)
}

// GetPool creates a connection pool to the test database
func (pc *PostgresContainer) GetPool(ctx context.Context) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(pc.ConnString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	// Wait for connection to be ready
	for i := 0; i < 10; i++ {
		if err := pool.Ping(ctx); err == nil {
			return pool, nil
		}
		time.Sleep(500 * time.Millisecond)
	}

	return nil, fmt.Errorf("failed to connect to database after retries")
}

// RunMigrations executes migration scripts on the test database
func (pc *PostgresContainer) RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migrations := []string{
		// Users table
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			username VARCHAR(255) UNIQUE NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			last_login TIMESTAMP,
			is_active BOOLEAN DEFAULT true
		)`,
		`CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,

		// OAuth tokens table
		`CREATE TABLE IF NOT EXISTS oauth_tokens (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			provider VARCHAR(50) NOT NULL DEFAULT 'oura',
			access_token TEXT NOT NULL,
			refresh_token TEXT NOT NULL,
			token_type VARCHAR(50) DEFAULT 'Bearer',
			expires_at TIMESTAMP NOT NULL,
			scope TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, provider)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_oauth_tokens_user_provider ON oauth_tokens(user_id, provider)`,

		// Sleep metrics table
		`CREATE TABLE IF NOT EXISTS sleep_metrics (
			id SERIAL PRIMARY KEY,
			oura_id VARCHAR(255) UNIQUE NOT NULL,
			day DATE NOT NULL,
			score INTEGER,
			duration INTEGER,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_sleep_day ON sleep_metrics(day)`,

		// Activity metrics table
		`CREATE TABLE IF NOT EXISTS activity_metrics (
			id SERIAL PRIMARY KEY,
			oura_id VARCHAR(255) UNIQUE NOT NULL,
			day DATE NOT NULL,
			score INTEGER,
			active_calories INTEGER,
			steps INTEGER,
			medium_activity_minutes INTEGER,
			high_activity_minutes INTEGER,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_activity_day ON activity_metrics(day)`,

		// Readiness metrics table
		`CREATE TABLE IF NOT EXISTS readiness_metrics (
			id SERIAL PRIMARY KEY,
			oura_id VARCHAR(255) UNIQUE NOT NULL,
			day DATE NOT NULL,
			score INTEGER,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_readiness_day ON readiness_metrics(day)`,
	}

	for _, migration := range migrations {
		if _, err := pool.Exec(ctx, migration); err != nil {
			return fmt.Errorf("failed to execute migration: %w", err)
		}
	}

	return nil
}

// TestMain can be used to set up and tear down containers for all tests
func TestMain(m *testing.M) {
	// This would be used in actual test files
	// For now, it's just a placeholder
	m.Run()
}
