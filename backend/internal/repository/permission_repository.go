package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"myerp-v2/internal/models"
)

// PermissionRepository handles database operations for permissions
// Note: Permissions table does NOT have RLS (global catalog)
type PermissionRepository struct {
	db *sqlx.DB
}

// NewPermissionRepository creates a new permission repository
func NewPermissionRepository(db *sqlx.DB) *PermissionRepository {
	return &PermissionRepository{db: db}
}

// FindByID retrieves a permission by ID
func (r *PermissionRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Permission, error) {
	var permission models.Permission
	query := `SELECT * FROM permissions WHERE id = $1 LIMIT 1`

	err := r.db.GetContext(ctx, &permission, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("permission not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find permission: %w", err)
	}

	return &permission, nil
}

// FindByResourceAction retrieves a permission by resource and action
func (r *PermissionRepository) FindByResourceAction(ctx context.Context, resource, action string) (*models.Permission, error) {
	var permission models.Permission
	query := `SELECT * FROM permissions WHERE resource = $1 AND action = $2 LIMIT 1`

	err := r.db.GetContext(ctx, &permission, query, resource, action)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("permission not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find permission: %w", err)
	}

	return &permission, nil
}

// List retrieves all permissions
func (r *PermissionRepository) List(ctx context.Context) ([]models.Permission, error) {
	var permissions []models.Permission
	query := `
		SELECT * FROM permissions
		ORDER BY category, resource, action
	`

	err := r.db.SelectContext(ctx, &permissions, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}

	return permissions, nil
}

// ListByCategory retrieves all permissions grouped by category
func (r *PermissionRepository) ListByCategory(ctx context.Context) ([]models.PermissionGroup, error) {
	permissions, err := r.List(ctx)
	if err != nil {
		return nil, err
	}

	// Group permissions by category
	categoryMap := make(map[string][]models.Permission)
	for _, perm := range permissions {
		category := "Other"
		if perm.Category != nil {
			category = *perm.Category
		}
		categoryMap[category] = append(categoryMap[category], perm)
	}

	// Convert map to slice
	var groups []models.PermissionGroup
	for category, perms := range categoryMap {
		groups = append(groups, models.PermissionGroup{
			Category:    category,
			Permissions: perms,
		})
	}

	return groups, nil
}

// ListByResource retrieves all permissions for a specific resource
func (r *PermissionRepository) ListByResource(ctx context.Context, resource string) ([]models.Permission, error) {
	var permissions []models.Permission
	query := `
		SELECT * FROM permissions
		WHERE resource = $1
		ORDER BY action
	`

	err := r.db.SelectContext(ctx, &permissions, query, resource)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions by resource: %w", err)
	}

	return permissions, nil
}

// ListByIDs retrieves multiple permissions by their IDs
func (r *PermissionRepository) ListByIDs(ctx context.Context, ids []uuid.UUID) ([]models.Permission, error) {
	if len(ids) == 0 {
		return []models.Permission{}, nil
	}

	var permissions []models.Permission
	query := `
		SELECT * FROM permissions
		WHERE id = ANY($1)
		ORDER BY category, resource, action
	`

	err := r.db.SelectContext(ctx, &permissions, query, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions by IDs: %w", err)
	}

	return permissions, nil
}

// Search searches permissions by name or description
func (r *PermissionRepository) Search(ctx context.Context, searchTerm string) ([]models.Permission, error) {
	var permissions []models.Permission
	query := `
		SELECT * FROM permissions
		WHERE display_name ILIKE $1
		   OR description ILIKE $1
		   OR resource ILIKE $1
		ORDER BY category, resource, action
		LIMIT 50
	`

	searchPattern := "%" + searchTerm + "%"
	err := r.db.SelectContext(ctx, &permissions, query, searchPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to search permissions: %w", err)
	}

	return permissions, nil
}

// Count returns the total number of permissions
func (r *PermissionRepository) Count(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM permissions`

	err := r.db.GetContext(ctx, &count, query)
	if err != nil {
		return 0, fmt.Errorf("failed to count permissions: %w", err)
	}

	return count, nil
}

// CountByCategory returns the count of permissions per category
func (r *PermissionRepository) CountByCategory(ctx context.Context) (map[string]int, error) {
	type CategoryCount struct {
		Category string `db:"category"`
		Count    int    `db:"count"`
	}

	var results []CategoryCount
	query := `
		SELECT COALESCE(category, 'Other') as category, COUNT(*) as count
		FROM permissions
		GROUP BY category
		ORDER BY category
	`

	err := r.db.SelectContext(ctx, &results, query)
	if err != nil {
		return nil, fmt.Errorf("failed to count by category: %w", err)
	}

	counts := make(map[string]int)
	for _, result := range results {
		counts[result.Category] = result.Count
	}

	return counts, nil
}

// ValidatePermissionIDs validates that all given permission IDs exist
func (r *PermissionRepository) ValidatePermissionIDs(ctx context.Context, ids []uuid.UUID) (bool, []uuid.UUID, error) {
	if len(ids) == 0 {
		return true, nil, nil
	}

	var existingIDs []uuid.UUID
	query := `SELECT id FROM permissions WHERE id = ANY($1)`

	err := r.db.SelectContext(ctx, &existingIDs, query, ids)
	if err != nil {
		return false, nil, fmt.Errorf("failed to validate permission IDs: %w", err)
	}

	// Check if all IDs exist
	if len(existingIDs) != len(ids) {
		// Find missing IDs
		existingMap := make(map[uuid.UUID]bool)
		for _, id := range existingIDs {
			existingMap[id] = true
		}

		var missingIDs []uuid.UUID
		for _, id := range ids {
			if !existingMap[id] {
				missingIDs = append(missingIDs, id)
			}
		}

		return false, missingIDs, nil
	}

	return true, nil, nil
}

// GetResourceList returns a list of all unique resources
func (r *PermissionRepository) GetResourceList(ctx context.Context) ([]string, error) {
	var resources []string
	query := `
		SELECT DISTINCT resource
		FROM permissions
		ORDER BY resource
	`

	err := r.db.SelectContext(ctx, &resources, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource list: %w", err)
	}

	return resources, nil
}

// GetActionsByResource returns all actions for a specific resource
func (r *PermissionRepository) GetActionsByResource(ctx context.Context, resource string) ([]string, error) {
	var actions []string
	query := `
		SELECT action
		FROM permissions
		WHERE resource = $1
		ORDER BY action
	`

	err := r.db.SelectContext(ctx, &actions, query, resource)
	if err != nil {
		return nil, fmt.Errorf("failed to get actions by resource: %w", err)
	}

	return actions, nil
}

// GetCategoryList returns a list of all unique categories
func (r *PermissionRepository) GetCategoryList(ctx context.Context) ([]string, error) {
	var categories []string
	query := `
		SELECT DISTINCT COALESCE(category, 'Other') as category
		FROM permissions
		ORDER BY category
	`

	err := r.db.SelectContext(ctx, &categories, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get category list: %w", err)
	}

	return categories, nil
}
