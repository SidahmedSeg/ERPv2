-- Enable Row-Level Security on users table
ALTER TABLE users ENABLE ROW LEVEL SECURITY;

-- Policy: Users can only see users in their tenant
CREATE POLICY tenant_isolation ON users
    USING (tenant_id = current_setting('app.current_tenant_id', true)::uuid);

-- Policy: Bypass RLS for superuser operations (migrations, admin tools)
CREATE POLICY bypass_rls_for_superuser ON users
    USING (current_setting('app.bypass_rls', true) = 'true')
    WITH CHECK (current_setting('app.bypass_rls', true) = 'true');

-- Comment
COMMENT ON POLICY tenant_isolation ON users IS 'Enforce tenant isolation - users only see data from their tenant';
COMMENT ON POLICY bypass_rls_for_superuser ON users IS 'Allow administrative operations to bypass RLS when app.bypass_rls = true';
