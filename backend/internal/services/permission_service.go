package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"myerp-v2/internal/database"
	"myerp-v2/internal/models"
	"myerp-v2/internal/repository"
)

// PermissionService handles permission-related business logic
type PermissionService struct {
	permissionRepo *repository.PermissionRepository
	userRoleRepo   *repository.UserRoleRepository
	roleRepo       *repository.RoleRepository
	redis          *redis.Client
}

// NewPermissionService creates a new permission service
func NewPermissionService(
	permissionRepo *repository.PermissionRepository,
	userRoleRepo *repository.UserRoleRepository,
	roleRepo *repository.RoleRepository,
	redisClient *redis.Client,
) *PermissionService {
	return &PermissionService{
		permissionRepo: permissionRepo,
		userRoleRepo:   userRoleRepo,
		roleRepo:       roleRepo,
		redis:          redisClient,
	}
}

const (
	permissionCacheTTL      = 15 * time.Minute
	userPermissionKeyPrefix = "user_perms:"
	rolePermissionKeyPrefix = "role_perms:"
)

// GetUserPermissions retrieves all permissions for a user (with caching)
func (s *PermissionService) GetUserPermissions(ctx context.Context, tenantID, userID uuid.UUID) ([]models.Permission, error) {
	// Try cache first
	cacheKey := database.CacheKey(userPermissionKeyPrefix, tenantID.String(), userID.String())
	cachedData, err := s.redis.Get(ctx, cacheKey).Result()

	if err == nil {
		// Cache hit
		var permissions []models.Permission
		if err := json.Unmarshal([]byte(cachedData), &permissions); err == nil {
			return permissions, nil
		}
	}

	// Cache miss - query database
	permissions, err := s.getUserPermissionsFromDB(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}

	// Cache the result
	data, _ := json.Marshal(permissions)
	s.redis.Set(ctx, cacheKey, data, permissionCacheTTL)

	return permissions, nil
}

