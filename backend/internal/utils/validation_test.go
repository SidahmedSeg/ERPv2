package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		// Valid emails
		{
			name:  "Simple valid email",
			email: "test@example.com",
			want:  true,
		},
		{
			name:  "Email with subdomain",
			email: "user@mail.example.com",
			want:  true,
		},
		{
			name:  "Email with plus",
			email: "user+tag@example.com",
			want:  true,
		},
		{
			name:  "Email with dots",
			email: "first.last@example.com",
			want:  true,
		},
		{
			name:  "Email with numbers",
			email: "user123@example456.com",
			want:  true,
		},
		{
			name:  "Email with hyphens",
			email: "user-name@my-domain.com",
			want:  true,
		},
		{
			name:  "Long domain",
			email: "test@very.long.subdomain.example.com",
			want:  true,
		},

		// Invalid emails
		{
			name:  "Empty email",
			email: "",
			want:  false,
		},
		{
			name:  "Missing @",
			email: "testexample.com",
			want:  false,
		},
		{
			name:  "Missing domain",
			email: "test@",
			want:  false,
		},
		{
			name:  "Missing local part",
			email: "@example.com",
			want:  false,
		},
		{
			name:  "Missing TLD",
			email: "test@example",
			want:  false,
		},
		{
			name:  "Double @",
			email: "test@@example.com",
			want:  false,
		},
		{
			name:  "Spaces",
			email: "test @example.com",
			want:  false,
		},
		{
			name:  "Special chars",
			email: "test!#$@example.com",
			want:  false,
		},
		{
			name:  "Just domain",
			email: "example.com",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateEmail(tt.email)
			assert.Equal(t, tt.want, result, "ValidateEmail(%q) = %v, want %v", tt.email, result, tt.want)
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name       string
		password   string
		wantErrors int // Number of expected validation errors
	}{
		// Valid passwords
		{
			name:       "Strong password",
			password:   "MyP@ssw0rd123",
			wantErrors: 0,
		},
		{
			name:       "Complex password",
			password:   "C0mpl3x!P@ss",
			wantErrors: 0,
		},
		{
			name:       "Long secure password",
			password:   "ThisIsAVerySecureP@ssw0rd123",
			wantErrors: 0,
		},
		{
			name:       "Password with all requirements",
			password:   "Abc123!@",
			wantErrors: 0,
		},

		// Invalid passwords
		{
			name:       "Too short",
			password:   "Abc12!",
			wantErrors: 1, // Length error
		},
		{
			name:       "No uppercase",
			password:   "mypassword123!",
			wantErrors: 1,
		},
		{
			name:       "No lowercase",
			password:   "MYPASSWORD123!",
			wantErrors: 1,
		},
		{
			name:       "No number",
			password:   "MyPassword!",
			wantErrors: 1,
		},
		{
			name:       "No special char",
			password:   "MyPassword123",
			wantErrors: 1,
		},
		{
			name:       "Only lowercase",
			password:   "mypassword",
			wantErrors: 3, // Missing uppercase, number, special
		},
		{
			name:       "Only numbers",
			password:   "12345678",
			wantErrors: 3, // Missing uppercase, lowercase, special
		},
		{
			name:       "Empty password",
			password:   "",
			wantErrors: 4, // All requirements fail
		},
		{
			name:       "Just spaces",
			password:   "        ",
			wantErrors: 3, // Missing upper, lower, number, special (but has length)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidatePassword(tt.password)
			assert.Equal(t, tt.wantErrors, len(errors),
				"ValidatePassword(%q) returned %d errors, want %d. Errors: %v",
				tt.password, len(errors), tt.wantErrors, errors)

			if tt.wantErrors == 0 {
				assert.Empty(t, errors, "Valid password should return no errors")
			} else {
				assert.NotEmpty(t, errors, "Invalid password should return errors")
			}
		})
	}
}

