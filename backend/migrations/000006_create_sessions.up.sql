-- Sessions table for tracking user sessions and devices
CREATE TABLE sessions (
    id UUID DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,

    -- Token (hashed for security)
    token_hash VARCHAR(255) UNIQUE NOT NULL,

    -- Device info
    device_type VARCHAR(50),      -- Desktop | Mobile | Tablet
    browser VARCHAR(100),
    os VARCHAR(100),
    ip_address INET,
    user_agent TEXT,

    -- Location (GeoIP - optional)
    country_code CHAR(2),
    city VARCHAR(100),

    -- Activity
    last_activity_at TIMESTAMPTZ DEFAULT NOW(),

    -- Expiry
    expires_at TIMESTAMPTZ NOT NULL,

    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW(),

    PRIMARY KEY (tenant_id, id),
    FOREIGN KEY (tenant_id, user_id) REFERENCES users(tenant_id, id) ON DELETE CASCADE
);

-- Indexes
CREATE INDEX idx_sessions_tenant_user ON sessions(tenant_id, user_id);
CREATE INDEX idx_sessions_token_hash ON sessions(token_hash);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX idx_sessions_last_activity ON sessions(tenant_id, user_id, last_activity_at DESC);

-- Enable RLS
ALTER TABLE sessions ENABLE ROW LEVEL SECURITY;

-- RLS Policy
CREATE POLICY tenant_isolation ON sessions
    USING (tenant_id = current_setting('app.current_tenant_id', true)::uuid);

CREATE POLICY bypass_rls_for_superuser ON sessions
    USING (current_setting('app.bypass_rls', true) = 'true')
    WITH CHECK (current_setting('app.bypass_rls', true) = 'true');

-- Comments
COMMENT ON TABLE sessions IS 'User session tracking with device information - RLS enforced';
COMMENT ON COLUMN sessions.token_hash IS 'SHA-256 hash of JWT token for validation';
COMMENT ON COLUMN sessions.device_type IS 'Parsed from user agent: Desktop, Mobile, or Tablet';
