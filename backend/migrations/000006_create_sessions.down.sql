DROP POLICY IF EXISTS bypass_rls_for_superuser ON sessions;
DROP POLICY IF EXISTS tenant_isolation ON sessions;
ALTER TABLE sessions DISABLE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS sessions CASCADE;
