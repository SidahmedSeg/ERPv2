-- Add Two-Factor Authentication fields to users table
ALTER TABLE users ADD COLUMN two_factor_enabled BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN two_factor_secret TEXT;  -- Encrypted TOTP secret
ALTER TABLE users ADD COLUMN two_factor_backup_codes TEXT[];  -- Encrypted array of backup codes
ALTER TABLE users ADD COLUMN two_factor_enabled_at TIMESTAMPTZ;
ALTER TABLE users ADD COLUMN two_factor_recovery_email VARCHAR(255);

-- Index for quick lookup of 2FA-enabled users
CREATE INDEX idx_users_2fa_enabled ON users(tenant_id, two_factor_enabled) WHERE two_factor_enabled = true;

-- Comments
COMMENT ON COLUMN users.two_factor_secret IS 'AES-256-GCM encrypted TOTP secret';
COMMENT ON COLUMN users.two_factor_backup_codes IS 'Array of encrypted backup codes (one-time use)';
COMMENT ON COLUMN users.two_factor_recovery_email IS 'Optional recovery email for 2FA reset';
