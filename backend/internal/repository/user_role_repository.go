package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"myerp-v2/internal/database"
	"myerp-v2/internal/models"
)

// UserRoleRepository handles database operations for user-role assignments
type UserRoleRepository struct {
	db *sqlx.DB
}

// NewUserRoleRepository creates a new user-role repository
func NewUserRoleRepository(db *sqlx.DB) *UserRoleRepository {
	return &UserRoleRepository{db: db}
}

// AssignRole assigns a role to a user
func (r *UserRoleRepository) AssignRole(ctx context.Context, tenantID, userID, roleID, assignedBy uuid.UUID) error {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO user_roles (tenant_id, user_id, role_id, assigned_by)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (tenant_id, user_id, role_id) DO NOTHING
	`

	_, err = tx.ExecContext(ctx, query, tenantID, userID, roleID, assignedBy)
	if err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	return tx.Commit()
}

// AssignRoles assigns multiple roles to a user (replaces existing roles)
func (r *UserRoleRepository) AssignRoles(ctx context.Context, tenantID, userID uuid.UUID, roleIDs []uuid.UUID, assignedBy uuid.UUID) error {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete existing roles
	deleteQuery := `DELETE FROM user_roles WHERE user_id = $1`
	_, err = tx.ExecContext(ctx, deleteQuery, userID)
	if err != nil {
		return fmt.Errorf("failed to delete existing roles: %w", err)
	}

	// Insert new roles
	insertQuery := `
		INSERT INTO user_roles (tenant_id, user_id, role_id, assigned_by)
		VALUES ($1, $2, $3, $4)
	`

	for _, roleID := range roleIDs {
		_, err = tx.ExecContext(ctx, insertQuery, tenantID, userID, roleID, assignedBy)
		if err != nil {
			return fmt.Errorf("failed to assign role: %w", err)
		}
	}

	return tx.Commit()
}

// UnassignRole removes a role from a user
func (r *UserRoleRepository) UnassignRole(ctx context.Context, tenantID, userID, roleID uuid.UUID) error {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2`

	result, err := tx.ExecContext(ctx, query, userID, roleID)
	if err != nil {
		return fmt.Errorf("failed to unassign role: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("role assignment not found")
	}

	return tx.Commit()
}

// UnassignAllRoles removes all roles from a user
func (r *UserRoleRepository) UnassignAllRoles(ctx context.Context, tenantID, userID uuid.UUID) error {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `DELETE FROM user_roles WHERE user_id = $1`

	_, err = tx.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to unassign all roles: %w", err)
	}

	return tx.Commit()
}

