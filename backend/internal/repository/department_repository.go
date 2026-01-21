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

// DepartmentRepository handles database operations for departments
type DepartmentRepository struct {
	db *sqlx.DB
}

// NewDepartmentRepository creates a new department repository
func NewDepartmentRepository(db *sqlx.DB) *DepartmentRepository {
	return &DepartmentRepository{db: db}
}

// Create creates a new department with RLS
func (r *DepartmentRepository) Create(ctx context.Context, tenantID uuid.UUID, dept *models.Department) error {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO departments (
			tenant_id, name, description, head_user_id, color, icon, status, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`

	err = tx.QueryRowContext(
		ctx, query,
		tenantID,
		dept.Name,
		dept.Description,
		dept.HeadUserID,
		dept.Color,
		dept.Icon,
		dept.Status,
		dept.CreatedBy,
	).Scan(&dept.ID, &dept.CreatedAt, &dept.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create department: %w", err)
	}

	dept.TenantID = tenantID
	return tx.Commit()
}

// FindByID retrieves a department by ID with RLS
func (r *DepartmentRepository) FindByID(ctx context.Context, tenantID, deptID uuid.UUID) (*models.Department, error) {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var dept models.Department
	// Explicit tenant_id filter for defense in depth
	query := `SELECT * FROM departments WHERE tenant_id = $1 AND id = $2 LIMIT 1`

	err = tx.GetContext(ctx, &dept, query, tenantID, deptID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("department not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find department: %w", err)
	}

	return &dept, nil
}

// FindByName retrieves a department by name with RLS
func (r *DepartmentRepository) FindByName(ctx context.Context, tenantID uuid.UUID, name string) (*models.Department, error) {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var dept models.Department
	query := `SELECT * FROM departments WHERE tenant_id = $1 AND name = $2 LIMIT 1`

	err = tx.GetContext(ctx, &dept, query, tenantID, name)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("department not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find department: %w", err)
	}

	return &dept, nil
}

// List retrieves all departments for a tenant with enriched data
func (r *DepartmentRepository) List(ctx context.Context, tenantID uuid.UUID) ([]models.Department, error) {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var departments []models.Department
	query := `
		SELECT
			d.*,
			u.first_name || ' ' || u.last_name AS head_user_name,
			COALESCE((SELECT COUNT(*) FROM users WHERE department_id = d.id AND tenant_id = d.tenant_id), 0) AS member_count
		FROM departments d
		LEFT JOIN users u ON d.head_user_id = u.id AND d.tenant_id = u.tenant_id
		ORDER BY d.created_at DESC
	`

	err = tx.SelectContext(ctx, &departments, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list departments: %w", err)
	}

	return departments, nil
}

// ListActive retrieves only active departments
func (r *DepartmentRepository) ListActive(ctx context.Context, tenantID uuid.UUID) ([]models.Department, error) {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var departments []models.Department
	query := `
		SELECT d.* FROM departments d
		WHERE d.status = 'active'
		ORDER BY d.name ASC
	`

	err = tx.SelectContext(ctx, &departments, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list active departments: %w", err)
	}

	return departments, nil
}

// GetWithDetails retrieves a department with enriched data
func (r *DepartmentRepository) GetWithDetails(ctx context.Context, tenantID, deptID uuid.UUID) (*models.Department, error) {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var dept models.Department
	query := `
		SELECT
			d.*,
			u.first_name || ' ' || u.last_name AS head_user_name,
			COALESCE((SELECT COUNT(*) FROM users WHERE department_id = d.id AND tenant_id = d.tenant_id), 0) AS member_count
		FROM departments d
		LEFT JOIN users u ON d.head_user_id = u.id AND d.tenant_id = u.tenant_id
		WHERE d.tenant_id = $1 AND d.id = $2
		LIMIT 1
	`

	err = tx.GetContext(ctx, &dept, query, tenantID, deptID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("department not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get department: %w", err)
	}

	return &dept, nil
}

// Update updates a department's information
func (r *DepartmentRepository) Update(ctx context.Context, tenantID uuid.UUID, dept *models.Department) error {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		UPDATE departments
		SET name = $1,
			description = $2,
			head_user_id = $3,
			color = $4,
			icon = $5,
			status = $6,
			updated_at = NOW()
		WHERE id = $7
		RETURNING updated_at
	`

	err = tx.QueryRowContext(
		ctx, query,
		dept.Name,
		dept.Description,
		dept.HeadUserID,
		dept.Color,
		dept.Icon,
		dept.Status,
		dept.ID,
	).Scan(&dept.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update department: %w", err)
	}

	return tx.Commit()
}

// Delete deletes a department (hard delete)
func (r *DepartmentRepository) Delete(ctx context.Context, tenantID, deptID uuid.UUID) error {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `DELETE FROM departments WHERE id = $1`

	result, err := tx.ExecContext(ctx, query, deptID)
	if err != nil {
		return fmt.Errorf("failed to delete department: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("department not found")
	}

	return tx.Commit()
}

// GetMembers retrieves all users in a department
func (r *DepartmentRepository) GetMembers(ctx context.Context, tenantID, deptID uuid.UUID) ([]models.User, error) {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var users []models.User
	query := `
		SELECT u.* FROM users u
		WHERE u.department_id = $1
		ORDER BY u.first_name, u.last_name
	`

	err = tx.SelectContext(ctx, &users, query, deptID)
	if err != nil {
		return nil, fmt.Errorf("failed to get department members: %w", err)
	}

	return users, nil
}

// CheckNameExists checks if a department name already exists for a tenant
func (r *DepartmentRepository) CheckNameExists(ctx context.Context, tenantID uuid.UUID, name string, excludeID *uuid.UUID) (bool, error) {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	var count int
	var query string

	if excludeID != nil {
		query = `SELECT COUNT(*) FROM departments WHERE name = $1 AND id != $2`
		err = tx.GetContext(ctx, &count, query, name, *excludeID)
	} else {
		query = `SELECT COUNT(*) FROM departments WHERE name = $1`
		err = tx.GetContext(ctx, &count, query, name)
	}

	if err != nil {
		return false, fmt.Errorf("failed to check department name: %w", err)
	}

	return count > 0, nil
}

// CountMembers counts members in a department
func (r *DepartmentRepository) CountMembers(ctx context.Context, tenantID, deptID uuid.UUID) (int, error) {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var count int
	query := `SELECT COUNT(*) FROM users WHERE department_id = $1`

	err = tx.GetContext(ctx, &count, query, deptID)
	if err != nil {
		return 0, fmt.Errorf("failed to count members: %w", err)
	}

	return count, nil
}
