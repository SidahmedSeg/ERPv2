package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"myerp-v2/internal/database"
	"myerp-v2/internal/models"
)

// UserRepository handles database operations for users
type UserRepository struct {
	db *sqlx.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user with RLS tenant context
func (r *UserRepository) Create(ctx context.Context, tenantID uuid.UUID, user *models.User) error {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO users (
			tenant_id, email, password_hash, first_name, last_name,
			phone, status, timezone, language, preferences, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at
	`

	err = tx.QueryRowContext(
		ctx, query,
		tenantID,
		user.Email,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.Phone,
		user.Status,
		user.Timezone,
		user.Language,
		user.Preferences,
		user.CreatedBy,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	user.TenantID = tenantID
	return tx.Commit()
}

// FindByID retrieves a user by ID with RLS
func (r *UserRepository) FindByID(ctx context.Context, tenantID, userID uuid.UUID) (*models.User, error) {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var user models.User
	query := `SELECT * FROM users WHERE id = $1 LIMIT 1`

	err = tx.GetContext(ctx, &user, query, userID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return &user, nil
}

// FindByEmail retrieves a user by email with RLS
func (r *UserRepository) FindByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*models.User, error) {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var user models.User
	query := `SELECT * FROM users WHERE email = $1 LIMIT 1`

	err = tx.GetContext(ctx, &user, query, email)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return &user, nil
}

// FindByResetToken retrieves a user by password reset token
func (r *UserRepository) FindByResetToken(ctx context.Context, tenantID uuid.UUID, token uuid.UUID) (*models.User, error) {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var user models.User
	query := `
		SELECT * FROM users
		WHERE reset_token = $1
		  AND reset_token_expires_at > NOW()
		LIMIT 1
	`

	err = tx.GetContext(ctx, &user, query, token)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("invalid or expired reset token")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return &user, nil
}

// Update updates a user's information
func (r *UserRepository) Update(ctx context.Context, tenantID uuid.UUID, user *models.User) error {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		UPDATE users
		SET first_name = $1,
		    last_name = $2,
		    phone = $3,
		    avatar_url = $4,
		    timezone = $5,
		    language = $6,
		    preferences = $7,
		    updated_at = NOW()
		WHERE id = $8
		RETURNING updated_at
	`

	err = tx.QueryRowContext(
		ctx, query,
		user.FirstName,
		user.LastName,
		user.Phone,
		user.AvatarURL,
		user.Timezone,
		user.Language,
		user.Preferences,
		user.ID,
	).Scan(&user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return tx.Commit()
}

// UpdatePassword updates a user's password
func (r *UserRepository) UpdatePassword(ctx context.Context, tenantID, userID uuid.UUID, passwordHash string) error {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		UPDATE users
		SET password_hash = $1,
		    reset_token = NULL,
		    reset_token_expires_at = NULL,
		    updated_at = NOW()
		WHERE id = $2
	`

	result, err := tx.ExecContext(ctx, query, passwordHash, userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return tx.Commit()
}

// UpdateStatus updates a user's status
func (r *UserRepository) UpdateStatus(ctx context.Context, tenantID, userID uuid.UUID, status string) error {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		UPDATE users
		SET status = $1,
		    updated_at = NOW()
		WHERE id = $2
	`

	result, err := tx.ExecContext(ctx, query, status, userID)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return tx.Commit()
}

// UpdateLastLogin updates the user's last login information
func (r *UserRepository) UpdateLastLogin(ctx context.Context, tenantID, userID uuid.UUID, ipAddress string) error {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		UPDATE users
		SET last_login_at = NOW(),
		    last_login_ip = $1,
		    last_active_at = NOW()
		WHERE id = $2
	`

	_, err = tx.ExecContext(ctx, query, ipAddress, userID)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	return tx.Commit()
}

// VerifyEmail marks a user's email as verified
func (r *UserRepository) VerifyEmail(ctx context.Context, tenantID, userID uuid.UUID) error {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		UPDATE users
		SET email_verified = true,
		    updated_at = NOW()
		WHERE id = $1
	`

	result, err := tx.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to verify email: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return tx.Commit()
}

// SetResetToken sets a password reset token for a user
func (r *UserRepository) SetResetToken(ctx context.Context, tenantID, userID, token uuid.UUID, expiresAt string) error {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		UPDATE users
		SET reset_token = $1,
		    reset_token_expires_at = $2,
		    updated_at = NOW()
		WHERE id = $3
	`

	result, err := tx.ExecContext(ctx, query, token, expiresAt, userID)
	if err != nil {
		return fmt.Errorf("failed to set reset token: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return tx.Commit()
}

// Enable2FA enables two-factor authentication for a user
func (r *UserRepository) Enable2FA(ctx context.Context, tenantID, userID uuid.UUID, secret string, backupCodes []string) error {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		UPDATE users
		SET two_factor_enabled = true,
		    two_factor_secret = $1,
		    two_factor_backup_codes = $2,
		    two_factor_enabled_at = NOW(),
		    updated_at = NOW()
		WHERE id = $3
	`

	_, err = tx.ExecContext(ctx, query, secret, backupCodes, userID)
	if err != nil {
		return fmt.Errorf("failed to enable 2FA: %w", err)
	}

	return tx.Commit()
}

// Disable2FA disables two-factor authentication for a user
func (r *UserRepository) Disable2FA(ctx context.Context, tenantID, userID uuid.UUID) error {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		UPDATE users
		SET two_factor_enabled = false,
		    two_factor_secret = NULL,
		    two_factor_backup_codes = NULL,
		    two_factor_enabled_at = NULL,
		    updated_at = NOW()
		WHERE id = $1
	`

	_, err = tx.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to disable 2FA: %w", err)
	}

	return tx.Commit()
}

// UseBackupCode uses a backup code and removes it from the list
func (r *UserRepository) UseBackupCode(ctx context.Context, tenantID, userID uuid.UUID, remainingCodes []string) error {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		UPDATE users
		SET two_factor_backup_codes = $1,
		    updated_at = NOW()
		WHERE id = $2
	`

	_, err = tx.ExecContext(ctx, query, remainingCodes, userID)
	if err != nil {
		return fmt.Errorf("failed to use backup code: %w", err)
	}

	return tx.Commit()
}

// List retrieves a paginated list of users with RLS
func (r *UserRepository) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]models.User, int, error) {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback()

	var users []models.User
	var totalCount int

	// Get total count
	countQuery := `SELECT COUNT(*) FROM users`
	err = tx.GetContext(ctx, &totalCount, countQuery)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Get paginated results
	query := `
		SELECT * FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	err = tx.SelectContext(ctx, &users, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	return users, totalCount, nil
}

// Search searches for users by name or email
func (r *UserRepository) Search(ctx context.Context, tenantID uuid.UUID, searchTerm string, limit int) ([]models.User, error) {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var users []models.User
	query := `
		SELECT * FROM users
		WHERE email ILIKE $1
		   OR first_name ILIKE $1
		   OR last_name ILIKE $1
		   OR (first_name || ' ' || last_name) ILIKE $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	searchPattern := "%" + searchTerm + "%"
	err = tx.SelectContext(ctx, &users, query, searchPattern, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	return users, nil
}

// Delete deletes a user (hard delete)
func (r *UserRepository) Delete(ctx context.Context, tenantID, userID uuid.UUID) error {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `DELETE FROM users WHERE id = $1`

	result, err := tx.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return tx.Commit()
}

// CheckEmailExists checks if an email is already registered in the tenant
func (r *UserRepository) CheckEmailExists(ctx context.Context, tenantID uuid.UUID, email string) (bool, error) {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	var count int
	query := `SELECT COUNT(*) FROM users WHERE email = $1`

	err = tx.GetContext(ctx, &count, query, email)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return count > 0, nil
}

// CountByStatus counts users by status
func (r *UserRepository) CountByStatus(ctx context.Context, tenantID uuid.UUID, status string) (int, error) {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var count int
	query := `SELECT COUNT(*) FROM users WHERE status = $1`

	err = tx.GetContext(ctx, &count, query, status)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}
