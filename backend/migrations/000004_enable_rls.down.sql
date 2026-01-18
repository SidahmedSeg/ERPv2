DROP POLICY IF EXISTS bypass_rls_for_superuser ON users;
DROP POLICY IF EXISTS tenant_isolation ON users;
ALTER TABLE users DISABLE ROW LEVEL SECURITY;
