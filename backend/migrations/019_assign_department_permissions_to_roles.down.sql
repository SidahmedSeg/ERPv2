-- Rollback: Remove department permissions from roles

DELETE FROM role_permissions
WHERE permission_id IN (
    SELECT id FROM permissions WHERE resource = 'departments'
);

-- Restore original provision_tenant_system_roles function without departments
CREATE OR REPLACE FUNCTION provision_tenant_system_roles(p_tenant_id UUID)
RETURNS VOID AS $$
DECLARE
    v_owner_role_id UUID;
    v_admin_role_id UUID;
    v_manager_role_id UUID;
    v_user_role_id UUID;
BEGIN
    -- Create Owner role (full access)
    INSERT INTO roles (tenant_id, name, display_name, description, is_system, level)
    VALUES (p_tenant_id, 'owner', 'Owner', 'Full system access - all permissions', TRUE, 0)
    RETURNING id INTO v_owner_role_id;

    -- Assign all permissions to Owner role
    INSERT INTO role_permissions (tenant_id, role_id, permission_id)
    SELECT p_tenant_id, v_owner_role_id, id
    FROM permissions
    WHERE action = '*';  -- Grant wildcard permissions for all resources

    -- Create Admin role (most permissions except some sensitive security operations)
    INSERT INTO roles (tenant_id, name, display_name, description, is_system, level)
    VALUES (p_tenant_id, 'admin', 'Administrator', 'Administrative access - manage users, roles, settings', TRUE, 1)
    RETURNING id INTO v_admin_role_id;

    -- Assign permissions to Admin role
    INSERT INTO role_permissions (tenant_id, role_id, permission_id)
    SELECT p_tenant_id, v_admin_role_id, id
    FROM permissions
    WHERE resource IN ('users', 'roles', 'settings')
      AND action != 'delete';  -- Admins can't delete (only owners)

    -- Create Manager role (limited permissions)
    INSERT INTO roles (tenant_id, name, display_name, description, is_system, level)
    VALUES (p_tenant_id, 'manager', 'Manager', 'View and manage team members', TRUE, 2)
    RETURNING id INTO v_manager_role_id;

    -- Assign permissions to Manager role
    INSERT INTO role_permissions (tenant_id, role_id, permission_id)
    SELECT p_tenant_id, v_manager_role_id, id
    FROM permissions
    WHERE (resource = 'users' AND action IN ('view', 'edit'))
       OR (resource = 'settings' AND action = 'view');

    -- Create User role (basic read-only access)
    INSERT INTO roles (tenant_id, name, display_name, description, is_system, level)
    VALUES (p_tenant_id, 'user', 'User', 'Basic user access - view only', TRUE, 3)
    RETURNING id INTO v_user_role_id;

    -- Assign permissions to User role
    INSERT INTO role_permissions (tenant_id, role_id, permission_id)
    SELECT p_tenant_id, v_user_role_id, id
    FROM permissions
    WHERE action = 'view';  -- View-only permissions

END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
