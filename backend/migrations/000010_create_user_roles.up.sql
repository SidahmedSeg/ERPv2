-- User-Roles junction table (with RLS)
CREATE TABLE user_roles (
    id UUID DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    role_id UUID NOT NULL,

    assigned_at TIMESTAMPTZ DEFAULT NOW(),
    assigned_by UUID,  -- References users(id)

    PRIMARY KEY (tenant_id, id),
    CONSTRAINT unique_user_role UNIQUE(tenant_id, user_id, role_id),
    FOREIGN KEY (tenant_id, user_id) REFERENCES users(tenant_id, id) ON DELETE CASCADE,
    FOREIGN KEY (tenant_id, role_id) REFERENCES roles(tenant_id, id) ON DELETE CASCADE
);

-- Indexes
CREATE INDEX idx_user_roles_tenant_user ON user_roles(tenant_id, user_id);
CREATE INDEX idx_user_roles_role ON user_roles(tenant_id, role_id);
CREATE INDEX idx_user_roles_assigned_at ON user_roles(tenant_id, assigned_at DESC);

-- Enable RLS
ALTER TABLE user_roles ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON user_roles
    USING (tenant_id = current_setting('app.current_tenant_id', true)::uuid);

CREATE POLICY bypass_rls_for_superuser ON user_roles
    USING (current_setting('app.bypass_rls', true) = 'true')
    WITH CHECK (current_setting('app.bypass_rls', true) = 'true');

-- Comments
COMMENT ON TABLE user_roles IS 'Maps users to roles - RLS enforced';
COMMENT ON COLUMN user_roles.assigned_by IS 'User who assigned this role';
