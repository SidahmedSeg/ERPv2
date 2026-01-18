package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"myerp-v2/internal/middleware"
	"myerp-v2/internal/models"
	"myerp-v2/internal/services"
	"myerp-v2/internal/utils"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService *services.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// RegisterTenant handles tenant registration
// POST /api/auth/register
func (h *AuthHandler) RegisterTenant(w http.ResponseWriter, r *http.Request) {
	var req models.TenantCreateRequest
	if err := utils.ParseJSONBody(r, &req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	// Validate request
	errors := utils.ValidationErrors{}
	utils.ValidateRequired("company_name", req.CompanyName, "Company name", &errors)
	utils.ValidateRequired("email", req.Email, "Email", &errors)
	utils.ValidateEmail("email", req.Email, &errors)

	if req.Slug != "" {
		utils.ValidateSlug("slug", req.Slug, &errors)
	}

	if errors.HasErrors() {
		utils.UnprocessableEntity(w, "Validation failed", errors.ToMap())
		return
	}

	// Register tenant
	tenant, err := h.authService.RegisterTenant(r.Context(), &req)
	if err != nil {
		utils.BadRequest(w, err.Error())
		return
	}

	utils.Created(w, map[string]interface{}{
		"tenant": tenant,
		"message": "Registration successful. Please check your email to verify your account.",
	})
}

// VerifyEmail handles email verification
// POST /api/auth/verify-email
func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}

	if err := utils.ParseJSONBody(r, &req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	// Parse token
	token, err := uuid.Parse(req.Token)
	if err != nil {
		utils.BadRequest(w, "Invalid verification token")
		return
	}

	// Verify email
	tenant, err := h.authService.VerifyTenantEmail(r.Context(), token)
	if err != nil {
		utils.BadRequest(w, err.Error())
		return
	}

	utils.Success(w, map[string]interface{}{
		"tenant":  tenant,
		"message": "Email verified successfully. Your account is now active!",
	})
}

// Login handles user login
// POST /api/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.UserLoginRequest
	if err := utils.ParseJSONBody(r, &req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	// Validate request
	errors := utils.ValidationErrors{}
	utils.ValidateRequired("email", req.Email, "Email", &errors)
	utils.ValidateEmail("email", req.Email, &errors)
	utils.ValidateRequired("password", req.Password, "Password", &errors)

	if errors.HasErrors() {
		utils.UnprocessableEntity(w, "Validation failed", errors.ToMap())
		return
	}

	// Extract tenant ID from context (set by tenant resolution middleware)
	tenantID, err := middleware.GetTenantIDFromContext(r.Context())
	if err != nil {
		utils.BadRequest(w, "Tenant context required")
		return
	}

	// Get device info
	deviceInfo := utils.ParseDeviceInfo(r)
	ipAddress := utils.GetClientIP(r)

	// Login
	response, err := h.authService.LoginWithTenant(r.Context(), tenantID, &req, deviceInfo, ipAddress)
	if err != nil {
		utils.Unauthorized(w, err.Error())
		return
	}

	utils.Success(w, response)
}

// Verify2FA handles 2FA verification
// POST /api/auth/verify-2fa
func (h *AuthHandler) Verify2FA(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TwoFactorToken string `json:"two_factor_token"`
		Code           string `json:"code"`
		RememberMe     bool   `json:"remember_me"`
	}

	if err := utils.ParseJSONBody(r, &req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	// Validate
	if req.TwoFactorToken == "" || req.Code == "" {
		utils.BadRequest(w, "Two-factor token and code are required")
		return
	}

	// Get device info
	deviceInfo := utils.ParseDeviceInfo(r)
	ipAddress := utils.GetClientIP(r)

	// Verify 2FA
	response, err := h.authService.Verify2FAAndLogin(
		r.Context(), req.TwoFactorToken, req.Code, deviceInfo, ipAddress, req.RememberMe,
	)
	if err != nil {
		utils.Unauthorized(w, err.Error())
		return
	}

	utils.Success(w, response)
}

// Logout handles user logout
// POST /api/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Get tenant ID and access token from context
	tenantID, err := middleware.GetTenantIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	accessToken, ok := r.Context().Value("access_token").(string)
	if !ok {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	// Hash token
	// TODO: Extract this to a utility function
	// For now, the auth service will handle it

	// Logout
	if err := h.authService.Logout(r.Context(), tenantID, accessToken); err != nil {
		utils.InternalServerError(w, "Failed to logout")
		return
	}

	utils.Success(w, map[string]interface{}{
		"message": "Logged out successfully",
	})
}

// LogoutAll handles logging out from all devices
// POST /api/auth/logout-all
func (h *AuthHandler) LogoutAll(w http.ResponseWriter, r *http.Request) {
	// Get user and tenant from context
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	tenantID, err := middleware.GetTenantIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	// Logout from all devices
	if err := h.authService.LogoutAll(r.Context(), tenantID, userID); err != nil {
		utils.InternalServerError(w, "Failed to logout from all devices")
		return
	}

	utils.Success(w, map[string]interface{}{
		"message": "Logged out from all devices successfully",
	})
}

