-- Assign department permissions to existing tenant roles
-- This fixes the issue where owner/admin roles don't have departments permissions

-- Assign ALL departments permissions to all existing Owner roles
INSERT INTO role_permissions (tenant_id, role_id, permission_id)
SELECT r.tenant_id, r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'owner'
  AND p.resource = 'departments'
  AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp
    WHERE rp.role_id = r.id AND rp.permission_id = p.id
  );

-- Assign view, create, edit, manage_members permissions to Admin roles
INSERT INTO role_permissions (tenant_id, role_id, permission_id)
SELECT r.tenant_id, r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'admin'
  AND p.resource = 'departments'
  AND p.action IN ('view', 'create', 'edit', 'manage_members')
  AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp
    WHERE rp.role_id = r.id AND rp.permission_id = p.id
  );

-- Assign view and edit permissions to Manager roles
INSERT INTO role_permissions (tenant_id, role_id, permission_id)
SELECT r.tenant_id, r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'manager'
  AND p.resource = 'departments'
  AND p.action IN ('view', 'edit')
  AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp
    WHERE rp.role_id = r.id AND rp.permission_id = p.id
  );

-- Assign view permission to User roles
INSERT INTO role_permissions (tenant_id, role_id, permission_id)
SELECT r.tenant_id, r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'user'
  AND p.resource = 'departments'
  AND p.action = 'view'
  AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp
    WHERE rp.role_id = r.id AND rp.permission_id = p.id
  );

-- Update the provision_tenant_system_roles function to include departments for new tenants
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

    -- Also assign all specific permissions (including departments)
    INSERT INTO role_permissions (tenant_id, role_id, permission_id)
    SELECT p_tenant_id, v_owner_role_id, id
    FROM permissions;

    -- Create Admin role (most permissions except some sensitive security operations)
    INSERT INTO roles (tenant_id, name, display_name, description, is_system, level)
    VALUES (p_tenant_id, 'admin', 'Administrator', 'Administrative access - manage users, roles, settings', TRUE, 1)
    RETURNING id INTO v_admin_role_id;

    -- Assign permissions to Admin role
    INSERT INTO role_permissions (tenant_id, role_id, permission_id)
    SELECT p_tenant_id, v_admin_role_id, id
    FROM permissions
    WHERE resource IN ('users', 'roles', 'settings', 'departments')
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
       OR (resource = 'settings' AND action = 'view')
       OR (resource = 'departments' AND action IN ('view', 'edit'));

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

COMMENT ON FUNCTION provision_tenant_system_roles IS 'Creates system roles (owner, admin, manager, user) for a new tenant during provisioning - UPDATED to include departments permissions';
