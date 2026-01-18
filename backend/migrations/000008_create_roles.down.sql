DROP TRIGGER IF EXISTS update_roles_updated_at ON roles;
DROP POLICY IF EXISTS bypass_rls_for_superuser ON roles;
DROP POLICY IF EXISTS tenant_isolation ON roles;
ALTER TABLE roles DISABLE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS roles CASCADE;
