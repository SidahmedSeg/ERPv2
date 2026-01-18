DROP POLICY IF EXISTS bypass_rls_for_superuser ON invitations;
DROP POLICY IF EXISTS tenant_isolation ON invitations;
ALTER TABLE invitations DISABLE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS invitations CASCADE;
