package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"myerp-v2/internal/middleware"
	"myerp-v2/internal/repository"
	"myerp-v2/internal/services"
	"myerp-v2/internal/utils"
)

// TwoFactorHandler handles two-factor authentication endpoints
type TwoFactorHandler struct {
	twoFactorService *services.TwoFactorService
	userRepo         *repository.UserRepository
}

// NewTwoFactorHandler creates a new two-factor authentication handler
func NewTwoFactorHandler(
	twoFactorService *services.TwoFactorService,
	userRepo *repository.UserRepository,
) *TwoFactorHandler {
	return &TwoFactorHandler{
		twoFactorService: twoFactorService,
		userRepo:         userRepo,
	}
}

// Setup generates a new TOTP secret and QR code for 2FA setup
// POST /api/2fa/setup
func (h *TwoFactorHandler) Setup(w http.ResponseWriter, r *http.Request) {
	tenantID, err := middleware.GetTenantIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	// Get user email
	user, err := h.userRepo.FindByID(r.Context(), tenantID, userID)
	if err != nil {
		utils.NotFound(w, "User not found")
		return
	}

	// Check if 2FA is already enabled
	if user.TwoFactorEnabled {
		utils.BadRequest(w, "2FA is already enabled")
		return
	}

	// Generate secret and QR code
	fullName := fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	setup, err := h.twoFactorService.GenerateSecret(r.Context(), user.Email, fullName)
	if err != nil {
		utils.InternalServerError(w, "Failed to generate 2FA setup")
		return
	}

	utils.Success(w, map[string]interface{}{
		"secret":        setup.Secret,
		"qr_code_url":   setup.QRCodeURL,
		"qr_code_image": setup.QRCodeImage,
		"backup_codes":  setup.BackupCodes,
		"message":       "Scan the QR code with your authenticator app and verify with a code to enable 2FA",
	})
}

// Enable enables 2FA after verifying the initial code
// POST /api/2fa/enable
func (h *TwoFactorHandler) Enable(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Secret          string   `json:"secret"`
		VerificationCode string   `json:"verification_code"`
		BackupCodes     []string `json:"backup_codes"`
	}

	if err := utils.ParseJSONBody(r, &req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	// Validate
	errors := utils.ValidationErrors{}
	utils.ValidateRequired("secret", req.Secret, "Secret", &errors)
	utils.ValidateRequired("verification_code", req.VerificationCode, "Verification code", &errors)

	if len(req.BackupCodes) == 0 {
		errors.Add("backup_codes", "Backup codes are required")
	}

	if errors.HasErrors() {
		utils.UnprocessableEntity(w, "Validation failed", errors.ToMap())
		return
	}

	tenantID, err := middleware.GetTenantIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	// Enable 2FA
	err = h.twoFactorService.EnableTwoFactor(
		r.Context(),
		tenantID,
		userID,
		req.Secret,
		req.VerificationCode,
		req.BackupCodes,
	)
	if err != nil {
		if err.Error() == "invalid verification code" {
			utils.BadRequest(w, "Invalid verification code")
			return
		}
		utils.InternalServerError(w, "Failed to enable 2FA")
		return
	}

	utils.Success(w, map[string]interface{}{
		"message": "Two-factor authentication enabled successfully",
	})
}

// Disable disables 2FA (requires password verification)
// POST /api/2fa/disable
func (h *TwoFactorHandler) Disable(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Password string `json:"password"`
	}

	if err := utils.ParseJSONBody(r, &req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	// Validate
	errors := utils.ValidationErrors{}
	utils.ValidateRequired("password", req.Password, "Password", &errors)

	if errors.HasErrors() {
		utils.UnprocessableEntity(w, "Validation failed", errors.ToMap())
		return
	}

	tenantID, err := middleware.GetTenantIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	// Get user and verify password
	user, err := h.userRepo.FindByID(r.Context(), tenantID, userID)
	if err != nil {
		utils.NotFound(w, "User not found")
		return
	}

	if !utils.VerifyPassword(req.Password, user.PasswordHash) {
		utils.BadRequest(w, "Invalid password")
		return
	}

	// Disable 2FA
	err = h.twoFactorService.DisableTwoFactor(r.Context(), tenantID, userID)
	if err != nil {
		utils.InternalServerError(w, "Failed to disable 2FA")
		return
	}

	utils.Success(w, map[string]interface{}{
		"message": "Two-factor authentication disabled successfully",
	})
}

// Verify verifies a TOTP code (used during login)
// POST /api/2fa/verify
func (h *TwoFactorHandler) Verify(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code string `json:"code"`
	}

	if err := utils.ParseJSONBody(r, &req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	// Validate
	errors := utils.ValidationErrors{}
	utils.ValidateRequired("code", req.Code, "Verification code", &errors)

	if errors.HasErrors() {
		utils.UnprocessableEntity(w, "Validation failed", errors.ToMap())
		return
	}

	tenantID, err := middleware.GetTenantIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	// Verify TOTP code
	valid, err := h.twoFactorService.VerifyTOTP(r.Context(), tenantID, userID, req.Code)
	if err != nil {
		utils.BadRequest(w, err.Error())
		return
	}

	if !valid {
		utils.BadRequest(w, "Invalid verification code")
		return
	}

	utils.Success(w, map[string]interface{}{
		"valid":   true,
		"message": "Verification successful",
	})
}

