-- Helper function: Get all permissions for a user
CREATE OR REPLACE FUNCTION get_user_permissions(p_tenant_id UUID, p_user_id UUID)
RETURNS TABLE(
    resource VARCHAR,
    action VARCHAR,
    permission_display_name VARCHAR,
    role_name VARCHAR
) AS $$
BEGIN
    RETURN QUERY
    SELECT DISTINCT
        p.resource,
        p.action,
        p.display_name AS permission_display_name,
        r.name AS role_name
    FROM user_roles ur
    JOIN roles r ON ur.tenant_id = r.tenant_id AND ur.role_id = r.id
    JOIN role_permissions rp ON r.tenant_id = rp.tenant_id AND r.id = rp.role_id
    JOIN permissions p ON rp.permission_id = p.id
    WHERE ur.tenant_id = p_tenant_id
      AND ur.user_id = p_user_id;
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

-- Helper function: Check if user has specific permission
CREATE OR REPLACE FUNCTION has_permission(
    p_tenant_id UUID,
    p_user_id UUID,
    p_resource VARCHAR,
    p_action VARCHAR
) RETURNS BOOLEAN AS $$
DECLARE
    v_has_permission BOOLEAN;
BEGIN
    SELECT EXISTS (
        SELECT 1
        FROM user_roles ur
        JOIN roles r ON ur.tenant_id = r.tenant_id AND ur.role_id = r.id
        JOIN role_permissions rp ON r.tenant_id = rp.tenant_id AND r.id = rp.role_id
        JOIN permissions p ON rp.permission_id = p.id
        WHERE ur.tenant_id = p_tenant_id
          AND ur.user_id = p_user_id
          AND p.resource = p_resource
          AND (p.action = p_action OR p.action = '*')  -- Wildcard support
    ) INTO v_has_permission;

    RETURN v_has_permission;
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

-- Helper function: Get role permissions
CREATE OR REPLACE FUNCTION get_role_permissions(p_tenant_id UUID, p_role_id UUID)
RETURNS TABLE(
    permission_id UUID,
    resource VARCHAR,
    action VARCHAR,
    display_name VARCHAR
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        p.id AS permission_id,
        p.resource,
        p.action,
        p.display_name
    FROM role_permissions rp
    JOIN permissions p ON rp.permission_id = p.id
    WHERE rp.tenant_id = p_tenant_id
      AND rp.role_id = p_role_id
    ORDER BY p.category, p.resource, p.action;
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

-- Comments
COMMENT ON FUNCTION get_user_permissions IS 'Returns all permissions for a user (aggregated from all roles)';
COMMENT ON FUNCTION has_permission IS 'Check if user has a specific permission (supports wildcard)';
COMMENT ON FUNCTION get_role_permissions IS 'Returns all permissions assigned to a role';
