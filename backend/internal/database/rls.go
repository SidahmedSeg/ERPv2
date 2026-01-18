package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// TenantContext holds the tenant context for RLS operations
type TenantContext struct {
	TenantID uuid.UUID
	DB       *sqlx.DB
}

// WithTenantContext sets the RLS context for tenant isolation.
// This function starts a new transaction and sets the app.current_tenant_id
// parameter, which is used by RLS policies to enforce tenant isolation.
//
// Example usage:
//
//	tx, err := database.WithTenantContext(ctx, db, tenantID)
//	if err != nil {
//	    return err
//	}
//	defer tx.Rollback() // Rollback if not committed
//
//	// All queries within this transaction are automatically filtered by tenant_id
//	_, err = tx.ExecContext(ctx, "INSERT INTO users (...) VALUES (...)")
//	if err != nil {
//	    return err
//	}
//
//	return tx.Commit()
func WithTenantContext(ctx context.Context, db *sqlx.DB, tenantID uuid.UUID) (*sqlx.Tx, error) {
	// Start a new transaction
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Set RLS context - this parameter is used by RLS policies
	// Note: SET LOCAL doesn't support parameterized queries, so we use string formatting
	// The UUID is safe since it's validated by the uuid.UUID type
	_, err = tx.ExecContext(ctx, fmt.Sprintf("SET LOCAL app.current_tenant_id = '%s'", tenantID.String()))
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	return tx, nil
}

// WithBypassRLS temporarily bypasses RLS for administrative operations.
// This should ONLY be used for:
// - System-level operations (migrations, seeds)
// - Cross-tenant analytics (with proper authorization)
// - Administrative tools (with audit logging)
//
// WARNING: Use this function with extreme caution. All operations using this
// function should be logged in audit_logs for security compliance.
//
// Example usage:
//
//	tx, err := database.WithBypassRLS(ctx, db)
//	if err != nil {
//	    return err
//	}
//	defer tx.Rollback()
//
//	// This query can access data across all tenants
//	_, err = tx.ExecContext(ctx, "SELECT COUNT(*) FROM users")
//	if err != nil {
//	    return err
//	}
//
//	return tx.Commit()
func WithBypassRLS(ctx context.Context, db *sqlx.DB) (*sqlx.Tx, error) {
	// Start a new transaction
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Set bypass RLS flag
	_, err = tx.ExecContext(ctx, "SET LOCAL app.bypass_rls = 'true'")
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to set bypass RLS: %w", err)
	}

	return tx, nil
}

// WithTenantContextReadOnly is similar to WithTenantContext but for read-only operations.
// This can be used for SELECT queries where you want to ensure tenant isolation
// but don't need write capabilities.
func WithTenantContextReadOnly(ctx context.Context, db *sqlx.DB, tenantID uuid.UUID) (*sqlx.Tx, error) {
	// Start a transaction with read-only option
	opts := &sql.TxOptions{ReadOnly: true}
	tx, err := db.BeginTxx(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to begin read-only transaction: %w", err)
	}

	// Set RLS context
	// Note: SET LOCAL doesn't support parameterized queries, so we use string formatting
	// The UUID is safe since it's validated by the uuid.UUID type
	_, err = tx.ExecContext(ctx, fmt.Sprintf("SET LOCAL app.current_tenant_id = '%s'", tenantID.String()))
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	return tx, nil
}

// GetTenantIDFromContext extracts tenant ID from the request context.
// This is typically set by the authentication middleware after validating the JWT token.
func GetTenantIDFromContext(ctx context.Context) (uuid.UUID, error) {
	tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
	if !ok {
		return uuid.Nil, fmt.Errorf("tenant_id not found in context")
	}
	return tenantID, nil
}

// GetUserIDFromContext extracts user ID from the request context.
// This is typically set by the authentication middleware after validating the JWT token.
func GetUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	userID, ok := ctx.Value("user_id").(uuid.UUID)
	if !ok {
		return uuid.Nil, fmt.Errorf("user_id not found in context")
	}
	return userID, nil
}

// GetTenantSlugFromContext extracts tenant slug from the request context.
func GetTenantSlugFromContext(ctx context.Context) (string, error) {
	slug, ok := ctx.Value("tenant_slug").(string)
	if !ok {
		return "", fmt.Errorf("tenant_slug not found in context")
	}
	return slug, nil
}
