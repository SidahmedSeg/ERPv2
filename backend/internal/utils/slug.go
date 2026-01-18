package utils

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	// Regex to match non-alphanumeric characters (except hyphens)
	nonAlphanumericRegex = regexp.MustCompile(`[^a-z0-9\-]+`)

	// Regex to match multiple consecutive hyphens
	multipleHyphensRegex = regexp.MustCompile(`-{2,}`)
)

// GenerateSlug converts a string into a URL-safe slug
// Example: "My Company Name!" -> "my-company-name"
func GenerateSlug(s string) string {
	// Convert to lowercase
	slug := strings.ToLower(s)

	// Replace spaces and underscores with hyphens
	slug = strings.Map(func(r rune) rune {
		if r == ' ' || r == '_' {
			return '-'
		}
		return r
	}, slug)

	// Remove non-alphanumeric characters (except hyphens)
	slug = nonAlphanumericRegex.ReplaceAllString(slug, "")

	// Replace multiple consecutive hyphens with a single hyphen
	slug = multipleHyphensRegex.ReplaceAllString(slug, "-")

	// Trim hyphens from start and end
	slug = strings.Trim(slug, "-")

	// Ensure slug starts with a letter (required for subdomains)
	if len(slug) > 0 && !unicode.IsLetter(rune(slug[0])) {
		slug = "a-" + slug
	}

	// Limit length to 63 characters (DNS subdomain limit)
	if len(slug) > 63 {
		slug = slug[:63]
		// Trim trailing hyphen if created by truncation
		slug = strings.TrimRight(slug, "-")
	}

	// Ensure minimum length of 3 characters
	if len(slug) < 3 {
		slug = slug + "-org"
	}

	return slug
}

// IsSlugAvailable is a placeholder for checking slug availability
// This would typically query the database to check if a slug is taken
func IsSlugAvailable(slug string) bool {
	// This would be implemented in the repository layer
	// Returning true here as a placeholder
	return true
}

// GenerateUniqueSlug generates a unique slug by appending a number if needed
// Example: "my-company" -> "my-company-2" if "my-company" is taken
func GenerateUniqueSlug(baseSlug string, existingSlugs []string) string {
	slug := baseSlug
	counter := 2

	// Create a map for O(1) lookups
	slugMap := make(map[string]bool)
	for _, s := range existingSlugs {
		slugMap[s] = true
	}

	// Keep incrementing until we find an available slug
	for slugMap[slug] {
		slug = baseSlug + "-" + string(rune(counter+'0'-2))
		counter++
	}

	return slug
}

// SanitizeSlug ensures a slug meets all requirements
func SanitizeSlug(slug string) string {
	return GenerateSlug(slug)
}