// RefreshToken handles token refresh
// POST /api/auth/refresh
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := utils.ParseJSONBody(r, &req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	if req.RefreshToken == "" {
		utils.BadRequest(w, "Refresh token is required")
		return
	}

	// Refresh token
	response, err := h.authService.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		utils.Unauthorized(w, err.Error())
		return
	}

	utils.Success(w, response)
}

// RequestPasswordReset handles password reset request
// POST /api/auth/forgot-password
func (h *AuthHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req models.UserPasswordResetRequest
	if err := utils.ParseJSONBody(r, &req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	// Validate email
	errors := utils.ValidationErrors{}
	utils.ValidateEmail("email", req.Email, &errors)

	if errors.HasErrors() {
		utils.UnprocessableEntity(w, "Validation failed", errors.ToMap())
		return
	}

	// Extract tenant ID from context
	tenantID, ok := r.Context().Value("tenant_id").(uuid.UUID)
	if !ok {
		utils.BadRequest(w, "Tenant context required")
		return
	}

	// Request password reset
	if err := h.authService.RequestPasswordReset(r.Context(), tenantID, req.Email); err != nil {
		// Don't reveal error details for security
		utils.Success(w, map[string]interface{}{
			"message": "If the email exists, a password reset link has been sent.",
		})
		return
	}

	utils.Success(w, map[string]interface{}{
		"message": "If the email exists, a password reset link has been sent.",
	})
}

// ResetPassword handles password reset with token
// POST /api/auth/reset-password
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req models.UserPasswordResetConfirm
	if err := utils.ParseJSONBody(r, &req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	// Validate
	errors := utils.ValidationErrors{}
	utils.ValidateRequired("token", req.Token, "Token", &errors)
	utils.ValidatePassword("new_password", req.NewPassword, &errors)

	if errors.HasErrors() {
		utils.UnprocessableEntity(w, "Validation failed", errors.ToMap())
		return
	}

	// Parse token
	token, err := uuid.Parse(req.Token)
	if err != nil {
		utils.BadRequest(w, "Invalid reset token")
		return
	}

	// Extract tenant ID from context
	tenantID, ok := r.Context().Value("tenant_id").(uuid.UUID)
	if !ok {
		utils.BadRequest(w, "Tenant context required")
		return
	}

	// Reset password
	if err := h.authService.ResetPassword(r.Context(), tenantID, token, req.NewPassword); err != nil {
		utils.BadRequest(w, err.Error())
		return
	}

	utils.Success(w, map[string]interface{}{
		"message": "Password reset successfully. You can now login with your new password.",
	})
}

// ChangePassword handles password change (requires authentication)
// POST /api/auth/change-password
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var req models.UserChangePasswordRequest
	if err := utils.ParseJSONBody(r, &req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	// Validate
	errors := utils.ValidationErrors{}
	utils.ValidateRequired("current_password", req.CurrentPassword, "Current password", &errors)
	utils.ValidatePassword("new_password", req.NewPassword, &errors)

	if errors.HasErrors() {
		utils.UnprocessableEntity(w, "Validation failed", errors.ToMap())
		return
	}

	// Get user and tenant from context
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	tenantID, err := middleware.GetTenantIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	// Change password
	if err := h.authService.ChangePassword(r.Context(), tenantID, userID, req.CurrentPassword, req.NewPassword); err != nil {
		utils.BadRequest(w, err.Error())
		return
	}

	utils.Success(w, map[string]interface{}{
		"message": "Password changed successfully",
	})
}

// GetCurrentUser returns the currently authenticated user
// GET /api/auth/me
func (h *AuthHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	utils.Success(w, map[string]interface{}{
		"user": user,
	})
}

// RegisterRoutes registers all auth routes
func (h *AuthHandler) RegisterRoutes(r chi.Router, authMiddleware *middleware.AuthMiddleware, tenantMiddleware *middleware.TenantMiddleware) {
	// Public routes
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", h.RegisterTenant)
		r.Post("/verify-email", h.VerifyEmail)

		// Login requires tenant context
		r.With(tenantMiddleware.RequireTenant).Post("/login", h.Login)
		r.With(tenantMiddleware.RequireTenant).Post("/verify-2fa", h.Verify2FA)

		r.Post("/refresh", h.RefreshToken)

		// Password reset requires tenant context
		r.With(tenantMiddleware.RequireTenant).Post("/forgot-password", h.RequestPasswordReset)
		r.With(tenantMiddleware.RequireTenant).Post("/reset-password", h.ResetPassword)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)
			r.Get("/me", h.GetCurrentUser)
			r.Post("/logout", h.Logout)
			r.Post("/logout-all", h.LogoutAll)
			r.Post("/change-password", h.ChangePassword)
		})
	})
}
