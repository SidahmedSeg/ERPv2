DROP POLICY IF EXISTS bypass_rls_for_superuser ON user_roles;
DROP POLICY IF EXISTS tenant_isolation ON user_roles;
ALTER TABLE user_roles DISABLE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS user_roles CASCADE;
