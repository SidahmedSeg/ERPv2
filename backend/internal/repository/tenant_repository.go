package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"myerp-v2/internal/models"
)

// TenantRepository handles database operations for tenants
type TenantRepository struct {
	db *sqlx.DB
}

// NewTenantRepository creates a new tenant repository
func NewTenantRepository(db *sqlx.DB) *TenantRepository {
	return &TenantRepository{db: db}
}

// Create creates a new tenant with verification token
func (r *TenantRepository) Create(ctx context.Context, tenant *models.Tenant) error {
	// Note: Tenants table does NOT have RLS, so no need to set tenant context
	query := `
		INSERT INTO tenants (
			slug, company_name, email, status, verification_token,
			verification_token_expires_at, plan_tier, settings
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx, query,
		tenant.Slug,
		tenant.CompanyName,
		tenant.Email,
		tenant.Status,
		tenant.VerificationToken,
		tenant.VerificationTokenExpiresAt,
		tenant.PlanTier,
		tenant.Settings,
	).Scan(&tenant.ID, &tenant.CreatedAt, &tenant.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create tenant: %w", err)
	}

	return nil
}

// FindByID retrieves a tenant by ID
func (r *TenantRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error) {
	var tenant models.Tenant
	query := `SELECT * FROM tenants WHERE id = $1`

	err := r.db.GetContext(ctx, &tenant, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tenant not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find tenant: %w", err)
	}

	return &tenant, nil
}

// FindBySlug retrieves a tenant by slug
func (r *TenantRepository) FindBySlug(ctx context.Context, slug string) (*models.Tenant, error) {
	var tenant models.Tenant
	query := `SELECT * FROM tenants WHERE slug = $1`

	err := r.db.GetContext(ctx, &tenant, query, slug)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tenant not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find tenant: %w", err)
	}

	return &tenant, nil
}

// FindByEmail retrieves a tenant by email
func (r *TenantRepository) FindByEmail(ctx context.Context, email string) (*models.Tenant, error) {
	var tenant models.Tenant
	query := `SELECT * FROM tenants WHERE email = $1`

	err := r.db.GetContext(ctx, &tenant, query, email)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tenant not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find tenant: %w", err)
	}

	return &tenant, nil
}

// FindByVerificationToken retrieves a tenant by verification token
func (r *TenantRepository) FindByVerificationToken(ctx context.Context, token uuid.UUID) (*models.Tenant, error) {
	var tenant models.Tenant
	query := `
		SELECT * FROM tenants
		WHERE verification_token = $1
		  AND verification_token_expires_at > NOW()
		  AND status = $2
	`

	err := r.db.GetContext(ctx, &tenant, query, token, models.TenantStatusPendingVerification)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("invalid or expired verification token")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find tenant: %w", err)
	}

	return &tenant, nil
}

// Update updates a tenant's information
func (r *TenantRepository) Update(ctx context.Context, tenant *models.Tenant) error {
	query := `
		UPDATE tenants
		SET company_name = $1,
		    email = $2,
		    status = $3,
		    plan_tier = $4,
		    settings = $5,
		    updated_at = NOW()
		WHERE id = $6
		RETURNING updated_at
	`

	err := r.db.QueryRowContext(
		ctx, query,
		tenant.CompanyName,
		tenant.Email,
		tenant.Status,
		tenant.PlanTier,
		tenant.Settings,
		tenant.ID,
	).Scan(&tenant.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update tenant: %w", err)
	}

	return nil
}

// VerifyEmail marks a tenant's email as verified and activates the tenant
func (r *TenantRepository) VerifyEmail(ctx context.Context, tenantID uuid.UUID) error {
	query := `
		UPDATE tenants
		SET email_verified = true,
		    email_verified_at = NOW(),
		    status = $1,
		    activated_at = NOW(),
		    verification_token = NULL,
		    verification_token_expires_at = NULL,
		    updated_at = NOW()
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, models.TenantStatusActive, tenantID)
	if err != nil {
		return fmt.Errorf("failed to verify email: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("tenant not found")
	}

	return nil
}

