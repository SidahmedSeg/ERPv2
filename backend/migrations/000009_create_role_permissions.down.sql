DROP POLICY IF EXISTS bypass_rls_for_superuser ON role_permissions;
DROP POLICY IF EXISTS tenant_isolation ON role_permissions;
ALTER TABLE role_permissions DISABLE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS role_permissions CASCADE;