// getUserPermissionsFromDB retrieves user permissions from database
func (s *PermissionService) getUserPermissionsFromDB(ctx context.Context, tenantID, userID uuid.UUID) ([]models.Permission, error) {
	// Get user's roles
	roles, err := s.userRoleRepo.GetUserRoles(ctx, tenantID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	if len(roles) == 0 {
		return []models.Permission{}, nil
	}

	// Collect all permissions from all roles
	permissionMap := make(map[uuid.UUID]models.Permission)

	for _, role := range roles {
		rolePerms, err := s.roleRepo.GetPermissions(ctx, tenantID, role.ID)
		if err != nil {
			continue // Skip on error
		}

		for _, perm := range rolePerms {
			permissionMap[perm.ID] = perm
		}
	}

	// Convert map to slice
	permissions := make([]models.Permission, 0, len(permissionMap))
	for _, perm := range permissionMap {
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// HasPermission checks if a user has a specific permission (with caching)
func (s *PermissionService) HasPermission(ctx context.Context, tenantID, userID uuid.UUID, resource, action string) (bool, error) {
	// Get user permissions (from cache or DB)
	permissions, err := s.GetUserPermissions(ctx, tenantID, userID)
	if err != nil {
		return false, err
	}

	// Check if user has the required permission
	for _, perm := range permissions {
		if perm.Matches(resource, action) {
			return true, nil
		}
	}

	return false, nil
}

// HasAnyPermission checks if a user has any of the specified permissions
func (s *PermissionService) HasAnyPermission(ctx context.Context, tenantID, userID uuid.UUID, checks []PermissionCheck) (bool, error) {
	// Get user permissions (from cache or DB)
	permissions, err := s.GetUserPermissions(ctx, tenantID, userID)
	if err != nil {
		return false, err
	}

	// Check if user has any of the required permissions
	for _, check := range checks {
		for _, perm := range permissions {
			if perm.Matches(check.Resource, check.Action) {
				return true, nil
			}
		}
	}

	return false, nil
}

// HasAllPermissions checks if a user has all of the specified permissions
func (s *PermissionService) HasAllPermissions(ctx context.Context, tenantID, userID uuid.UUID, checks []PermissionCheck) (bool, error) {
	// Get user permissions (from cache or DB)
	permissions, err := s.GetUserPermissions(ctx, tenantID, userID)
	if err != nil {
		return false, err
	}

	// Check if user has all required permissions
	for _, check := range checks {
		found := false
		for _, perm := range permissions {
			if perm.Matches(check.Resource, check.Action) {
				found = true
				break
			}
		}
		if !found {
			return false, nil
		}
	}

	return true, nil
}

// PermissionCheck represents a permission to check
type PermissionCheck struct {
	Resource string
	Action   string
}

// InvalidateUserPermissions invalidates the permission cache for a user
func (s *PermissionService) InvalidateUserPermissions(ctx context.Context, tenantID, userID uuid.UUID) error {
	cacheKey := database.CacheKey(userPermissionKeyPrefix, tenantID.String(), userID.String())
	return s.redis.Del(ctx, cacheKey).Err()
}

// InvalidateRolePermissions invalidates permission cache for all users with a role
func (s *PermissionService) InvalidateRolePermissions(ctx context.Context, tenantID, roleID uuid.UUID) error {
	// Get all users with this role
	users, err := s.userRoleRepo.GetUsersByRole(ctx, tenantID, roleID)
	if err != nil {
		return err
	}

	// Invalidate cache for each user
	for _, user := range users {
		if err := s.InvalidateUserPermissions(ctx, tenantID, user.ID); err != nil {
			// Log error but continue
			fmt.Printf("Failed to invalidate cache for user %s: %v\n", user.ID, err)
		}
	}

	return nil
}

// InvalidateTenantPermissions invalidates all permission caches for a tenant
func (s *PermissionService) InvalidateTenantPermissions(ctx context.Context, tenantID uuid.UUID) error {
	pattern := database.CacheKey(userPermissionKeyPrefix, tenantID.String(), "*")
	return database.InvalidatePattern(ctx, s.redis, pattern)
}

// GetRolePermissions retrieves permissions for a role (with caching)
func (s *PermissionService) GetRolePermissions(ctx context.Context, tenantID, roleID uuid.UUID) ([]models.Permission, error) {
	// Try cache first
	cacheKey := database.CacheKey(rolePermissionKeyPrefix, tenantID.String(), roleID.String())
	cachedData, err := s.redis.Get(ctx, cacheKey).Result()

	if err == nil {
		// Cache hit
		var permissions []models.Permission
		if err := json.Unmarshal([]byte(cachedData), &permissions); err == nil {
			return permissions, nil
		}
	}

	// Cache miss - query database
	permissions, err := s.roleRepo.GetPermissions(ctx, tenantID, roleID)
	if err != nil {
		return nil, err
	}

	// Cache the result
	data, _ := json.Marshal(permissions)
	s.redis.Set(ctx, cacheKey, data, permissionCacheTTL)

	return permissions, nil
}

// ListAllPermissions retrieves all permissions (no caching needed - static data)
func (s *PermissionService) ListAllPermissions(ctx context.Context) ([]models.Permission, error) {
	return s.permissionRepo.List(ctx)
}

// ListPermissionsByCategory retrieves permissions grouped by category
func (s *PermissionService) ListPermissionsByCategory(ctx context.Context) ([]models.PermissionGroup, error) {
	return s.permissionRepo.ListByCategory(ctx)
}

// SearchPermissions searches for permissions
func (s *PermissionService) SearchPermissions(ctx context.Context, searchTerm string) ([]models.Permission, error) {
	return s.permissionRepo.Search(ctx, searchTerm)
}

// ValidatePermissionIDs validates that permission IDs exist
func (s *PermissionService) ValidatePermissionIDs(ctx context.Context, permissionIDs []uuid.UUID) (bool, []uuid.UUID, error) {
	return s.permissionRepo.ValidatePermissionIDs(ctx, permissionIDs)
}

// GetPermissionStats returns statistics about permissions
func (s *PermissionService) GetPermissionStats(ctx context.Context) (map[string]interface{}, error) {
	totalCount, err := s.permissionRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	categoryCount, err := s.permissionRepo.CountByCategory(ctx)
	if err != nil {
		return nil, err
	}

	resources, err := s.permissionRepo.GetResourceList(ctx)
	if err != nil {
		return nil, err
	}

	categories, err := s.permissionRepo.GetCategoryList(ctx)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_permissions": totalCount,
		"by_category":       categoryCount,
		"resources":         resources,
		"categories":        categories,
		"resource_count":    len(resources),
		"category_count":    len(categories),
	}, nil
}

// CheckUserRole checks if a user has a specific role
func (s *PermissionService) CheckUserRole(ctx context.Context, tenantID, userID uuid.UUID, roleName string) (bool, error) {
	// Get user's roles
	roles, err := s.userRoleRepo.GetUserRoles(ctx, tenantID, userID)
	if err != nil {
		return false, err
	}

	// Check if user has the role
	for _, role := range roles {
		if role.Name == roleName {
			return true, nil
		}
	}

	return false, nil
}

// IsOwner checks if a user is an owner
func (s *PermissionService) IsOwner(ctx context.Context, tenantID, userID uuid.UUID) (bool, error) {
	return s.CheckUserRole(ctx, tenantID, userID, models.RoleOwner)
}

// IsAdmin checks if a user is an admin or owner
func (s *PermissionService) IsAdmin(ctx context.Context, tenantID, userID uuid.UUID) (bool, error) {
	roles, err := s.userRoleRepo.GetUserRoles(ctx, tenantID, userID)
	if err != nil {
		return false, err
	}

	for _, role := range roles {
		if role.Name == models.RoleOwner || role.Name == models.RoleAdmin {
			return true, nil
		}
	}

	return false, nil
}

// GetUserRoleNames returns role names for a user
func (s *PermissionService) GetUserRoleNames(ctx context.Context, tenantID, userID uuid.UUID) ([]string, error) {
	roles, err := s.userRoleRepo.GetUserRoles(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}

	roleNames := make([]string, len(roles))
	for i, role := range roles {
		roleNames[i] = role.Name
	}

	return roleNames, nil
}

// ComparePermissions compares permissions between two users
func (s *PermissionService) ComparePermissions(ctx context.Context, tenantID, user1ID, user2ID uuid.UUID) (map[string]interface{}, error) {
	user1Perms, err := s.GetUserPermissions(ctx, tenantID, user1ID)
	if err != nil {
		return nil, err
	}

	user2Perms, err := s.GetUserPermissions(ctx, tenantID, user2ID)
	if err != nil {
		return nil, err
	}

	// Create permission maps for comparison
	user1PermMap := make(map[string]bool)
	user2PermMap := make(map[string]bool)

	for _, perm := range user1Perms {
		user1PermMap[perm.String()] = true
	}

	for _, perm := range user2Perms {
		user2PermMap[perm.String()] = true
	}

	// Find common, user1-only, and user2-only permissions
	common := []string{}
	user1Only := []string{}
	user2Only := []string{}

	for permStr := range user1PermMap {
		if user2PermMap[permStr] {
			common = append(common, permStr)
		} else {
			user1Only = append(user1Only, permStr)
		}
	}

	for permStr := range user2PermMap {
		if !user1PermMap[permStr] {
			user2Only = append(user2Only, permStr)
		}
	}

	return map[string]interface{}{
		"common_permissions":       common,
		"user1_only_permissions":   user1Only,
		"user2_only_permissions":   user2Only,
		"user1_permission_count":   len(user1Perms),
		"user2_permission_count":   len(user2Perms),
		"common_permission_count":  len(common),
	}, nil
}
