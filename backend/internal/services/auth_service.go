package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"myerp-v2/internal/config"
	"myerp-v2/internal/models"
	"myerp-v2/internal/repository"
	"myerp-v2/internal/utils"
)

// AuthService handles authentication operations
type AuthService struct {
	tenantRepo   *repository.TenantRepository
	userRepo     *repository.UserRepository
	sessionRepo  *repository.SessionRepository
	roleRepo     *repository.RoleRepository
	userRoleRepo *repository.UserRoleRepository
	jwtService   *JWTService
	emailService *EmailService
	config       *config.Config
}

// NewAuthService creates a new auth service
func NewAuthService(
	tenantRepo *repository.TenantRepository,
	userRepo *repository.UserRepository,
	sessionRepo *repository.SessionRepository,
	roleRepo *repository.RoleRepository,
	userRoleRepo *repository.UserRoleRepository,
	jwtService *JWTService,
	emailService *EmailService,
	cfg *config.Config,
) *AuthService {
	return &AuthService{
		tenantRepo:   tenantRepo,
		userRepo:     userRepo,
		sessionRepo:  sessionRepo,
		roleRepo:     roleRepo,
		userRoleRepo: userRoleRepo,
		jwtService:   jwtService,
		emailService: emailService,
		config:       cfg,
	}
}

