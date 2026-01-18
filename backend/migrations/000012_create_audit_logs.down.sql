DROP POLICY IF EXISTS bypass_rls_for_superuser ON audit_logs;
DROP POLICY IF EXISTS tenant_isolation ON audit_logs;
ALTER TABLE audit_logs DISABLE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS audit_logs CASCADE;
