-- Permissions table (global catalog - no RLS)
CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Resource.Action format
    resource VARCHAR(100) NOT NULL,  -- e.g., users, roles, settings
    action VARCHAR(100) NOT NULL,    -- e.g., view, create, edit, delete, *

    display_name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100),  -- For UI grouping

    created_at TIMESTAMPTZ DEFAULT NOW(),

    CONSTRAINT unique_resource_action UNIQUE(resource, action)
);

-- Indexes
CREATE INDEX idx_permissions_resource ON permissions(resource);
CREATE INDEX idx_permissions_category ON permissions(category);
CREATE INDEX idx_permissions_resource_action ON permissions(resource, action);

-- Seed default permissions
INSERT INTO permissions (resource, action, display_name, description, category) VALUES
    -- Users
    ('users', 'view', 'View Users', 'View user list and details', 'User Management'),
    ('users', 'create', 'Create Users', 'Invite new users', 'User Management'),
    ('users', 'edit', 'Edit Users', 'Update user information', 'User Management'),
    ('users', 'delete', 'Delete Users', 'Remove users', 'User Management'),
    ('users', 'manage_status', 'Manage User Status', 'Activate/suspend users', 'User Management'),
    ('users', '*', 'All User Permissions', 'Full user management access', 'User Management'),

    -- Roles
    ('roles', 'view', 'View Roles', 'View role list', 'Access Control'),
    ('roles', 'create', 'Create Roles', 'Create custom roles', 'Access Control'),
    ('roles', 'edit', 'Edit Roles', 'Modify role permissions', 'Access Control'),
    ('roles', 'delete', 'Delete Roles', 'Remove custom roles', 'Access Control'),
    ('roles', 'assign', 'Assign Roles', 'Assign roles to users', 'Access Control'),
    ('roles', '*', 'All Role Permissions', 'Full role management', 'Access Control'),

    -- Settings
    ('settings', 'view', 'View Settings', 'View company settings', 'Settings'),
    ('settings', 'edit', 'Edit Settings', 'Update company settings', 'Settings'),
    ('settings', '*', 'All Settings Permissions', 'Full settings access', 'Settings'),

    -- Security
    ('security', 'view_logs', 'View Security Logs', 'View audit logs', 'Security'),
    ('security', 'view_sessions', 'View Sessions', 'View active sessions', 'Security'),
    ('security', 'manage_sessions', 'Manage Sessions', 'Revoke user sessions', 'Security'),
    ('security', '*', 'All Security Permissions', 'Full security access', 'Security');

-- Comments
COMMENT ON TABLE permissions IS 'Global permission catalog - shared across all tenants, no RLS';
COMMENT ON COLUMN permissions.resource IS 'Module or resource name (e.g., users, products)';
COMMENT ON COLUMN permissions.action IS 'Action type (view, create, edit, delete, *) - wildcard grants all';
