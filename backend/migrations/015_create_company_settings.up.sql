-- Create company_settings table
CREATE TABLE IF NOT EXISTS company_settings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,

    -- Company Information
    company_name VARCHAR(255) NOT NULL,
    legal_business_name VARCHAR(255),
    industry VARCHAR(100),
    speciality VARCHAR(100),
    company_size VARCHAR(50),
    founded_date DATE,
    website_url TEXT,
    logo_url TEXT,

    -- Contact Details
    primary_email VARCHAR(255),
    support_email VARCHAR(255),
    phone_number VARCHAR(50),
    fax VARCHAR(50),

    -- Address
    street_address TEXT,
    city VARCHAR(100),
    state VARCHAR(100),
    postal_code VARCHAR(20),
    country VARCHAR(100),

    -- Business Hours
    timezone VARCHAR(100) NOT NULL DEFAULT 'UTC',
    working_days JSONB NOT NULL DEFAULT '{"monday": true, "tuesday": true, "wednesday": true, "thursday": true, "friday": true, "saturday": false, "sunday": false}',
    working_hours_start TIME NOT NULL DEFAULT '09:00',
    working_hours_end TIME NOT NULL DEFAULT '17:00',

    -- Fiscal Settings
    fiscal_year_start VARCHAR(10),
    default_currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    date_format VARCHAR(20) NOT NULL DEFAULT 'DD/MM/YYYY',
    number_format VARCHAR(20) NOT NULL DEFAULT '1,000.00',

    -- Tax/Legal IDs (Algeria specific but can be used globally)
    rc_number VARCHAR(50),
    nif_number VARCHAR(50),
    nis_number VARCHAR(50),
    ai_number VARCHAR(50),
    capital_social DECIMAL(15,2),

    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID,
    updated_by UUID,

    CONSTRAINT unique_settings_per_tenant UNIQUE(tenant_id)
);

-- Indexes
CREATE INDEX idx_company_settings_tenant ON company_settings(tenant_id);

-- Enable RLS
ALTER TABLE company_settings ENABLE ROW LEVEL SECURITY;

-- RLS Policy: Users can only see settings in their tenant
CREATE POLICY tenant_isolation ON company_settings
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

-- Trigger: Update updated_at
CREATE TRIGGER update_company_settings_updated_at
    BEFORE UPDATE ON company_settings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
