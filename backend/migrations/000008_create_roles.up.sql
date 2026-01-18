-- Roles table (with RLS)
CREATE TABLE roles (
    id UUID DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,

    name VARCHAR(100) NOT NULL,  -- Unique per tenant
    display_name VARCHAR(255) NOT NULL,
    description TEXT,

    -- Hierarchy support (for future ABAC)
    parent_role_id UUID,  -- References roles(id) in same tenant
    level INT DEFAULT 0,   -- 0 = root, higher = more restrictive

    -- System roles cannot be deleted/modified
    is_system BOOLEAN DEFAULT FALSE,

    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID,

    PRIMARY KEY (tenant_id, id),
    CONSTRAINT unique_role_name_per_tenant UNIQUE(tenant_id, name),
    FOREIGN KEY (tenant_id, parent_role_id) REFERENCES roles(tenant_id, id) ON DELETE SET NULL
);

-- Indexes
CREATE INDEX idx_roles_tenant ON roles(tenant_id);
CREATE INDEX idx_roles_is_system ON roles(tenant_id, is_system);
CREATE INDEX idx_roles_parent ON roles(tenant_id, parent_role_id);
CREATE INDEX idx_roles_name ON roles(tenant_id, name);

-- Enable RLS
ALTER TABLE roles ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON roles
    USING (tenant_id = current_setting('app.current_tenant_id', true)::uuid);

CREATE POLICY bypass_rls_for_superuser ON roles
    USING (current_setting('app.bypass_rls', true) = 'true')
    WITH CHECK (current_setting('app.bypass_rls', true) = 'true');

-- Trigger for updated_at
CREATE TRIGGER update_roles_updated_at
    BEFORE UPDATE ON roles
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Comments
COMMENT ON TABLE roles IS 'Tenant-specific roles - RLS enforced';
COMMENT ON COLUMN roles.is_system IS 'System roles cannot be modified or deleted';
COMMENT ON COLUMN roles.parent_role_id IS 'Parent role for hierarchical RBAC (future)';
COMMENT ON COLUMN roles.level IS 'Hierarchy level: 0 = root, higher = child';