// RegisterTenant registers a new tenant with email verification
func (s *AuthService) RegisterTenant(ctx context.Context, req *models.TenantCreateRequest) (*models.Tenant, error) {
	// Check if email is already registered
	exists, err := s.tenantRepo.CheckEmailExists(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("email already registered")
	}

	// Generate slug from company name if not provided
	slug := req.Slug
	if slug == "" {
		slug = utils.GenerateSlug(req.CompanyName)
	}

	// Check if slug is available
	available, err := s.tenantRepo.CheckSlugAvailability(ctx, slug)
	if err != nil {
		return nil, err
	}
	if !available {
		// Generate a unique slug by appending a number
		baseSlug := slug
		counter := 2
		for !available {
			slug = fmt.Sprintf("%s-%d", baseSlug, counter)
			available, err = s.tenantRepo.CheckSlugAvailability(ctx, slug)
			if err != nil {
				return nil, err
			}
			counter++
		}
	}

	// Create verification token
	verificationToken := uuid.New()
	expiresAt := time.Now().Add(s.config.Security.VerificationExpiry)

	tenant := &models.Tenant{
		Slug:                       slug,
		CompanyName:                req.CompanyName,
		Email:                      req.Email,
		Status:                     models.TenantStatusPendingVerification,
		VerificationToken:          &verificationToken,
		VerificationTokenExpiresAt: &expiresAt,
		PlanTier:                   models.PlanTierFree,
		Settings:                   []byte("{}"),
	}

	// Create tenant
	if err := s.tenantRepo.Create(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	// Hash password for initial admin user
	hashedPassword, err := utils.HashPassword(req.Password, s.config.Security.BcryptCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create initial admin user with pending status
	user := &models.User{
		TenantID:     tenant.ID,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Status:       "pending", // Will be activated when email is verified
		Timezone:     "UTC",
		Language:     "en",
		Preferences:  []byte("{}"),
	}

	if err := s.userRepo.Create(ctx, tenant.ID, user); err != nil {
		// If user creation fails, we should probably rollback tenant creation
		// For now, log error and continue (tenant can re-register or admin can fix)
		fmt.Printf("Failed to create initial admin user: %v\n", err)
		return nil, fmt.Errorf("failed to create initial admin user: %w", err)
	}

	// Send verification email
	if err := s.emailService.SendTenantVerificationEmail(tenant.Email, tenant.CompanyName, verificationToken); err != nil {
		// Log error but don't fail registration
		// In production, you might want to retry sending or queue it
		fmt.Printf("Failed to send verification email: %v\n", err)
	}

	return tenant, nil
}

// VerifyTenantEmail verifies a tenant's email and provisions system roles
func (s *AuthService) VerifyTenantEmail(ctx context.Context, token uuid.UUID) (*models.Tenant, error) {
	// Find tenant by verification token
	tenant, err := s.tenantRepo.FindByVerificationToken(ctx, token)
	if err != nil {
		return nil, err
	}

	// Verify email
	if err := s.tenantRepo.VerifyEmail(ctx, tenant.ID); err != nil {
		return nil, fmt.Errorf("failed to verify email: %w", err)
	}

	// Provision system roles
	if err := s.tenantRepo.ProvisionSystemRoles(ctx, tenant.ID); err != nil {
		return nil, fmt.Errorf("failed to provision system roles: %w", err)
	}

	// Activate the initial admin user
	user, err := s.userRepo.FindByEmail(ctx, tenant.ID, tenant.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to find initial admin user: %w", err)
	}

	// Update user status to active
	if err := s.userRepo.UpdateStatus(ctx, tenant.ID, user.ID, "active"); err != nil {
		return nil, fmt.Errorf("failed to activate user: %w", err)
	}

	// Mark user's email as verified
	if err := s.userRepo.VerifyEmail(ctx, tenant.ID, user.ID); err != nil {
		return nil, fmt.Errorf("failed to verify user email: %w", err)
	}

	// Assign owner role to the initial admin user
	// First, find the owner role for this tenant
	ownerRole, err := s.roleRepo.FindByName(ctx, tenant.ID, "owner")
	if err != nil {
		// If owner role doesn't exist, log warning but continue
		fmt.Printf("Warning: Owner role not found for tenant %s: %v\n", tenant.ID, err)
	} else {
		// Assign owner role to user (assigned_by = user.ID since they're the first user)
		if err := s.userRoleRepo.AssignRole(ctx, tenant.ID, user.ID, ownerRole.ID, user.ID); err != nil {
			return nil, fmt.Errorf("failed to assign owner role: %w", err)
		}
	}

	// Update tenant status
	tenant.Status = models.TenantStatusActive
	tenant.EmailVerified = true
	now := time.Now()
	tenant.EmailVerifiedAt = &now
	tenant.ActivatedAt = &now

	return tenant, nil
}

// Login authenticates a user and creates a session
func (s *AuthService) Login(ctx context.Context, req *models.UserLoginRequest, deviceInfo utils.DeviceInfo, ipAddress string) (*models.UserLoginResponse, error) {
	// Find all users with this email across all tenants
	users, err := s.userRepo.FindAllByEmail(ctx, req.Email)
	if err != nil || len(users) == 0 {
		return nil, fmt.Errorf("invalid email or password")
	}

	// If tenant_id is specified (second step of multi-tenant login)
	if req.TenantID != "" {
		tenantUUID, err := uuid.Parse(req.TenantID)
		if err != nil {
			return nil, fmt.Errorf("invalid tenant ID")
		}

		// Find the specific user in that tenant
		for _, user := range users {
			if user.TenantID == tenantUUID {
				// Verify password
				if !utils.VerifyPassword(req.Password, user.PasswordHash) {
					return nil, fmt.Errorf("invalid email or password")
				}

				// Check user status
				if !user.CanLogin() {
					return nil, fmt.Errorf("user account is not active")
				}

				// Authenticate this specific user
				return s.LoginWithTenant(ctx, user.TenantID, req, deviceInfo, ipAddress)
			}
		}
		return nil, fmt.Errorf("invalid tenant selection")
	}

	// First login attempt - verify password with first user found
	// (all users with same email should have same password)
	if !utils.VerifyPassword(req.Password, users[0].PasswordHash) {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Check how many tenants user belongs to
	if len(users) > 1 {
		// User belongs to multiple tenants - return tenant list for selection
		tenants := make([]*models.Tenant, len(users))
		for i, user := range users {
			tenant, err := s.tenantRepo.FindByID(ctx, user.TenantID)
			if err != nil {
				continue // Skip this tenant if error
			}
			tenants[i] = tenant
		}

		return &models.UserLoginResponse{
			Tenants: tenants, // Frontend will show tenant selector
		}, nil
	}

	// Single tenant - authenticate directly
	user := &users[0]

	// Check user status
	if !user.CanLogin() {
		return nil, fmt.Errorf("user account is not active")
	}

	return s.LoginWithTenant(ctx, user.TenantID, req, deviceInfo, ipAddress)
}

// LoginWithTenant authenticates a user within a specific tenant
func (s *AuthService) LoginWithTenant(ctx context.Context, tenantID uuid.UUID, req *models.UserLoginRequest, deviceInfo utils.DeviceInfo, ipAddress string) (*models.UserLoginResponse, error) {
	// Find tenant
	tenant, err := s.tenantRepo.FindByID(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("tenant not found")
	}

	// Check tenant status
	if !tenant.CanAccess() {
		return nil, fmt.Errorf("tenant account is not active")
	}

	// Find user by email
	user, err := s.userRepo.FindByEmail(ctx, tenantID, req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Verify password
	if !utils.VerifyPassword(req.Password, user.PasswordHash) {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Check user status
	if !user.CanLogin() {
		return nil, fmt.Errorf("user account is not active")
	}

	// Check if 2FA is enabled
	if user.TwoFactorEnabled {
		// Generate temporary 2FA token
		twoFactorToken, err := s.jwtService.Generate2FAToken(user.ID, tenantID, tenant.Slug, user.Email)
		if err != nil {
			return nil, fmt.Errorf("failed to generate 2FA token: %w", err)
		}

		return &models.UserLoginResponse{
			User:              user,
			RequiresTwoFactor: true,
			TwoFactorToken:    twoFactorToken,
		}, nil
	}

	// Create session and generate tokens
	return s.createSessionAndTokens(ctx, user, tenant, req.RememberMe, deviceInfo, ipAddress)
}

// Verify2FAAndLogin verifies 2FA code and completes login
func (s *AuthService) Verify2FAAndLogin(ctx context.Context, twoFactorToken, code string, deviceInfo utils.DeviceInfo, ipAddress string, rememberMe bool) (*models.UserLoginResponse, error) {
	// Validate 2FA token
	claims, err := s.jwtService.Validate2FAToken(twoFactorToken)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired 2FA token")
	}

	// Find user
	_, err = s.userRepo.FindByID(ctx, claims.TenantID, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Verify 2FA code
	// TODO: Implement TOTP verification using user.TwoFactorSecret
	// For now, return error
	return nil, fmt.Errorf("2FA verification not yet implemented")
}

// Logout logs out a user by deleting their session
func (s *AuthService) Logout(ctx context.Context, tenantID uuid.UUID, tokenHash string) error {
	return s.sessionRepo.DeleteByTokenHash(ctx, tenantID, tokenHash)
}

// LogoutAll logs out a user from all devices
func (s *AuthService) LogoutAll(ctx context.Context, tenantID, userID uuid.UUID) error {
	return s.sessionRepo.DeleteAllByUser(ctx, tenantID, userID)
}

// RefreshToken generates new access token from refresh token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*models.UserLoginResponse, error) {
	// Validate refresh token
	claims, err := s.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Find user
	user, err := s.userRepo.FindByID(ctx, claims.TenantID, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Check if user is still active
	if !user.CanLogin() {
		return nil, fmt.Errorf("user account is not active")
	}

	// Generate new access token
	accessToken, expiresIn, err := s.jwtService.GenerateAccessToken(
		user.ID, claims.TenantID, claims.TenantSlug, user.Email, false,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	return &models.UserLoginResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken, // Reuse existing refresh token
		ExpiresIn:    expiresIn,
	}, nil
}

// RequestPasswordReset initiates password reset flow
func (s *AuthService) RequestPasswordReset(ctx context.Context, tenantID uuid.UUID, email string) error {
	// Find user
	user, err := s.userRepo.FindByEmail(ctx, tenantID, email)
	if err != nil {
		// Don't reveal if email exists or not for security
		return nil
	}

	// Generate reset token
	resetToken := uuid.New()
	expiresAt := time.Now().Add(s.config.Security.PasswordResetExpiry).Format(time.RFC3339)

	// Save reset token
	if err := s.userRepo.SetResetToken(ctx, tenantID, user.ID, resetToken, expiresAt); err != nil {
		return fmt.Errorf("failed to set reset token: %w", err)
	}

	// Send reset email
	if err := s.emailService.SendPasswordResetEmail(user.Email, user.FirstName, resetToken); err != nil {
		fmt.Printf("Failed to send password reset email: %v\n", err)
	}

	return nil
}

// ResetPassword resets a user's password using reset token
func (s *AuthService) ResetPassword(ctx context.Context, tenantID uuid.UUID, token uuid.UUID, newPassword string) error {
	// Validate password
	if valid, msg := utils.IsValidPassword(newPassword); !valid {
		return fmt.Errorf(msg)
	}

	// Find user by reset token
	user, err := s.userRepo.FindByResetToken(ctx, tenantID, token)
	if err != nil {
		return fmt.Errorf("invalid or expired reset token")
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(newPassword, s.config.Security.BcryptCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	if err := s.userRepo.UpdatePassword(ctx, tenantID, user.ID, hashedPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Revoke all existing sessions for security
	s.sessionRepo.DeleteAllByUser(ctx, tenantID, user.ID)

	return nil
}

// ChangePassword changes a user's password (requires current password)
func (s *AuthService) ChangePassword(ctx context.Context, tenantID, userID uuid.UUID, currentPassword, newPassword string) error {
	// Validate new password
	if valid, msg := utils.IsValidPassword(newPassword); !valid {
		return fmt.Errorf(msg)
	}

	// Find user
	user, err := s.userRepo.FindByID(ctx, tenantID, userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Verify current password
	if !utils.VerifyPassword(currentPassword, user.PasswordHash) {
		return fmt.Errorf("current password is incorrect")
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(newPassword, s.config.Security.BcryptCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	if err := s.userRepo.UpdatePassword(ctx, tenantID, userID, hashedPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// createSessionAndTokens creates a session and generates JWT tokens
func (s *AuthService) createSessionAndTokens(
	ctx context.Context,
	user *models.User,
	tenant *models.Tenant,
	rememberMe bool,
	deviceInfo utils.DeviceInfo,
	ipAddress string,
) (*models.UserLoginResponse, error) {
	// Generate tokens
	accessToken, expiresIn, err := s.jwtService.GenerateAccessToken(
		user.ID, tenant.ID, tenant.Slug, user.Email, rememberMe,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(
		user.ID, tenant.ID, tenant.Slug, user.Email,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Hash access token for storage
	hash := sha256.Sum256([]byte(accessToken))
	tokenHash := hex.EncodeToString(hash[:])

	// Determine session expiry
	var sessionExpiry time.Duration
	if rememberMe {
		sessionExpiry = s.config.JWT.RememberMeExpiry
	} else {
		sessionExpiry = s.config.JWT.AccessTokenExpiry
	}

	// Create session
	sessionReq := &models.SessionCreateRequest{
		UserID:     user.ID,
		TenantID:   tenant.ID,
		Token:      tokenHash,
		DeviceType: deviceInfo.DeviceType,
		Browser:    deviceInfo.Browser,
		OS:         deviceInfo.OS,
		IPAddress:  ipAddress,
		UserAgent:  deviceInfo.UserAgent,
		ExpiresAt:  time.Now().Add(sessionExpiry),
	}

	_, err = s.sessionRepo.Create(ctx, sessionReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Update last login
	if err := s.userRepo.UpdateLastLogin(ctx, tenant.ID, user.ID, ipAddress); err != nil {
		// Log error but don't fail login
		fmt.Printf("Failed to update last login: %v\n", err)
	}

	return &models.UserLoginResponse{
		User:         user,
		Tenant:       tenant,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

// ValidateSession validates a session token and returns the user
func (s *AuthService) ValidateSession(ctx context.Context, accessToken string) (*models.User, *models.Tenant, error) {
	// Validate access token
	claims, err := s.jwtService.ValidateAccessToken(accessToken)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid token: %w", err)
	}

	// Hash token to find session
	hash := sha256.Sum256([]byte(accessToken))
	tokenHash := hex.EncodeToString(hash[:])

	// Find session
	session, err := s.sessionRepo.FindByTokenHash(ctx, claims.TenantID, tokenHash)
	if err != nil {
		return nil, nil, fmt.Errorf("session not found or expired")
	}

	// Update session activity
	if err := s.sessionRepo.UpdateActivity(ctx, claims.TenantID, session.ID); err != nil {
		// Log error but don't fail validation
		fmt.Printf("Failed to update session activity: %v\n", err)
	}

	// Find user
	user, err := s.userRepo.FindByID(ctx, claims.TenantID, claims.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("user not found")
	}

	// Find tenant
	tenant, err := s.tenantRepo.FindByID(ctx, claims.TenantID)
	if err != nil {
		return nil, nil, fmt.Errorf("tenant not found")
	}

	return user, tenant, nil
}