// UpdateStatus updates a tenant's status
func (r *TenantRepository) UpdateStatus(ctx context.Context, tenantID uuid.UUID, status string) error {
	query := `
		UPDATE tenants
		SET status = $1,
		    updated_at = NOW()
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, status, tenantID)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("tenant not found")
	}

	return nil
}

// Suspend suspends a tenant
func (r *TenantRepository) Suspend(ctx context.Context, tenantID uuid.UUID) error {
	query := `
		UPDATE tenants
		SET status = $1,
		    suspended_at = NOW(),
		    updated_at = NOW()
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, models.TenantStatusSuspended, tenantID)
	if err != nil {
		return fmt.Errorf("failed to suspend tenant: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("tenant not found")
	}

	return nil
}

// Delete deletes a tenant (soft delete by setting status to canceled)
func (r *TenantRepository) Delete(ctx context.Context, tenantID uuid.UUID) error {
	query := `
		UPDATE tenants
		SET status = $1,
		    updated_at = NOW()
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, models.TenantStatusCanceled, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("tenant not found")
	}

	return nil
}

// CheckSlugAvailability checks if a slug is available
func (r *TenantRepository) CheckSlugAvailability(ctx context.Context, slug string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM tenants WHERE slug = $1`

	err := r.db.GetContext(ctx, &count, query, slug)
	if err != nil {
		return false, fmt.Errorf("failed to check slug availability: %w", err)
	}

	return count == 0, nil
}

// CheckEmailExists checks if an email is already registered
func (r *TenantRepository) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM tenants WHERE email = $1`

	err := r.db.GetContext(ctx, &count, query, email)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return count > 0, nil
}

// List retrieves a paginated list of tenants
func (r *TenantRepository) List(ctx context.Context, limit, offset int) ([]models.Tenant, int, error) {
	var tenants []models.Tenant
	var totalCount int

	// Get total count
	countQuery := `SELECT COUNT(*) FROM tenants`
	err := r.db.GetContext(ctx, &totalCount, countQuery)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count tenants: %w", err)
	}

	// Get paginated results
	query := `
		SELECT * FROM tenants
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	err = r.db.SelectContext(ctx, &tenants, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list tenants: %w", err)
	}

	return tenants, totalCount, nil
}

// CleanupExpiredVerificationTokens removes expired verification tokens
func (r *TenantRepository) CleanupExpiredVerificationTokens(ctx context.Context) (int64, error) {
	query := `
		UPDATE tenants
		SET verification_token = NULL,
		    verification_token_expires_at = NULL
		WHERE verification_token_expires_at < NOW()
		  AND verification_token IS NOT NULL
	`

	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired tokens: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	return rowsAffected, nil
}

// ProvisionSystemRoles creates system roles for a tenant
func (r *TenantRepository) ProvisionSystemRoles(ctx context.Context, tenantID uuid.UUID) error {
	// Call the PostgreSQL function to provision system roles
	query := `SELECT provision_tenant_system_roles($1)`

	_, err := r.db.ExecContext(ctx, query, tenantID)
	if err != nil {
		return fmt.Errorf("failed to provision system roles: %w", err)
	}

	return nil
}

// RegenerateVerificationToken generates a new verification token for a tenant
func (r *TenantRepository) RegenerateVerificationToken(ctx context.Context, tenantID uuid.UUID, expiresIn time.Duration) (*uuid.UUID, error) {
	token := uuid.New()
	expiresAt := time.Now().Add(expiresIn)

	query := `
		UPDATE tenants
		SET verification_token = $1,
		    verification_token_expires_at = $2,
		    updated_at = NOW()
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query, token, expiresAt, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to regenerate verification token: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, fmt.Errorf("tenant not found")
	}

	return &token, nil
}
