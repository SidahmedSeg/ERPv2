DROP INDEX IF EXISTS idx_users_2fa_enabled;
ALTER TABLE users DROP COLUMN IF EXISTS two_factor_recovery_email;
ALTER TABLE users DROP COLUMN IF EXISTS two_factor_enabled_at;
ALTER TABLE users DROP COLUMN IF EXISTS two_factor_backup_codes;
ALTER TABLE users DROP COLUMN IF EXISTS two_factor_secret;
ALTER TABLE users DROP COLUMN IF EXISTS two_factor_enabled;
