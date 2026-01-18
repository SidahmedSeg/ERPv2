package services

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"

	"github.com/google/uuid"
	"myerp-v2/internal/config"
)

// EmailService handles sending emails
type EmailService struct {
	config *config.EmailConfig
	app    *config.AppConfig
}

// NewEmailService creates a new email service
func NewEmailService(emailConfig *config.EmailConfig, appConfig *config.AppConfig) *EmailService {
	return &EmailService{
		config: emailConfig,
		app:    appConfig,
	}
}

// SendEmail sends a plain text email
func (s *EmailService) SendEmail(to, subject, body string) error {
	from := s.config.FromEmail

	// Compose message
	message := []byte(fmt.Sprintf(
		"From: %s <%s>\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"Content-Type: text/html; charset=UTF-8\r\n"+
			"\r\n"+
			"%s\r\n",
		s.config.FromName, from, to, subject, body,
	))

	// Connect to SMTP server
	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)

	// For local development (Mailpit), no authentication is needed
	if s.config.SMTPUser == "" && s.config.SMTPPassword == "" {
		client, err := smtp.Dial(addr)
		if err != nil {
			return fmt.Errorf("failed to connect to SMTP server: %w", err)
		}
		defer client.Close()

		if err := client.Mail(from); err != nil {
			return fmt.Errorf("failed to set sender: %w", err)
		}

		if err := client.Rcpt(to); err != nil {
			return fmt.Errorf("failed to set recipient: %w", err)
		}

		w, err := client.Data()
		if err != nil {
			return fmt.Errorf("failed to get data writer: %w", err)
		}

		_, err = w.Write(message)
		if err != nil {
			return fmt.Errorf("failed to write message: %w", err)
		}

		err = w.Close()
		if err != nil {
			return fmt.Errorf("failed to close data writer: %w", err)
		}

		return client.Quit()
	}

	// For production with authentication
	auth := smtp.PlainAuth("", s.config.SMTPUser, s.config.SMTPPassword, s.config.SMTPHost)
	err := smtp.SendMail(addr, auth, from, []string{to}, message)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// SendTenantVerificationEmail sends a verification email for tenant registration
