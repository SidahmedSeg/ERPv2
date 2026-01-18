package utils

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// Email validation regex (RFC 5322 simplified)
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// ValidationError represents a field validation error
type ValidationError struct {
	Field   string
	Message string
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []ValidationError

// Add adds a new validation error
func (ve *ValidationErrors) Add(field, message string) {
	*ve = append(*ve, ValidationError{
		Field:   field,
		Message: message,
	})
}

// HasErrors returns true if there are any validation errors
func (ve ValidationErrors) HasErrors() bool {
	return len(ve) > 0
}

// ToMap converts validation errors to a map for JSON response
func (ve ValidationErrors) ToMap() map[string]string {
	result := make(map[string]string)
	for _, err := range ve {
		result[err.Field] = err.Message
	}
	return result
}

// IsValidEmail validates an email address
func IsValidEmail(email string) bool {
	if len(email) < 3 || len(email) > 254 {
		return false
	}
	return emailRegex.MatchString(email)
}

// IsValidPassword validates a password based on security requirements
func IsValidPassword(password string) (bool, string) {
	if len(password) < 8 {
		return false, "Password must be at least 8 characters long"
	}

	if len(password) > 128 {
		return false, "Password must not exceed 128 characters"
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return false, "Password must contain at least one uppercase letter"
	}
	if !hasLower {
		return false, "Password must contain at least one lowercase letter"
	}
	if !hasNumber {
		return false, "Password must contain at least one number"
	}
	if !hasSpecial {
		return false, "Password must contain at least one special character"
	}

	return true, ""
}

// IsValidPhone validates a phone number (basic validation)
func IsValidPhone(phone string) bool {
	if phone == "" {
		return true // Optional field
	}

	// Remove common separators
	cleaned := strings.Map(func(r rune) rune {
		if r == ' ' || r == '-' || r == '(' || r == ')' || r == '+' {
			return -1
		}
		return r
	}, phone)

	// Check if it's all digits and reasonable length (7-15 digits)
	if len(cleaned) < 7 || len(cleaned) > 15 {
		return false
	}

	for _, char := range cleaned {
		if !unicode.IsDigit(char) {
			return false
		}
	}

	return true
}

// IsValidName validates a person's name
func IsValidName(name string) bool {
	trimmed := strings.TrimSpace(name)
	if len(trimmed) < 1 || len(trimmed) > 100 {
		return false
	}

	// Name should only contain letters, spaces, hyphens, and apostrophes
	for _, char := range trimmed {
		if !unicode.IsLetter(char) && char != ' ' && char != '-' && char != '\'' {
			return false
		}
	}

	return true
}

// IsValidSlug validates a URL-safe slug
func IsValidSlug(slug string) bool {
	if len(slug) < 3 || len(slug) > 63 {
		return false
	}

	// Slug must start with a letter
	if !unicode.IsLetter(rune(slug[0])) {
		return false
	}

	// Slug can only contain lowercase letters, numbers, and hyphens
	for _, char := range slug {
		if !unicode.IsLower(char) && !unicode.IsDigit(char) && char != '-' {
			return false
		}
	}

	// Slug cannot start or end with a hyphen
	if slug[0] == '-' || slug[len(slug)-1] == '-' {
		return false
	}

	// Slug cannot contain consecutive hyphens
	if strings.Contains(slug, "--") {
		return false
	}

	return true
}

// SanitizeString removes leading/trailing whitespace and limits length
func SanitizeString(s string, maxLength int) string {
	trimmed := strings.TrimSpace(s)
	if len(trimmed) > maxLength {
		return trimmed[:maxLength]
	}
	return trimmed
}

// ValidateRequired checks if a required field is present and non-empty
func ValidateRequired(field, value, fieldName string, errors *ValidationErrors) {
	if strings.TrimSpace(value) == "" {
		errors.Add(field, fmt.Sprintf("%s is required", fieldName))
	}
}

// ValidateEmail validates an email field
func ValidateEmail(field, email string, errors *ValidationErrors) {
	if !IsValidEmail(email) {
		errors.Add(field, "Invalid email address")
	}
}

// ValidatePassword validates a password field
func ValidatePassword(field, password string, errors *ValidationErrors) {
	if valid, msg := IsValidPassword(password); !valid {
		errors.Add(field, msg)
	}
}

// ValidateName validates a name field
func ValidateName(field, name, fieldName string, errors *ValidationErrors) {
	if !IsValidName(name) {
		errors.Add(field, fmt.Sprintf("%s must be 1-100 characters and contain only letters, spaces, hyphens, and apostrophes", fieldName))
	}
}

// ValidateSlug validates a slug field
func ValidateSlug(field, slug string, errors *ValidationErrors) {
	if !IsValidSlug(slug) {
		errors.Add(field, "Slug must be 3-63 characters, start with a letter, and contain only lowercase letters, numbers, and hyphens")
	}
}

// ValidateStringLength validates string length
func ValidateStringLength(field, value string, min, max int, fieldName string, errors *ValidationErrors) {
	length := len(strings.TrimSpace(value))
	if length < min || length > max {
		errors.Add(field, fmt.Sprintf("%s must be between %d and %d characters", fieldName, min, max))
	}
}

// ValidateEnum validates that a value is one of the allowed options
func ValidateEnum(field, value string, allowed []string, fieldName string, errors *ValidationErrors) {
	for _, option := range allowed {
		if value == option {
			return
		}
	}
	errors.Add(field, fmt.Sprintf("%s must be one of: %s", fieldName, strings.Join(allowed, ", ")))
}
