package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/mileusna/useragent"
)

// DeviceInfo holds information about a user's device
type DeviceInfo struct {
	DeviceType string // Desktop | Mobile | Tablet
	Browser    string // Chrome, Firefox, Safari, etc.
	OS         string // Windows, macOS, Linux, iOS, Android, etc.
	UserAgent  string // Full user agent string
}

// ParseDeviceInfo extracts device information from an HTTP request
func ParseDeviceInfo(r *http.Request) DeviceInfo {
	userAgentString := r.UserAgent()
	ua := useragent.Parse(userAgentString)

	deviceType := "Desktop"
	if ua.Mobile {
		deviceType = "Mobile"
	} else if ua.Tablet {
		deviceType = "Tablet"
	}

	browser := ua.Name
	if browser == "" {
		browser = "Unknown"
	}

	os := ua.OS
	if os == "" {
		os = "Unknown"
	}

	return DeviceInfo{
		DeviceType: deviceType,
		Browser:    browser,
		OS:         os,
		UserAgent:  userAgentString,
	}
}

// GenerateDeviceFingerprint creates a hash-based fingerprint from device info
// This is used for "remember this device" functionality
func GenerateDeviceFingerprint(r *http.Request) string {
	components := []string{
		r.UserAgent(),
		r.Header.Get("Accept-Language"),
		// Note: Don't include IP address as it may change (mobile networks, VPNs)
	}

	// Combine all components
	combined := strings.Join(components, "|")

	// Generate SHA-256 hash
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:])
}

// GetClientIP extracts the client's IP address from the request
// Checks X-Forwarded-For, X-Real-IP headers first (for proxies/load balancers)
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (may contain multiple IPs)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Take the first IP (client's original IP)
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	// RemoteAddr is in the format "IP:port" or "[IPv6]:port"
	// We need to strip the port correctly for both IPv4 and IPv6
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// If SplitHostPort fails, return the original (might be just an IP without port)
		return r.RemoteAddr
	}

	return host
}

// DeviceString returns a human-readable device description
func (d DeviceInfo) DeviceString() string {
	return fmt.Sprintf("%s on %s (%s)", d.Browser, d.OS, d.DeviceType)
}

// IsBot checks if the user agent appears to be a bot/crawler
func IsBot(userAgent string) bool {
	ua := strings.ToLower(userAgent)
	bots := []string{
		"bot", "crawler", "spider", "scraper",
		"googlebot", "bingbot", "yahoo", "baiduspider",
		"yandex", "duckduckbot", "slurp", "facebookexternalhit",
	}

	for _, bot := range bots {
		if strings.Contains(ua, bot) {
			return true
		}
	}

	return false
}

// IsMobile checks if the request is from a mobile device
func IsMobile(r *http.Request) bool {
	ua := useragent.Parse(r.UserAgent())
	return ua.Mobile
}

// IsTablet checks if the request is from a tablet device
func IsTablet(r *http.Request) bool {
	ua := useragent.Parse(r.UserAgent())
	return ua.Tablet
}

// IsDesktop checks if the request is from a desktop device
func IsDesktop(r *http.Request) bool {
	ua := useragent.Parse(r.UserAgent())
	return !ua.Mobile && !ua.Tablet
}