func (s *EmailService) SendTenantVerificationEmail(email, companyName string, token uuid.UUID) error {
	verifyURL := fmt.Sprintf("%s/verify-email?token=%s", s.app.FrontendURL, token)

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #4F46E5; color: white; padding: 20px; text-align: center; }
        .content { padding: 30px 20px; background-color: #f9f9f9; }
        .button { display: inline-block; padding: 12px 30px; background-color: #4F46E5; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { text-align: center; padding: 20px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.AppName}}</h1>
        </div>
        <div class="content">
            <h2>Welcome to {{.AppName}}!</h2>
            <p>Hi {{.CompanyName}},</p>
            <p>Thank you for registering with {{.AppName}}. To complete your registration and activate your account, please verify your email address by clicking the button below:</p>
            <p style="text-align: center;">
                <a href="{{.VerifyURL}}" class="button">Verify Email Address</a>
            </p>
            <p>Or copy and paste this link into your browser:</p>
            <p style="word-break: break-all; color: #4F46E5;">{{.VerifyURL}}</p>
            <p><strong>This link will expire in 24 hours.</strong></p>
            <p>If you didn't create an account with {{.AppName}}, you can safely ignore this email.</p>
        </div>
        <div class="footer">
            <p>&copy; 2026 {{.AppName}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

	data := map[string]string{
		"AppName":     s.app.Name,
		"CompanyName": companyName,
		"VerifyURL":   verifyURL,
	}

	body, err := s.renderTemplate(tmpl, data)
	if err != nil {
		return err
	}

	subject := fmt.Sprintf("Verify your %s account", s.app.Name)
	return s.SendEmail(email, subject, body)
}

// SendPasswordResetEmail sends a password reset email
func (s *EmailService) SendPasswordResetEmail(email, firstName string, token uuid.UUID) error {
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", s.app.FrontendURL, token)

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #4F46E5; color: white; padding: 20px; text-align: center; }
        .content { padding: 30px 20px; background-color: #f9f9f9; }
        .button { display: inline-block; padding: 12px 30px; background-color: #4F46E5; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { text-align: center; padding: 20px; font-size: 12px; color: #666; }
        .warning { background-color: #FEF3C7; padding: 15px; border-left: 4px solid #F59E0B; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.AppName}}</h1>
        </div>
        <div class="content">
            <h2>Password Reset Request</h2>
            <p>Hi {{.FirstName}},</p>
            <p>We received a request to reset your password. Click the button below to create a new password:</p>
            <p style="text-align: center;">
                <a href="{{.ResetURL}}" class="button">Reset Password</a>
            </p>
            <p>Or copy and paste this link into your browser:</p>
            <p style="word-break: break-all; color: #4F46E5;">{{.ResetURL}}</p>
            <p><strong>This link will expire in 1 hour.</strong></p>
            <div class="warning">
                <strong>Security Notice:</strong> If you didn't request a password reset, please ignore this email. Your password will remain unchanged.
            </div>
        </div>
        <div class="footer">
            <p>&copy; 2026 {{.AppName}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

	data := map[string]string{
		"AppName":   s.app.Name,
		"FirstName": firstName,
		"ResetURL":  resetURL,
	}

	body, err := s.renderTemplate(tmpl, data)
	if err != nil {
		return err
	}

	subject := fmt.Sprintf("Reset your %s password", s.app.Name)
	return s.SendEmail(email, subject, body)
}

// SendInvitationEmail sends a team invitation email
func (s *EmailService) SendInvitationEmail(email, companyName, inviterName string, token uuid.UUID, message string) error {
	acceptURL := fmt.Sprintf("%s/accept-invitation?token=%s", s.app.FrontendURL, token)

	customMessage := ""
	if message != "" {
		customMessage = fmt.Sprintf(`<div style="background-color: #EFF6FF; padding: 15px; border-left: 4px solid #3B82F6; margin: 20px 0;">
            <p><strong>Message from %s:</strong></p>
            <p>%s</p>
        </div>`, inviterName, template.HTMLEscapeString(message))
	}

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #4F46E5; color: white; padding: 20px; text-align: center; }
        .content { padding: 30px 20px; background-color: #f9f9f9; }
        .button { display: inline-block; padding: 12px 30px; background-color: #4F46E5; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { text-align: center; padding: 20px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.AppName}}</h1>
        </div>
        <div class="content">
            <h2>You've been invited to join {{.CompanyName}}</h2>
            <p>Hi there,</p>
            <p>{{.InviterName}} has invited you to join their team on {{.AppName}}.</p>
            {{.CustomMessage}}
            <p>Click the button below to accept the invitation and create your account:</p>
            <p style="text-align: center;">
                <a href="{{.AcceptURL}}" class="button">Accept Invitation</a>
            </p>
            <p>Or copy and paste this link into your browser:</p>
            <p style="word-break: break-all; color: #4F46E5;">{{.AcceptURL}}</p>
            <p><strong>This invitation will expire in 7 days.</strong></p>
            <p>If you don't want to join this team, you can safely ignore this email.</p>
        </div>
        <div class="footer">
            <p>&copy; 2026 {{.AppName}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

	data := map[string]interface{}{
		"AppName":       s.app.Name,
		"CompanyName":   companyName,
		"InviterName":   inviterName,
		"AcceptURL":     acceptURL,
		"CustomMessage": template.HTML(customMessage),
	}

	body, err := s.renderTemplateInterface(tmpl, data)
	if err != nil {
		return err
	}

	subject := fmt.Sprintf("You've been invited to join %s on %s", companyName, s.app.Name)
	return s.SendEmail(email, subject, body)
}

// SendWelcomeEmail sends a welcome email after successful registration
func (s *EmailService) SendWelcomeEmail(email, firstName string) error {
	dashboardURL := fmt.Sprintf("%s/dashboard", s.app.FrontendURL)

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #4F46E5; color: white; padding: 20px; text-align: center; }
        .content { padding: 30px 20px; background-color: #f9f9f9; }
        .button { display: inline-block; padding: 12px 30px; background-color: #4F46E5; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { text-align: center; padding: 20px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Welcome to {{.AppName}}!</h1>
        </div>
        <div class="content">
            <h2>Your account is ready</h2>
            <p>Hi {{.FirstName}},</p>
            <p>Your {{.AppName}} account has been successfully activated! You're all set to start using our platform.</p>
            <p style="text-align: center;">
                <a href="{{.DashboardURL}}" class="button">Go to Dashboard</a>
            </p>
            <p>If you have any questions or need assistance, feel free to reach out to our support team.</p>
        </div>
        <div class="footer">
            <p>&copy; 2026 {{.AppName}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

	data := map[string]string{
		"AppName":      s.app.Name,
		"FirstName":    firstName,
		"DashboardURL": dashboardURL,
	}

	body, err := s.renderTemplate(tmpl, data)
	if err != nil {
		return err
	}

	subject := fmt.Sprintf("Welcome to %s!", s.app.Name)
	return s.SendEmail(email, subject, body)
}

// renderTemplate renders an HTML template with string data
func (s *EmailService) renderTemplate(tmplStr string, data map[string]string) (string, error) {
	tmpl, err := template.New("email").Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// renderTemplateInterface renders an HTML template with interface{} data
func (s *EmailService) renderTemplateInterface(tmplStr string, data map[string]interface{}) (string, error) {
	tmpl, err := template.New("email").Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}
