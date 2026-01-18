-- Invitations table (with RLS)
CREATE TABLE invitations (
    id UUID DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,

    email VARCHAR(255) NOT NULL,
    token UUID UNIQUE NOT NULL DEFAULT uuid_generate_v4(),

    -- Roles to assign upon acceptance
    role_ids UUID[] NOT NULL,

    -- Status: pending | accepted | expired | revoked
    status VARCHAR(20) DEFAULT 'pending',

    -- Welcome message
    message TEXT,

    -- Metadata
    invited_by UUID NOT NULL,  -- References users(id)
    invited_at TIMESTAMPTZ DEFAULT NOW(),
    accepted_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ NOT NULL,

    PRIMARY KEY (tenant_id, id),
    FOREIGN KEY (tenant_id, invited_by) REFERENCES users(tenant_id, id) ON DELETE CASCADE,
    CONSTRAINT valid_status CHECK (status IN ('pending', 'accepted', 'expired', 'revoked'))
);

-- Unique constraint: only one pending invitation per email per tenant
CREATE UNIQUE INDEX idx_invitations_unique_pending
    ON invitations(tenant_id, email)
    WHERE status = 'pending';

-- Other indexes
CREATE INDEX idx_invitations_tenant ON invitations(tenant_id);
CREATE INDEX idx_invitations_email ON invitations(tenant_id, email);
CREATE INDEX idx_invitations_token ON invitations(token) WHERE status = 'pending';
CREATE INDEX idx_invitations_status ON invitations(tenant_id, status);
CREATE INDEX idx_invitations_expires_at ON invitations(expires_at);

-- Enable RLS
ALTER TABLE invitations ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON invitations
    USING (tenant_id = current_setting('app.current_tenant_id', true)::uuid);

CREATE POLICY bypass_rls_for_superuser ON invitations
    USING (current_setting('app.bypass_rls', true) = 'true')
    WITH CHECK (current_setting('app.bypass_rls', true) = 'true');

-- Comments
COMMENT ON TABLE invitations IS 'Team invitation system - RLS enforced';
COMMENT ON COLUMN invitations.token IS 'Unique invitation token (sent via email)';
COMMENT ON COLUMN invitations.role_ids IS 'Array of role IDs to assign upon acceptance';
COMMENT ON COLUMN invitations.expires_at IS 'Invitation expiry (typically 7 days)';