// GetUserRoles retrieves all roles assigned to a user
func (r *UserRoleRepository) GetUserRoles(ctx context.Context, tenantID, userID uuid.UUID) ([]models.Role, error) {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var roles []models.Role
	query := `
		SELECT r.*
		FROM roles r
		INNER JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
		ORDER BY r.level ASC, r.name ASC
	`

	err = tx.SelectContext(ctx, &roles, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	return roles, nil
}

// GetUsersByRole retrieves all users assigned to a role
func (r *UserRoleRepository) GetUsersByRole(ctx context.Context, tenantID, roleID uuid.UUID) ([]models.User, error) {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var users []models.User
	query := `
		SELECT u.*
		FROM users u
		INNER JOIN user_roles ur ON u.id = ur.user_id
		WHERE ur.role_id = $1
		ORDER BY u.first_name, u.last_name
	`

	err = tx.SelectContext(ctx, &users, query, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get users by role: %w", err)
	}

	return users, nil
}

// HasRole checks if a user has a specific role
func (r *UserRoleRepository) HasRole(ctx context.Context, tenantID, userID, roleID uuid.UUID) (bool, error) {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	var count int
	query := `
		SELECT COUNT(*)
		FROM user_roles
		WHERE user_id = $1 AND role_id = $2
	`

	err = tx.GetContext(ctx, &count, query, userID, roleID)
	if err != nil {
		return false, fmt.Errorf("failed to check role: %w", err)
	}

	return count > 0, nil
}

// HasAnyRole checks if a user has any of the specified roles
func (r *UserRoleRepository) HasAnyRole(ctx context.Context, tenantID, userID uuid.UUID, roleIDs []uuid.UUID) (bool, error) {
	if len(roleIDs) == 0 {
		return false, nil
	}

	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	var count int
	query := `
		SELECT COUNT(*)
		FROM user_roles
		WHERE user_id = $1 AND role_id = ANY($2)
	`

	err = tx.GetContext(ctx, &count, query, userID, roleIDs)
	if err != nil {
		return false, fmt.Errorf("failed to check roles: %w", err)
	}

	return count > 0, nil
}

// CountUsersByRole counts the number of users assigned to a role
func (r *UserRoleRepository) CountUsersByRole(ctx context.Context, tenantID, roleID uuid.UUID) (int, error) {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var count int
	query := `
		SELECT COUNT(DISTINCT user_id)
		FROM user_roles
		WHERE role_id = $1
	`

	err = tx.GetContext(ctx, &count, query, roleID)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}

// GetRoleAssignmentDetails retrieves detailed role assignment information
func (r *UserRoleRepository) GetRoleAssignmentDetails(ctx context.Context, tenantID, userID uuid.UUID) ([]map[string]interface{}, error) {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	type AssignmentDetail struct {
		RoleID       uuid.UUID `db:"role_id"`
		RoleName     string    `db:"role_name"`
		DisplayName  string    `db:"display_name"`
		IsSystem     bool      `db:"is_system"`
		AssignedAt   string    `db:"assigned_at"`
		AssignedBy   uuid.UUID `db:"assigned_by"`
		AssignerName string    `db:"assigner_name"`
	}

	var details []AssignmentDetail
	query := `
		SELECT
			r.id as role_id,
			r.name as role_name,
			r.display_name,
			r.is_system,
			ur.assigned_at,
			ur.assigned_by,
			COALESCE(u.first_name || ' ' || u.last_name, 'System') as assigner_name
		FROM user_roles ur
		INNER JOIN roles r ON ur.role_id = r.id
		LEFT JOIN users u ON ur.assigned_by = u.id
		WHERE ur.user_id = $1
		ORDER BY r.level ASC
	`

	err = tx.SelectContext(ctx, &details, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get assignment details: %w", err)
	}

	// Convert to map slice for easier JSON marshaling
	result := make([]map[string]interface{}, len(details))
	for i, detail := range details {
		result[i] = map[string]interface{}{
			"role_id":      detail.RoleID,
			"role_name":    detail.RoleName,
			"display_name": detail.DisplayName,
			"is_system":    detail.IsSystem,
			"assigned_at":  detail.AssignedAt,
			"assigned_by":  detail.AssignedBy,
			"assigner_name": detail.AssignerName,
		}
	}

	return result, nil
}

// BulkAssignRole assigns a role to multiple users
func (r *UserRoleRepository) BulkAssignRole(ctx context.Context, tenantID uuid.UUID, userIDs []uuid.UUID, roleID, assignedBy uuid.UUID) error {
	if len(userIDs) == 0 {
		return nil
	}

	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO user_roles (tenant_id, user_id, role_id, assigned_by)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (tenant_id, user_id, role_id) DO NOTHING
	`

	for _, userID := range userIDs {
		_, err = tx.ExecContext(ctx, query, tenantID, userID, roleID, assignedBy)
		if err != nil {
			return fmt.Errorf("failed to assign role to user %s: %w", userID, err)
		}
	}

	return tx.Commit()
}

// BulkUnassignRole removes a role from multiple users
func (r *UserRoleRepository) BulkUnassignRole(ctx context.Context, tenantID uuid.UUID, userIDs []uuid.UUID, roleID uuid.UUID) error {
	if len(userIDs) == 0 {
		return nil
	}

	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		DELETE FROM user_roles
		WHERE role_id = $1 AND user_id = ANY($2)
	`

	_, err = tx.ExecContext(ctx, query, roleID, userIDs)
	if err != nil {
		return fmt.Errorf("failed to bulk unassign role: %w", err)
	}

	return tx.Commit()
}
