-- Role-Permissions junction table (with RLS)
CREATE TABLE role_permissions (
    id UUID DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    role_id UUID NOT NULL,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID,

    PRIMARY KEY (tenant_id, id),
    CONSTRAINT unique_role_permission UNIQUE(tenant_id, role_id, permission_id),
    FOREIGN KEY (tenant_id, role_id) REFERENCES roles(tenant_id, id) ON DELETE CASCADE
);

-- Indexes
CREATE INDEX idx_role_permissions_tenant_role ON role_permissions(tenant_id, role_id);
CREATE INDEX idx_role_permissions_permission ON role_permissions(permission_id);
CREATE INDEX idx_role_permissions_tenant ON role_permissions(tenant_id);

-- Enable RLS
ALTER TABLE role_permissions ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON role_permissions
    USING (tenant_id = current_setting('app.current_tenant_id', true)::uuid);

CREATE POLICY bypass_rls_for_superuser ON role_permissions
    USING (current_setting('app.bypass_rls', true) = 'true')
    WITH CHECK (current_setting('app.bypass_rls', true) = 'true');

-- Comments
COMMENT ON TABLE role_permissions IS 'Maps roles to permissions - RLS enforced';
