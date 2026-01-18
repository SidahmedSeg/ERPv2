DROP TRIGGER IF EXISTS update_tenants_updated_at ON tenants;
DROP TABLE IF EXISTS tenants CASCADE;
DROP FUNCTION IF EXISTS update_updated_at_column();
