// Role: PostgreSQL connection pool and infrastructure manager.
// Connects with: cmd/api/main.go (called on startup) and internal/models/task.go (builds matching schema).
// Responsibilities:
// - Initializes *sql.DB with production pool limits (max open/idle, timeouts).
// - Pings PostgreSQL defensively and runs initial table migrations for tasks.

package database

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// NewPostgresDB establishes and tunes the PostgreSQL connection pool using pgx.
func NewPostgresDB(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	// Establish connection pool quotas to prevent resource starvation under load
	db.SetConnMaxIdleTime(2 * time.Minute)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetMaxIdleConns(25)
	db.SetMaxOpenConns(25)

	// Perform a defensive ping to verify network connectivity before application startup
	if err = db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

// Migrate applies the initial idempotent database schema.
// In production environments, schema migrations are usually handled by dedicated migration tools; therefore, this direct DDL query is used for local setup and testing.
func Migrate(ctx context.Context, db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS tasks (
		id          TEXT PRIMARY KEY,
		title       TEXT NOT NULL,
		description TEXT NOT NULL,
		status      TEXT NOT NULL,
		metadata    JSONB NOT NULL,
		created_at  TIMESTAMPTZ NOT NULL,
		updated_at  TIMESTAMPTZ NOT NULL
	);`

	_, err := db.ExecContext(ctx, query)
	return err
}
