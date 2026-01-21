-- Create departments table
-- Departments provide organizational structure for team members

CREATE TABLE departments (
    id UUID DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,

    -- Basic Info
    name VARCHAR(255) NOT NULL,
    description TEXT,

    -- Leadership
    head_user_id UUID,  -- Nullable - references users(id) in same tenant

    -- Visual Customization
    color VARCHAR(7) NOT NULL DEFAULT '#3B82F6',  -- Hex color code
    icon VARCHAR(50) NOT NULL DEFAULT 'users',     -- Icon identifier

    -- Status: active | inactive
    status VARCHAR(20) DEFAULT 'active',

    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID,  -- References users(id) in same tenant

    PRIMARY KEY (tenant_id, id),
    CONSTRAINT unique_department_name_per_tenant UNIQUE(tenant_id, name),
    CONSTRAINT valid_status CHECK (status IN ('active', 'inactive')),
    FOREIGN KEY (tenant_id, head_user_id) REFERENCES users(tenant_id, id) ON DELETE SET NULL
);

-- Indexes for performance
CREATE INDEX idx_departments_tenant_id ON departments(tenant_id);
CREATE INDEX idx_departments_status ON departments(tenant_id, status);
CREATE INDEX idx_departments_head_user_id ON departments(tenant_id, head_user_id) WHERE head_user_id IS NOT NULL;
CREATE INDEX idx_departments_name ON departments(tenant_id, name);
CREATE INDEX idx_departments_created_at ON departments(tenant_id, created_at DESC);

-- Enable RLS
ALTER TABLE departments ENABLE ROW LEVEL SECURITY;

-- RLS Policy: Users can only see departments in their tenant
CREATE POLICY tenant_isolation ON departments
    USING (tenant_id = current_setting('app.current_tenant_id', true)::uuid);

-- RLS Policy: Allow bypass for superuser operations
CREATE POLICY bypass_rls_for_superuser ON departments
    USING (current_setting('app.bypass_rls', true) = 'true')
    WITH CHECK (current_setting('app.bypass_rls', true) = 'true');

-- Trigger for updated_at
CREATE TRIGGER update_departments_updated_at
    BEFORE UPDATE ON departments
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Comments
COMMENT ON TABLE departments IS 'Departments for organizational structure - RLS enforced';
COMMENT ON COLUMN departments.head_user_id IS 'Department head - nullable FK to users';
COMMENT ON COLUMN departments.color IS 'Hex color code for UI customization';
COMMENT ON COLUMN departments.icon IS 'Icon identifier for UI display';

-- Add department permissions to permissions table
INSERT INTO permissions (resource, action, display_name, description, category) VALUES
    ('departments', 'view', 'View Departments', 'View department information', 'Organization'),
    ('departments', 'create', 'Create Departments', 'Create new departments', 'Organization'),
    ('departments', 'edit', 'Edit Departments', 'Edit department information', 'Organization'),
    ('departments', 'delete', 'Delete Departments', 'Delete departments', 'Organization'),
    ('departments', 'manage_members', 'Manage Department Members', 'Add/remove users from departments', 'Organization')
ON CONFLICT (resource, action) DO NOTHING;
