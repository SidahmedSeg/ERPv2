package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
	"myerp-v2/internal/config"
)

// NewPostgresDB creates a new PostgreSQL database connection with connection pooling
func NewPostgresDB(cfg *config.DatabaseConfig) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)       // Maximum number of open connections
	db.SetMaxIdleConns(cfg.MaxIdleConns)       // Maximum number of idle connections
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime) // Maximum lifetime of a connection
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime) // Maximum idle time before closing

	// Verify connection is working
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// Close gracefully closes the database connection
func Close(db *sqlx.DB) error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// HealthCheck performs a health check on the database connection
func HealthCheck(ctx context.Context, db *sqlx.DB) error {
	// Set a timeout for the health check
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// Ping the database
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	// Check if we can execute a simple query
	var result int
	err := db.GetContext(ctx, &result, "SELECT 1")
	if err != nil {
		return fmt.Errorf("database query health check failed: %w", err)
	}

	return nil
}

// Stats returns database connection pool statistics
func Stats(db *sqlx.DB) map[string]interface{} {
	stats := db.Stats()
	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}
}

// ExecInTransaction executes a function within a database transaction.
// If the function returns an error, the transaction is rolled back.
// Otherwise, it is committed.
func ExecInTransaction(ctx context.Context, db *sqlx.DB, fn func(*sqlx.Tx) error) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			// Rollback on panic
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx error: %v, rollback error: %w", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ExecInTenantTransaction executes a function within a tenant-scoped transaction.
// This automatically sets the RLS context before executing the function.
func ExecInTenantTransaction(ctx context.Context, db *sqlx.DB, tenantID string, fn func(*sqlx.Tx) error) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	// Set RLS context
	_, err = tx.ExecContext(ctx, "SET LOCAL app.current_tenant_id = $1", tenantID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to set tenant context: %w", err)
	}

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx error: %v, rollback error: %w", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
