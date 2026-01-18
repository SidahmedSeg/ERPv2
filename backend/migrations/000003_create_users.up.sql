-- Users table (with RLS)
CREATE TABLE users (
    id UUID DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,

    -- Authentication
    email VARCHAR(255) NOT NULL,
    email_verified BOOLEAN DEFAULT FALSE,
    password_hash VARCHAR(255) NOT NULL,

    -- Profile
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    phone VARCHAR(20),
    avatar_url TEXT,

    -- Status: active | suspended | deactivated | pending
    status VARCHAR(20) DEFAULT 'active',

    -- Password reset
    reset_token UUID,
    reset_token_expires_at TIMESTAMPTZ,

    -- Activity tracking
    last_login_at TIMESTAMPTZ,
    last_login_ip INET,
    last_active_at TIMESTAMPTZ,

    -- Preferences
    timezone VARCHAR(50) DEFAULT 'UTC',
    language VARCHAR(10) DEFAULT 'en',
    preferences JSONB DEFAULT '{}'::jsonb,

    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID,  -- References users(id) in same tenant

    PRIMARY KEY (tenant_id, id),
    CONSTRAINT unique_email_per_tenant UNIQUE(tenant_id, email),
    CONSTRAINT valid_status CHECK (status IN ('active', 'suspended', 'deactivated', 'pending'))
);

-- Indexes for performance
CREATE INDEX idx_users_tenant_id ON users(tenant_id);
CREATE INDEX idx_users_email ON users(tenant_id, email);
CREATE INDEX idx_users_status ON users(tenant_id, status);
CREATE INDEX idx_users_reset_token ON users(reset_token) WHERE reset_token IS NOT NULL;
CREATE INDEX idx_users_created_at ON users(tenant_id, created_at DESC);

-- Trigger for updated_at
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Comments
COMMENT ON TABLE users IS 'User accounts - RLS enforced for tenant isolation';
COMMENT ON COLUMN users.tenant_id IS 'Foreign key to tenants table - used for RLS';
COMMENT ON COLUMN users.password_hash IS 'bcrypt hashed password';
