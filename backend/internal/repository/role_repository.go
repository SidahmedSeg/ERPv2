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

// RoleRepository handles database operations for roles
type RoleRepository struct {
	db *sqlx.DB
}

// NewRoleRepository creates a new role repository
func NewRoleRepository(db *sqlx.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

// Create creates a new role with RLS
func (r *RoleRepository) Create(ctx context.Context, tenantID uuid.UUID, role *models.Role) error {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO roles (
			tenant_id, name, display_name, description, parent_role_id,
			level, is_system, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`

	err = tx.QueryRowContext(
		ctx, query,
		tenantID,
		role.Name,
		role.DisplayName,
		role.Description,
		role.ParentRoleID,
		role.Level,
		role.IsSystem,
		role.CreatedBy,
	).Scan(&role.ID, &role.CreatedAt, &role.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}

	role.TenantID = tenantID
	return tx.Commit()
}

// FindByID retrieves a role by ID with RLS
func (r *RoleRepository) FindByID(ctx context.Context, tenantID, roleID uuid.UUID) (*models.Role, error) {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var role models.Role
	query := `SELECT * FROM roles WHERE id = $1 LIMIT 1`

	err = tx.GetContext(ctx, &role, query, roleID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("role not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find role: %w", err)
	}

	return &role, nil
}

// FindByName retrieves a role by name with RLS
func (r *RoleRepository) FindByName(ctx context.Context, tenantID uuid.UUID, name string) (*models.Role, error) {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var role models.Role
	query := `SELECT * FROM roles WHERE name = $1 LIMIT 1`

	err = tx.GetContext(ctx, &role, query, name)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("role not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find role: %w", err)
	}

	return &role, nil
}

// List retrieves all roles for a tenant
func (r *RoleRepository) List(ctx context.Context, tenantID uuid.UUID) ([]models.Role, error) {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var roles []models.Role
	query := `
		SELECT * FROM roles
		ORDER BY level ASC, name ASC
	`

	err = tx.SelectContext(ctx, &roles, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	return roles, nil
}

// ListSystemRoles retrieves all system roles for a tenant
func (r *RoleRepository) ListSystemRoles(ctx context.Context, tenantID uuid.UUID) ([]models.Role, error) {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var roles []models.Role
	query := `
		SELECT * FROM roles
		WHERE is_system = true
		ORDER BY level ASC
	`

	err = tx.SelectContext(ctx, &roles, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list system roles: %w", err)
	}

	return roles, nil
}

// ListCustomRoles retrieves all custom (non-system) roles for a tenant
func (r *RoleRepository) ListCustomRoles(ctx context.Context, tenantID uuid.UUID) ([]models.Role, error) {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var roles []models.Role
	query := `
		SELECT * FROM roles
		WHERE is_system = false
		ORDER BY name ASC
	`

	err = tx.SelectContext(ctx, &roles, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list custom roles: %w", err)
	}

	return roles, nil
}

// Update updates a role's information
func (r *RoleRepository) Update(ctx context.Context, tenantID uuid.UUID, role *models.Role) error {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		UPDATE roles
		SET display_name = $1,
		    description = $2,
		    parent_role_id = $3,
		    updated_at = NOW()
		WHERE id = $4 AND is_system = false
		RETURNING updated_at
	`

	err = tx.QueryRowContext(
		ctx, query,
		role.DisplayName,
		role.Description,
		role.ParentRoleID,
		role.ID,
	).Scan(&role.UpdatedAt)

	if err == sql.ErrNoRows {
		return fmt.Errorf("role not found or is a system role")
	}
	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	return tx.Commit()
}

// Delete deletes a role (only custom roles)
func (r *RoleRepository) Delete(ctx context.Context, tenantID, roleID uuid.UUID) error {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `DELETE FROM roles WHERE id = $1 AND is_system = false`

	result, err := tx.ExecContext(ctx, query, roleID)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("role not found or is a system role")
	}

	return tx.Commit()
}

// AssignPermissions assigns permissions to a role
func (r *RoleRepository) AssignPermissions(ctx context.Context, tenantID, roleID uuid.UUID, permissionIDs []uuid.UUID, assignedBy uuid.UUID) error {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete existing permissions
	deleteQuery := `DELETE FROM role_permissions WHERE role_id = $1`
	_, err = tx.ExecContext(ctx, deleteQuery, roleID)
	if err != nil {
		return fmt.Errorf("failed to delete existing permissions: %w", err)
	}

	// Insert new permissions
	insertQuery := `
		INSERT INTO role_permissions (tenant_id, role_id, permission_id, created_by)
		VALUES ($1, $2, $3, $4)
	`

	for _, permID := range permissionIDs {
		_, err = tx.ExecContext(ctx, insertQuery, tenantID, roleID, permID, assignedBy)
		if err != nil {
			return fmt.Errorf("failed to assign permission: %w", err)
		}
	}

	return tx.Commit()
}

// GetPermissions retrieves all permissions for a role
func (r *RoleRepository) GetPermissions(ctx context.Context, tenantID, roleID uuid.UUID) ([]models.Permission, error) {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var permissions []models.Permission
	query := `
		SELECT p.*
		FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
		ORDER BY p.category, p.resource, p.action
	`

	err = tx.SelectContext(ctx, &permissions, query, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}

	return permissions, nil
}

// CountUsers counts the number of users assigned to a role
func (r *RoleRepository) CountUsers(ctx context.Context, tenantID, roleID uuid.UUID) (int, error) {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var count int
	query := `SELECT COUNT(DISTINCT user_id) FROM user_roles WHERE role_id = $1`

	err = tx.GetContext(ctx, &count, query, roleID)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}

// CheckNameExists checks if a role name already exists for a tenant
func (r *RoleRepository) CheckNameExists(ctx context.Context, tenantID uuid.UUID, name string, excludeID *uuid.UUID) (bool, error) {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	var count int
	var query string

	if excludeID != nil {
		query = `SELECT COUNT(*) FROM roles WHERE name = $1 AND id != $2`
		err = tx.GetContext(ctx, &count, query, name, *excludeID)
	} else {
		query = `SELECT COUNT(*) FROM roles WHERE name = $1`
		err = tx.GetContext(ctx, &count, query, name)
	}

	if err != nil {
		return false, fmt.Errorf("failed to check role name: %w", err)
	}

	return count > 0, nil
}

// GetRoleWithDetails retrieves a role with permission and user counts
func (r *RoleRepository) GetRoleWithDetails(ctx context.Context, tenantID, roleID uuid.UUID) (*models.Role, error) {
	role, err := r.FindByID(ctx, tenantID, roleID)
	if err != nil {
		return nil, err
	}

	// Get permissions
	permissions, err := r.GetPermissions(ctx, tenantID, roleID)
	if err != nil {
		return nil, err
	}
	role.Permissions = permissions
	role.PermissionCount = len(permissions)

	// Get user count
	userCount, err := r.CountUsers(ctx, tenantID, roleID)
	if err != nil {
		return nil, err
	}
	role.UserCount = userCount

	return role, nil
}

// ListWithDetails retrieves all roles with permission and user counts
func (r *RoleRepository) ListWithDetails(ctx context.Context, tenantID uuid.UUID) ([]models.Role, error) {
	roles, err := r.List(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// Enrich each role with details
	for i := range roles {
		// Get permissions
		permissions, err := r.GetPermissions(ctx, tenantID, roles[i].ID)
		if err != nil {
			continue // Skip on error
		}
		roles[i].Permissions = permissions
		roles[i].PermissionCount = len(permissions)

		// Get user count
		userCount, err := r.CountUsers(ctx, tenantID, roles[i].ID)
		if err != nil {
			continue // Skip on error
		}
		roles[i].UserCount = userCount
	}

	return roles, nil
}