func TestValidateRequired(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{
			name:  "Non-empty string",
			value: "test",
			want:  true,
		},
		{
			name:  "String with spaces",
			value: "  test  ",
			want:  true,
		},
		{
			name:  "Empty string",
			value: "",
			want:  false,
		},
		{
			name:  "Only spaces",
			value: "   ",
			want:  false,
		},
		{
			name:  "Newline only",
			value: "\n",
			want:  false,
		},
		{
			name:  "Tab only",
			value: "\t",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateRequired(tt.value)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestValidateMinLength(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		minLength int
		want      bool
	}{
		{
			name:      "Exact length",
			value:     "12345",
			minLength: 5,
			want:      true,
		},
		{
			name:      "Longer than min",
			value:     "123456",
			minLength: 5,
			want:      true,
		},
		{
			name:      "Shorter than min",
			value:     "1234",
			minLength: 5,
			want:      false,
		},
		{
			name:      "Empty string",
			value:     "",
			minLength: 1,
			want:      false,
		},
		{
			name:      "Zero min length",
			value:     "",
			minLength: 0,
			want:      true,
		},
		{
			name:      "Unicode characters",
			value:     "你好世界",
			minLength: 4,
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateMinLength(tt.value, tt.minLength)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestValidateMaxLength(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		maxLength int
		want      bool
	}{
		{
			name:      "Exact length",
			value:     "12345",
			maxLength: 5,
			want:      true,
		},
		{
			name:      "Shorter than max",
			value:     "1234",
			maxLength: 5,
			want:      true,
		},
		{
			name:      "Longer than max",
			value:     "123456",
			maxLength: 5,
			want:      false,
		},
		{
			name:      "Empty string",
			value:     "",
			maxLength: 5,
			want:      true,
		},
		{
			name:      "Very long string",
			value:     "This is a very long string with many characters",
			maxLength: 10,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateMaxLength(tt.value, tt.maxLength)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestValidateSlug(t *testing.T) {
	tests := []struct {
		name string
		slug string
		want bool
	}{
		// Valid slugs
		{
			name: "Simple slug",
			slug: "my-company",
			want: true,
		},
		{
			name: "Slug with numbers",
			slug: "company-123",
			want: true,
		},
		{
			name: "Short slug",
			slug: "abc",
			want: true,
		},
		{
			name: "Long slug",
			slug: "my-very-long-company-name-slug",
			want: true,
		},
		{
			name: "Only lowercase letters",
			slug: "mycompany",
			want: true,
		},
		{
			name: "Only numbers",
			slug: "123456",
			want: true,
		},

		// Invalid slugs
		{
			name: "Empty slug",
			slug: "",
			want: false,
		},
		{
			name: "Uppercase letters",
			slug: "MyCompany",
			want: false,
		},
		{
			name: "Spaces",
			slug: "my company",
			want: false,
		},
		{
			name: "Special characters",
			slug: "my_company",
			want: false,
		},
		{
			name: "Starts with hyphen",
			slug: "-mycompany",
			want: false,
		},
		{
			name: "Ends with hyphen",
			slug: "mycompany-",
			want: false,
		},
		{
			name: "Double hyphen",
			slug: "my--company",
			want: false,
		},
		{
			name: "Dots",
			slug: "my.company",
			want: false,
		},
		{
			name: "Too long",
			slug: "this-is-a-very-long-slug-that-exceeds-the-maximum-allowed-length-for-slugs",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateSlug(tt.slug)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestValidateUUID(t *testing.T) {
	tests := []struct {
		name string
		uuid string
		want bool
	}{
		{
			name: "Valid UUID v4",
			uuid: "550e8400-e29b-41d4-a716-446655440000",
			want: true,
		},
		{
			name: "Valid UUID v1",
			uuid: "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
			want: true,
		},
		{
			name: "Valid UUID uppercase",
			uuid: "550E8400-E29B-41D4-A716-446655440000",
			want: true,
		},
		{
			name: "Invalid format - missing hyphens",
			uuid: "550e8400e29b41d4a716446655440000",
			want: false,
		},
		{
			name: "Invalid format - wrong positions",
			uuid: "550e8400-e29b41-d4a7-16446655440000",
			want: false,
		},
		{
			name: "Too short",
			uuid: "550e8400-e29b-41d4",
			want: false,
		},
		{
			name: "Too long",
			uuid: "550e8400-e29b-41d4-a716-446655440000-extra",
			want: false,
		},
		{
			name: "Empty string",
			uuid: "",
			want: false,
		},
		{
			name: "Invalid characters",
			uuid: "550e8400-e29b-41d4-g716-446655440000",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateUUID(tt.uuid)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestSanitizeInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Clean input",
			input: "Hello World",
			want:  "Hello World",
		},
		{
			name:  "Trim spaces",
			input: "  Hello World  ",
			want:  "Hello World",
		},
		{
			name:  "Remove control characters",
			input: "Hello\x00World\x01",
			want:  "HelloWorld",
		},
		{
			name:  "Multiple spaces",
			input: "Hello    World",
			want:  "Hello World",
		},
		{
			name:  "Newlines and tabs",
			input: "Hello\n\tWorld",
			want:  "Hello World",
		},
		{
			name:  "Empty string",
			input: "",
			want:  "",
		},
		{
			name:  "Only spaces",
			input: "     ",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeInput(tt.input)
			assert.Equal(t, tt.want, result)
		})
	}
}