// VerifyBackupCode verifies a backup code (alternative to TOTP)
// POST /api/2fa/verify-backup
func (h *TwoFactorHandler) VerifyBackupCode(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code string `json:"code"`
	}

	if err := utils.ParseJSONBody(r, &req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	// Validate
	errors := utils.ValidationErrors{}
	utils.ValidateRequired("code", req.Code, "Backup code", &errors)

	if errors.HasErrors() {
		utils.UnprocessableEntity(w, "Validation failed", errors.ToMap())
		return
	}

	tenantID, err := middleware.GetTenantIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	// Verify backup code
	valid, err := h.twoFactorService.VerifyBackupCode(r.Context(), tenantID, userID, req.Code)
	if err != nil {
		utils.BadRequest(w, err.Error())
		return
	}

	if !valid {
		utils.BadRequest(w, "Invalid backup code")
		return
	}

	// Get remaining backup codes count
	remaining, _ := h.twoFactorService.GetRemainingBackupCodes(r.Context(), tenantID, userID)

	utils.Success(w, map[string]interface{}{
		"valid":                true,
		"message":              "Backup code verified successfully",
		"remaining_backup_codes": remaining,
	})
}

// RegenerateBackupCodes generates new backup codes
// POST /api/2fa/backup-codes/regenerate
func (h *TwoFactorHandler) RegenerateBackupCodes(w http.ResponseWriter, r *http.Request) {
	var req struct {
		VerificationCode string `json:"verification_code"`
	}

	if err := utils.ParseJSONBody(r, &req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	// Validate
	errors := utils.ValidationErrors{}
	utils.ValidateRequired("verification_code", req.VerificationCode, "Verification code", &errors)

	if errors.HasErrors() {
		utils.UnprocessableEntity(w, "Validation failed", errors.ToMap())
		return
	}

	tenantID, err := middleware.GetTenantIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	// Verify TOTP code first
	valid, err := h.twoFactorService.VerifyTOTP(r.Context(), tenantID, userID, req.VerificationCode)
	if err != nil || !valid {
		utils.BadRequest(w, "Invalid verification code")
		return
	}

	// Regenerate backup codes
	backupCodes, err := h.twoFactorService.RegenerateBackupCodes(r.Context(), tenantID, userID)
	if err != nil {
		utils.InternalServerError(w, "Failed to regenerate backup codes")
		return
	}

	utils.Success(w, map[string]interface{}{
		"backup_codes": backupCodes,
		"message":      "Backup codes regenerated successfully. Save these codes in a safe place.",
	})
}

// GetBackupCodesCount returns the count of remaining backup codes
// GET /api/2fa/backup-codes/count
func (h *TwoFactorHandler) GetBackupCodesCount(w http.ResponseWriter, r *http.Request) {
	tenantID, err := middleware.GetTenantIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	count, err := h.twoFactorService.GetRemainingBackupCodes(r.Context(), tenantID, userID)
	if err != nil {
		utils.InternalServerError(w, "Failed to get backup codes count")
		return
	}

	utils.Success(w, map[string]interface{}{
		"count": count,
	})
}

// TrustDevice marks the current device as trusted (skip 2FA for 30 days)
// POST /api/2fa/device/trust
func (h *TwoFactorHandler) TrustDevice(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DeviceFingerprint string `json:"device_fingerprint"`
	}

	if err := utils.ParseJSONBody(r, &req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	// Validate
	errors := utils.ValidationErrors{}
	utils.ValidateRequired("device_fingerprint", req.DeviceFingerprint, "Device fingerprint", &errors)

	if errors.HasErrors() {
		utils.UnprocessableEntity(w, "Validation failed", errors.ToMap())
		return
	}

	tenantID, err := middleware.GetTenantIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	// Remember device
	deviceToken, err := h.twoFactorService.RememberDevice(r.Context(), tenantID, userID, req.DeviceFingerprint)
	if err != nil {
		utils.InternalServerError(w, "Failed to trust device")
		return
	}

	utils.Success(w, map[string]interface{}{
		"device_token": deviceToken,
		"message":      "Device trusted for 30 days. You won't need 2FA on this device.",
	})
}

// RegisterRoutes registers all two-factor authentication routes
func (h *TwoFactorHandler) RegisterRoutes(r chi.Router, authMiddleware *middleware.AuthMiddleware) {
	r.Route("/2fa", func(r chi.Router) {
		// All 2FA routes require authentication
		r.Use(authMiddleware.Authenticate)

		// Setup 2FA (get QR code and secret)
		r.Post("/setup", h.Setup)

		// Enable 2FA with verification
		r.Post("/enable", h.Enable)

		// Disable 2FA (requires password)
		r.Post("/disable", h.Disable)

		// Verify TOTP code
		r.Post("/verify", h.Verify)

		// Verify backup code
		r.Post("/verify-backup", h.VerifyBackupCode)

		// Backup codes management
		r.Post("/backup-codes/regenerate", h.RegenerateBackupCodes)
		r.Get("/backup-codes/count", h.GetBackupCodesCount)

		// Device trust
		r.Post("/device/trust", h.TrustDevice)
	})
}
