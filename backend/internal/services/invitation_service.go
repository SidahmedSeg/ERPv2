package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"myerp-v2/internal/database"
	"myerp-v2/internal/models"
	"myerp-v2/internal/repository"
)

const (
	invitationExpiryDuration = 7 * 24 * time.Hour // 7 days
)

// InvitationService handles team invitation operations
type InvitationService struct {
	db           *sqlx.DB
	userRepo     *repository.UserRepository
	userRoleRepo *repository.UserRoleRepository
	emailService *EmailService
}

// NewInvitationService creates a new invitation service
func NewInvitationService(
	db *sqlx.DB,
	userRepo *repository.UserRepository,
	userRoleRepo *repository.UserRoleRepository,
	emailService *EmailService,
) *InvitationService {
	return &InvitationService{
		db:           db,
		userRepo:     userRepo,
		userRoleRepo: userRoleRepo,
		emailService: emailService,
	}
}

// CreateInvitation creates a new team invitation
func (s *InvitationService) CreateInvitation(
	ctx context.Context,
	tenantID, invitedBy uuid.UUID,
	email string,
	roleIDs []uuid.UUID,
	message string,
) (*models.Invitation, error) {
	tx, err := database.WithTenantContext(ctx, s.db, tenantID)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Check if user already exists
	existingUser, err := s.userRepo.FindByEmail(ctx, tenantID, email)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("user with this email already exists")
	}

	// Check if there's already a pending invitation
	var existingInvitation models.Invitation
	checkQuery := `
		SELECT id, email, status
		FROM invitations
		WHERE email = $1 AND status = 'pending'
		LIMIT 1
	`
	err = tx.GetContext(ctx, &existingInvitation, checkQuery, email)
	if err == nil {
		// Pending invitation exists - revoke it first
		revokeQuery := `
			UPDATE invitations
			SET status = 'revoked'
			WHERE id = $1
		`
		_, _ = tx.ExecContext(ctx, revokeQuery, existingInvitation.ID)
	}

	// Create new invitation
	invitation := &models.Invitation{
		Email:     email,
		Token:     uuid.New(),
		RoleIDs:   roleIDs,
		Status:    models.InvitationStatusPending,
		Message:   &message,
		InvitedBy: invitedBy,
		ExpiresAt: time.Now().Add(invitationExpiryDuration),
	}

	query := `
		INSERT INTO invitations (
			tenant_id, email, token, role_ids, status, message, invited_by, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, invited_at
	`

	err = tx.QueryRowContext(ctx, query,
		tenantID, invitation.Email, invitation.Token, pq.Array(invitation.RoleIDs),
		invitation.Status, invitation.Message, invitation.InvitedBy, invitation.ExpiresAt,
	).Scan(&invitation.ID, &invitation.InvitedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Send invitation email (async - don't fail if email fails)
	go func() {
		ctx := context.Background()

		// Get tenant company name
		var companyName string
		_ = s.db.QueryRowContext(ctx, "SELECT company_name FROM tenants WHERE id = $1", tenantID).Scan(&companyName)
		if companyName == "" {
			companyName = "MyERP"
		}

		// Get inviter info
		inviter, _ := s.userRepo.FindByID(ctx, tenantID, invitedBy)
		inviterName := "Team member"
		if inviter != nil {
			inviterName = fmt.Sprintf("%s %s", inviter.FirstName, inviter.LastName)
		}

		_ = s.emailService.SendInvitationEmail(
			email,
			companyName,
			inviterName,
			invitation.Token,
			message,
		)
	}()

	return invitation, nil
}

// AcceptInvitation accepts an invitation and creates the user account
func (s *InvitationService) AcceptInvitation(
	ctx context.Context,
	token, password, firstName, lastName string,
) (*models.User, error) {
	// Find invitation (bypass RLS - we don't have tenant context yet)
	tx, err := database.WithBypassRLS(ctx, s.db)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var invitation models.Invitation
	query := `
		SELECT
			id, tenant_id, email, token, role_ids, status,
			message, invited_by, invited_at, accepted_at, expires_at
		FROM invitations
		WHERE token = $1
		LIMIT 1
	`

	err = tx.GetContext(ctx, &invitation, query, token)
	if err != nil {
		return nil, fmt.Errorf("invitation not found")
	}

	// Validate invitation status
	if invitation.Status != models.InvitationStatusPending {
		return nil, fmt.Errorf("invitation has already been %s", invitation.Status)
	}

	// Check expiry
	if time.Now().After(invitation.ExpiresAt) {
		// Mark as expired
		updateQuery := `UPDATE invitations SET status = 'expired' WHERE id = $1`
		_, _ = tx.ExecContext(ctx, updateQuery, invitation.ID)
		return nil, fmt.Errorf("invitation has expired")
	}

	// Create user account (switch to tenant context)
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Hash password
	passwordHash, err := hashPassword(password, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		Email:        invitation.Email,
		PasswordHash: passwordHash,
		FirstName:    firstName,
		LastName:     lastName,
		Status:       models.UserStatusActive,
		Timezone:     "UTC",
		Language:     "en",
		Preferences:  []byte("{}"),
	}

	err = s.userRepo.Create(ctx, invitation.TenantID, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Assign roles
	err = s.userRoleRepo.AssignRoles(ctx, invitation.TenantID, user.ID, invitation.RoleIDs, invitation.InvitedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to assign roles: %w", err)
	}

	// Mark invitation as accepted
	txAccept, err := database.WithBypassRLS(ctx, s.db)
	if err != nil {
		return nil, err
	}
	defer txAccept.Rollback()

	updateQuery := `
		UPDATE invitations
		SET status = 'accepted',
		    accepted_at = NOW()
		WHERE id = $1
	`
	_, err = txAccept.ExecContext(ctx, updateQuery, invitation.ID)
	if err != nil {
		return nil, err
	}

	if err := txAccept.Commit(); err != nil {
		return nil, err
	}

	// Send welcome email (async)
	go func() {
		_ = s.emailService.SendWelcomeEmail(user.Email, user.FirstName)
	}()

	return user, nil
}

// RevokeInvitation revokes a pending invitation
func (s *InvitationService) RevokeInvitation(ctx context.Context, tenantID, invitationID uuid.UUID) error {
	tx, err := database.WithTenantContext(ctx, s.db, tenantID)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		UPDATE invitations
		SET status = 'revoked'
		WHERE id = $1 AND status = 'pending'
	`

	result, err := tx.ExecContext(ctx, query, invitationID)
	if err != nil {
		return fmt.Errorf("failed to revoke invitation: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("invitation not found or already processed")
	}

	return tx.Commit()
}

// ListInvitations lists all invitations for a tenant
func (s *InvitationService) ListInvitations(ctx context.Context, tenantID uuid.UUID, status string, limit, offset int) ([]models.Invitation, int, error) {
	tx, err := database.WithTenantContext(ctx, s.db, tenantID)
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback()

	// Build query
	baseQuery := `
		FROM invitations
		WHERE 1=1
	`
	var args []interface{}
	argIndex := 1

	if status != "" {
		baseQuery += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	// Get total count
	var totalCount int
	countQuery := "SELECT COUNT(*) " + baseQuery
	err = tx.GetContext(ctx, &totalCount, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// Get invitations
	selectQuery := `
		SELECT
			id, tenant_id, email, token, role_ids, status,
			message, invited_by, invited_at, accepted_at, expires_at
	` + baseQuery + ` ORDER BY invited_at DESC LIMIT $` + fmt.Sprintf("%d", argIndex) + ` OFFSET $` + fmt.Sprintf("%d", argIndex+1)

	args = append(args, limit, offset)

	var invitations []models.Invitation
	err = tx.SelectContext(ctx, &invitations, selectQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	return invitations, totalCount, tx.Commit()
}

// GetInvitation retrieves a single invitation
func (s *InvitationService) GetInvitation(ctx context.Context, tenantID, invitationID uuid.UUID) (*models.Invitation, error) {
	tx, err := database.WithTenantContext(ctx, s.db, tenantID)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var invitation models.Invitation
	query := `
		SELECT
			id, tenant_id, email, token, role_ids, status,
			message, invited_by, invited_at, accepted_at, expires_at
		FROM invitations
		WHERE id = $1
		LIMIT 1
	`

	err = tx.GetContext(ctx, &invitation, query, invitationID)
	if err != nil {
		return nil, fmt.Errorf("invitation not found")
	}

	return &invitation, tx.Commit()
}

// ResendInvitation resends an invitation email
func (s *InvitationService) ResendInvitation(ctx context.Context, tenantID, invitationID, resendBy uuid.UUID) error {
	invitation, err := s.GetInvitation(ctx, tenantID, invitationID)
	if err != nil {
		return err
	}

	if invitation.Status != models.InvitationStatusPending {
		return fmt.Errorf("can only resend pending invitations")
	}

	if time.Now().After(invitation.ExpiresAt) {
		return fmt.Errorf("invitation has expired")
	}

	// Get tenant company name
	var companyName string
	_ = s.db.QueryRowContext(ctx, "SELECT company_name FROM tenants WHERE id = $1", tenantID).Scan(&companyName)
	if companyName == "" {
		companyName = "MyERP"
	}

	// Get inviter info
	inviter, _ := s.userRepo.FindByID(ctx, tenantID, resendBy)
	inviterName := "Team member"
	if inviter != nil {
		inviterName = fmt.Sprintf("%s %s", inviter.FirstName, inviter.LastName)
	}

	message := ""
	if invitation.Message != nil {
		message = *invitation.Message
	}

	// Send invitation email
	return s.emailService.SendInvitationEmail(
		invitation.Email,
		companyName,
		inviterName,
		invitation.Token,
		message,
	)
}

// CleanupExpiredInvitations marks expired invitations as expired (should be run periodically)
func (s *InvitationService) CleanupExpiredInvitations(ctx context.Context) (int, error) {
	// Use bypass RLS for cleanup task (system operation)
	tx, err := database.WithBypassRLS(ctx, s.db)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	query := `
		UPDATE invitations
		SET status = 'expired'
		WHERE status = 'pending'
		  AND expires_at < NOW()
	`

	result, err := tx.ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired invitations: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()

	return int(rowsAffected), tx.Commit()
}

// Helper function to hash password (duplicate of utils.HashPassword for now)
func hashPassword(password string, cost int) (string, error) {
	// This will use utils.HashPassword when available
	// For now, placeholder
	return "", fmt.Errorf("not implemented - use utils.HashPassword")
}
