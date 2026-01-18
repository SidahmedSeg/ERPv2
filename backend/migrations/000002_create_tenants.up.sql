-- Create trigger function for updating updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Tenants table (central registry)
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    slug VARCHAR(63) UNIQUE NOT NULL,  -- DNS-safe subdomain
    company_name VARCHAR(255) NOT NULL,

    -- Status: pending_verification | active | suspended | canceled
    status VARCHAR(30) NOT NULL DEFAULT 'pending_verification',

    -- Email verification
    email VARCHAR(255) NOT NULL,
    email_verified BOOLEAN DEFAULT FALSE,
    email_verified_at TIMESTAMPTZ,
    verification_token UUID UNIQUE,
    verification_token_expires_at TIMESTAMPTZ,

    -- Subscription
    plan_tier VARCHAR(30) DEFAULT 'free',  -- free | starter | professional | enterprise
    trial_ends_at TIMESTAMPTZ,

    -- Settings (JSONB for flexibility)
    settings JSONB DEFAULT '{}'::jsonb,

    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    activated_at TIMESTAMPTZ,
    suspended_at TIMESTAMPTZ,

    CONSTRAINT valid_status CHECK (status IN ('pending_verification', 'active', 'suspended', 'canceled')),
    CONSTRAINT valid_plan CHECK (plan_tier IN ('free', 'starter', 'professional', 'enterprise'))
);

-- Indexes for performance
CREATE INDEX idx_tenants_slug ON tenants(slug);
CREATE INDEX idx_tenants_status ON tenants(status) WHERE status = 'active';
CREATE INDEX idx_tenants_email ON tenants(email);
CREATE INDEX idx_tenants_verification_token ON tenants(verification_token) WHERE verification_token IS NOT NULL;

-- Trigger for updated_at
CREATE TRIGGER update_tenants_updated_at
    BEFORE UPDATE ON tenants
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Comment on table
COMMENT ON TABLE tenants IS 'Central tenant registry - no RLS applied';
COMMENT ON COLUMN tenants.slug IS 'Subdomain identifier (e.g., acme-corp)';
COMMENT ON COLUMN tenants.verification_token IS 'Email verification token (24h expiry)';
