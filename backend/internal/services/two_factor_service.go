package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/redis/go-redis/v9"
	"github.com/skip2/go-qrcode"

	"myerp-v2/internal/config"
	"myerp-v2/internal/database"
	"myerp-v2/internal/utils"
)

const (
	twoFAAttemptKeyPrefix = "2fa_attempts"
	trustedDeviceKeyPrefix = "trusted_device"
	twoFARateLimit = 5
	twoFARateLimitWindow = 15 * time.Minute
	trustedDeviceDuration = 30 * 24 * time.Hour
)

// TwoFactorSetup contains the initial 2FA setup data
type TwoFactorSetup struct {
	Secret      string   `json:"secret"`
	QRCodeURL   string   `json:"qr_code_url"`
	QRCodeImage string   `json:"qr_code_image"` // Base64 encoded PNG
	BackupCodes []string `json:"backup_codes"`
}

// TwoFactorService handles two-factor authentication operations
type TwoFactorService struct {
	db     *sqlx.DB
	redis  *redis.Client
	config *config.Config
}

// NewTwoFactorService creates a new two-factor authentication service
func NewTwoFactorService(db *sqlx.DB, redis *redis.Client, cfg *config.Config) *TwoFactorService {
	return &TwoFactorService{
		db:     db,
		redis:  redis,
		config: cfg,
	}
}

// GenerateSecret generates a new TOTP secret, QR code, and backup codes
func (s *TwoFactorService) GenerateSecret(ctx context.Context, email, accountName string) (*TwoFactorSetup, error) {
	// Generate TOTP secret
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "MyERP v2",
		AccountName: email,
		Period:      30,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP secret: %w", err)
	}

	// Generate QR code image
	qrCodePNG, err := qrcode.Encode(key.URL(), qrcode.Medium, 256)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}

	// Encode QR code to base64
	qrCodeBase64 := base64.StdEncoding.EncodeToString(qrCodePNG)

	// Generate 10 backup codes
	backupCodes, err := s.generateBackupCodes(10)
	if err != nil {
		return nil, fmt.Errorf("failed to generate backup codes: %w", err)
	}

	return &TwoFactorSetup{
		Secret:      key.Secret(),
		QRCodeURL:   key.URL(),
		QRCodeImage: qrCodeBase64,
		BackupCodes: backupCodes,
	}, nil
}

// EnableTwoFactor verifies the initial code and enables 2FA for a user
func (s *TwoFactorService) EnableTwoFactor(ctx context.Context, tenantID, userID uuid.UUID, secret string, verificationCode string, backupCodes []string) error {
	// Verify the initial code
	valid := totp.Validate(verificationCode, secret)
	if !valid {
		return fmt.Errorf("invalid verification code")
	}

	// Encrypt the secret
	encryptedSecret, err := utils.Encrypt(secret, []byte(s.config.Security.EncryptionKey))
	if err != nil {
		return fmt.Errorf("failed to encrypt secret: %w", err)
	}

	// Encrypt backup codes
	encryptedCodes := make([]string, len(backupCodes))
	for i, code := range backupCodes {
		encrypted, err := utils.Encrypt(code, []byte(s.config.Security.EncryptionKey))
		if err != nil {
			return fmt.Errorf("failed to encrypt backup code: %w", err)
		}
		encryptedCodes[i] = encrypted
	}

	// Update user record
	tx, err := database.WithTenantContext(ctx, s.db, tenantID)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		UPDATE users
		SET two_factor_enabled = true,
		    two_factor_secret = $1,
		    two_factor_backup_codes = $2,
		    two_factor_enabled_at = NOW(),
		    updated_at = NOW()
		WHERE id = $3
	`

	_, err = tx.ExecContext(ctx, query, encryptedSecret, pq.Array(encryptedCodes), userID)
	if err != nil {
		return fmt.Errorf("failed to enable 2FA: %w", err)
	}

	return tx.Commit()
}

// DisableTwoFactor disables 2FA for a user (requires password verification first)
func (s *TwoFactorService) DisableTwoFactor(ctx context.Context, tenantID, userID uuid.UUID) error {
	tx, err := database.WithTenantContext(ctx, s.db, tenantID)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		UPDATE users
		SET two_factor_enabled = false,
		    two_factor_secret = NULL,
		    two_factor_backup_codes = NULL,
		    two_factor_enabled_at = NULL,
		    updated_at = NOW()
		WHERE id = $1
	`

	result, err := tx.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to disable 2FA: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	// Clear all trusted devices
	pattern := fmt.Sprintf("%s:%s:%s:*", trustedDeviceKeyPrefix, tenantID, userID)
	iter := s.redis.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		s.redis.Del(ctx, iter.Val())
	}

	return tx.Commit()
}

