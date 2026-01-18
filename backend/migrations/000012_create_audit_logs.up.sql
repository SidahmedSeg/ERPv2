-- Audit logs table for security events (with RLS)
CREATE TABLE audit_logs (
    id UUID DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID,  -- Can be NULL for system events

    -- Event details
    action VARCHAR(100) NOT NULL,  -- e.g., user.login, user.created, role.assigned
    resource_type VARCHAR(100),     -- e.g., user, role, session
    resource_id UUID,

    -- Request context
    ip_address INET,
    user_agent TEXT,

    -- Status: success | failure
    status VARCHAR(20) NOT NULL,

    -- Additional context (flexible)
    metadata JSONB DEFAULT '{}'::jsonb,

    -- Timestamp
    created_at TIMESTAMPTZ DEFAULT NOW(),

    PRIMARY KEY (tenant_id, id),
    FOREIGN KEY (tenant_id, user_id) REFERENCES users(tenant_id, id) ON DELETE SET NULL,
    CONSTRAINT valid_status CHECK (status IN ('success', 'failure'))
);

-- Indexes for querying audit logs
CREATE INDEX idx_audit_logs_tenant ON audit_logs(tenant_id);
CREATE INDEX idx_audit_logs_user ON audit_logs(tenant_id, user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(tenant_id, created_at DESC);
CREATE INDEX idx_audit_logs_resource ON audit_logs(tenant_id, resource_type, resource_id);
CREATE INDEX idx_audit_logs_status ON audit_logs(tenant_id, status);

-- GIN index for JSONB metadata queries
CREATE INDEX idx_audit_logs_metadata ON audit_logs USING GIN(metadata);

-- Enable RLS
ALTER TABLE audit_logs ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON audit_logs
    USING (tenant_id = current_setting('app.current_tenant_id', true)::uuid);

CREATE POLICY bypass_rls_for_superuser ON audit_logs
    USING (current_setting('app.bypass_rls', true) = 'true')
    WITH CHECK (current_setting('app.bypass_rls', true) = 'true');

-- Comments
COMMENT ON TABLE audit_logs IS 'Security audit trail - RLS enforced';
COMMENT ON COLUMN audit_logs.action IS 'Action performed (e.g., user.login, role.created)';
COMMENT ON COLUMN audit_logs.metadata IS 'Additional context as JSON (e.g., changed fields)';

-- Note: For high-volume tenants, consider partitioning by created_at (monthly)
-- Example: CREATE TABLE audit_logs_2026_01 PARTITION OF audit_logs
--          FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');
