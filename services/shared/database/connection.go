package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	MaxConns int
	SSLMode  string // "disable", "require", "verify-ca", "verify-full"
}

func NewPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	// Default to 'require' for security, allow override for local dev
	sslMode := cfg.SSLMode
	if sslMode == "" {
		sslMode = "require"
	}

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database, sslMode,
	)

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	poolCfg.MaxConns = int32(cfg.MaxConns)

	return pgxpool.NewWithConfig(ctx, poolCfg)
}