// VerifyTOTP verifies a TOTP code with time drift tolerance
func (s *TwoFactorService) VerifyTOTP(ctx context.Context, tenantID, userID uuid.UUID, code string) (bool, error) {
	// Rate limit check
	rateLimitKey := database.CacheKey(twoFAAttemptKeyPrefix, tenantID.String(), userID.String())
	attempts, err := s.redis.Incr(ctx, rateLimitKey).Result()
	if err == nil {
		if attempts == 1 {
			s.redis.Expire(ctx, rateLimitKey, twoFARateLimitWindow)
		}
		if attempts > twoFARateLimit {
			return false, fmt.Errorf("too many failed 2FA attempts, please try again in 15 minutes")
		}
	}

	// Get user's encrypted secret
	tx, err := database.WithTenantContext(ctx, s.db, tenantID)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	var encryptedSecret string
	query := `SELECT two_factor_secret FROM users WHERE id = $1 AND two_factor_enabled = true`
	err = tx.QueryRowContext(ctx, query, userID).Scan(&encryptedSecret)
	if err != nil {
		return false, fmt.Errorf("2FA not enabled or user not found")
	}

	// Decrypt secret
	secret, err := utils.Decrypt(encryptedSecret, []byte(s.config.Security.EncryptionKey))
	if err != nil {
		return false, fmt.Errorf("failed to decrypt secret: %w", err)
	}

	// Verify TOTP code with time drift tolerance (±1 period = ±30 seconds)
	valid := totp.Validate(code, secret)

	if valid {
		// Reset rate limit on success
		s.redis.Del(ctx, rateLimitKey)
		return true, tx.Commit()
	}

	return false, tx.Commit()
}

// VerifyBackupCode verifies and consumes a backup code
func (s *TwoFactorService) VerifyBackupCode(ctx context.Context, tenantID, userID uuid.UUID, code string) (bool, error) {
	// Normalize code (remove spaces and dashes)
	code = strings.ReplaceAll(code, " ", "")
	code = strings.ReplaceAll(code, "-", "")
	code = strings.ToUpper(code)

	tx, err := database.WithTenantContext(ctx, s.db, tenantID)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	// Get encrypted backup codes
	var encryptedCodesArray pq.StringArray
	query := `SELECT two_factor_backup_codes FROM users WHERE id = $1 AND two_factor_enabled = true`
	err = tx.QueryRowContext(ctx, query, userID).Scan(&encryptedCodesArray)
	if err != nil {
		return false, fmt.Errorf("2FA not enabled or user not found")
	}

	if len(encryptedCodesArray) == 0 {
		return false, fmt.Errorf("no backup codes available")
	}

	// Decrypt and check codes
	var remainingCodes []string
	found := false

	for _, encrypted := range encryptedCodesArray {
		decrypted, err := utils.Decrypt(encrypted, []byte(s.config.Security.EncryptionKey))
		if err != nil {
			continue
		}

		decrypted = strings.ReplaceAll(decrypted, " ", "")
		decrypted = strings.ReplaceAll(decrypted, "-", "")
		decrypted = strings.ToUpper(decrypted)

		if decrypted == code && !found {
			// Code matches - don't add to remaining codes (consume it)
			found = true
			continue
		}

		remainingCodes = append(remainingCodes, encrypted)
	}

	if !found {
		return false, nil
	}

	// Update backup codes (remove used code)
	encryptedRemaining := make([]string, len(remainingCodes))
	copy(encryptedRemaining, remainingCodes)

	updateQuery := `
		UPDATE users
		SET two_factor_backup_codes = $1,
		    updated_at = NOW()
		WHERE id = $2
	`

	_, err = tx.ExecContext(ctx, updateQuery, pq.Array(encryptedRemaining), userID)
	if err != nil {
		return false, fmt.Errorf("failed to update backup codes: %w", err)
	}

	return true, tx.Commit()
}

// RegenerateBackupCodes generates new backup codes (requires TOTP verification first)
func (s *TwoFactorService) RegenerateBackupCodes(ctx context.Context, tenantID, userID uuid.UUID) ([]string, error) {
	// Generate new backup codes
	backupCodes, err := s.generateBackupCodes(10)
	if err != nil {
		return nil, fmt.Errorf("failed to generate backup codes: %w", err)
	}

	// Encrypt backup codes
	encryptedCodes := make([]string, len(backupCodes))
	for i, code := range backupCodes {
		encrypted, err := utils.Encrypt(code, []byte(s.config.Security.EncryptionKey))
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt backup code: %w", err)
		}
		encryptedCodes[i] = encrypted
	}

	// Update user record
	tx, err := database.WithTenantContext(ctx, s.db, tenantID)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `
		UPDATE users
		SET two_factor_backup_codes = $1,
		    updated_at = NOW()
		WHERE id = $2 AND two_factor_enabled = true
	`

	result, err := tx.ExecContext(ctx, query, pq.Array(encryptedCodes), userID)
	if err != nil {
		return nil, fmt.Errorf("failed to update backup codes: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, fmt.Errorf("2FA not enabled")
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return backupCodes, nil
}

// RememberDevice creates a trusted device token
func (s *TwoFactorService) RememberDevice(ctx context.Context, tenantID, userID uuid.UUID, deviceFingerprint string) (string, error) {
	deviceToken := uuid.New().String()

	// Store device token in Redis with 30-day expiry
	key := database.CacheKey(trustedDeviceKeyPrefix, tenantID.String(), userID.String(), deviceFingerprint)

	deviceInfo := map[string]interface{}{
		"token":      deviceToken,
		"created_at": time.Now().Unix(),
		"tenant_id":  tenantID.String(),
		"user_id":    userID.String(),
	}

	data, err := json.Marshal(deviceInfo)
	if err != nil {
		return "", err
	}

	err = s.redis.Set(ctx, key, data, trustedDeviceDuration).Err()
	if err != nil {
		return "", fmt.Errorf("failed to store trusted device: %w", err)
	}

	return deviceToken, nil
}

// IsDeviceTrusted checks if a device is trusted
func (s *TwoFactorService) IsDeviceTrusted(ctx context.Context, tenantID, userID uuid.UUID, deviceFingerprint, deviceToken string) bool {
	key := database.CacheKey(trustedDeviceKeyPrefix, tenantID.String(), userID.String(), deviceFingerprint)

	data, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		return false
	}

	var deviceInfo map[string]interface{}
	if err := json.Unmarshal([]byte(data), &deviceInfo); err != nil {
		return false
	}

	storedToken, ok := deviceInfo["token"].(string)
	return ok && storedToken == deviceToken
}

// GetRemainingBackupCodes returns the count of remaining backup codes
func (s *TwoFactorService) GetRemainingBackupCodes(ctx context.Context, tenantID, userID uuid.UUID) (int, error) {
	tx, err := database.WithTenantContext(ctx, s.db, tenantID)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var encryptedCodesArray pq.StringArray
	query := `SELECT two_factor_backup_codes FROM users WHERE id = $1 AND two_factor_enabled = true`
	err = tx.QueryRowContext(ctx, query, userID).Scan(&encryptedCodesArray)
	if err != nil {
		return 0, fmt.Errorf("2FA not enabled or user not found")
	}

	return len(encryptedCodesArray), tx.Commit()
}

// generateBackupCodes generates random backup codes
func (s *TwoFactorService) generateBackupCodes(count int) ([]string, error) {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		code, err := s.generateSingleBackupCode()
		if err != nil {
			return nil, err
		}
		codes[i] = code
	}
	return codes, nil
}

// generateSingleBackupCode generates a single 8-character backup code (format: XXXX-XXXX)
func (s *TwoFactorService) generateSingleBackupCode() (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 8)

	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}

	// Format as XXXX-XXXX
	return fmt.Sprintf("%s-%s", string(b[:4]), string(b[4:])), nil
}
