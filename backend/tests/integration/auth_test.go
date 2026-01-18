// +build integration

package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"myerp-v2/internal/config"
	"myerp-v2/internal/server"
)

// TestAuthFlow tests the complete authentication flow
func TestAuthFlow(t *testing.T) {
	// Setup test server
	cfg := loadTestConfig()
	router := server.NewRouter(cfg)
	srv := httptest.NewServer(router.Setup())
	defer srv.Close()

	// Test data
	tenantSlug := "test-company-" + randomString(8)
	email := "test@example.com"
	password := "Test@Password123"

	t.Run("Register new tenant", func(t *testing.T) {
		payload := map[string]interface{}{
			"company_name": "Test Company",
			"slug":         tenantSlug,
			"email":        email,
			"first_name":   "John",
			"last_name":    "Doe",
			"password":     password,
		}

		resp, err := makeRequest(srv.URL+"/api/auth/register", "POST", payload, "")
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		assert.Equal(t, "success", result["status"])
		assert.NotNil(t, result["data"])
	})

	t.Run("Login with credentials", func(t *testing.T) {
		payload := map[string]interface{}{
			"email":       email,
			"password":    password,
			"tenant_slug": tenantSlug,
		}

		resp, err := makeRequest(srv.URL+"/api/auth/login", "POST", payload, "")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		data := result["data"].(map[string]interface{})
		assert.NotEmpty(t, data["access_token"])
		assert.NotEmpty(t, data["refresh_token"])
		assert.NotNil(t, data["user"])
		assert.NotNil(t, data["tenant"])
	})

	t.Run("Login with wrong password", func(t *testing.T) {
		payload := map[string]interface{}{
			"email":       email,
			"password":    "WrongPassword123",
			"tenant_slug": tenantSlug,
		}

		resp, err := makeRequest(srv.URL+"/api/auth/login", "POST", payload, "")
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Access protected endpoint without token", func(t *testing.T) {
		resp, err := makeRequest(srv.URL+"/api/auth/me", "GET", nil, "")
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Access protected endpoint with token", func(t *testing.T) {
		// First login to get token
		payload := map[string]interface{}{
			"email":       email,
			"password":    password,
			"tenant_slug": tenantSlug,
		}

		loginResp, _ := makeRequest(srv.URL+"/api/auth/login", "POST", payload, "")
		var loginResult map[string]interface{}
		json.NewDecoder(loginResp.Body).Decode(&loginResult)
		data := loginResult["data"].(map[string]interface{})
		token := data["access_token"].(string)

		// Access protected endpoint
		resp, err := makeRequest(srv.URL+"/api/auth/me", "GET", nil, token)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		userData := result["data"].(map[string]interface{})
		assert.Equal(t, email, userData["email"])
	})
}

// TestUserManagement tests user CRUD operations
func TestUserManagement(t *testing.T) {
	// Setup
	cfg := loadTestConfig()
	router := server.NewRouter(cfg)
	srv := httptest.NewServer(router.Setup())
	defer srv.Close()

	// Create tenant and login
	token := createTenantAndLogin(t, srv.URL)

	t.Run("List users", func(t *testing.T) {
		resp, err := makeRequest(srv.URL+"/api/users?page=1&page_size=10", "GET", nil, token)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		data := result["data"].(map[string]interface{})
		assert.NotNil(t, data["users"])
		assert.NotNil(t, result["meta"])
	})

	t.Run("Search users", func(t *testing.T) {
		resp, err := makeRequest(srv.URL+"/api/users/search?query=john&page=1", "GET", nil, token)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

// TestRoleManagement tests role and permission operations
func TestRoleManagement(t *testing.T) {
	cfg := loadTestConfig()
	router := server.NewRouter(cfg)
	srv := httptest.NewServer(router.Setup())
	defer srv.Close()

	token := createTenantAndLogin(t, srv.URL)

	t.Run("List roles", func(t *testing.T) {
		resp, err := makeRequest(srv.URL+"/api/roles", "GET", nil, token)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		data := result["data"].(map[string]interface{})
		assert.NotNil(t, data["roles"])
	})

	t.Run("List permissions", func(t *testing.T) {
		resp, err := makeRequest(srv.URL+"/api/permissions", "GET", nil, token)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		data := result["data"].(map[string]interface{})
		assert.NotNil(t, data["permissions"])
	})

	t.Run("Create custom role", func(t *testing.T) {
		// First get permission IDs
		permResp, _ := makeRequest(srv.URL+"/api/permissions", "GET", nil, token)
		var permResult map[string]interface{}
		json.NewDecoder(permResp.Body).Decode(&permResult)
		permData := permResult["data"].(map[string]interface{})
		permissions := permData["permissions"].([]interface{})
		permissionID := permissions[0].(map[string]interface{})["id"].(string)

		// Create role
		payload := map[string]interface{}{
			"name":           "test-role",
			"display_name":   "Test Role",
			"description":    "A test role",
			"permission_ids": []string{permissionID},
		}

		resp, err := makeRequest(srv.URL+"/api/roles", "POST", payload, token)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})
}

// TestSecurityFeatures tests 2FA, sessions, and audit logs
func TestSecurityFeatures(t *testing.T) {
	cfg := loadTestConfig()
	router := server.NewRouter(cfg)
	srv := httptest.NewServer(router.Setup())
	defer srv.Close()

	token := createTenantAndLogin(t, srv.URL)

	t.Run("Setup 2FA", func(t *testing.T) {
		resp, err := makeRequest(srv.URL+"/api/2fa/setup", "POST", nil, token)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		data := result["data"].(map[string]interface{})
		assert.NotEmpty(t, data["qr_code"])
		assert.NotEmpty(t, data["secret"])
		assert.NotNil(t, data["backup_codes"])
	})

	t.Run("List active sessions", func(t *testing.T) {
		resp, err := makeRequest(srv.URL+"/api/sessions", "GET", nil, token)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		data := result["data"].(map[string]interface{})
		assert.NotNil(t, data["sessions"])
	})

	t.Run("Query audit logs", func(t *testing.T) {
		resp, err := makeRequest(srv.URL+"/api/audit?page=1&page_size=10", "GET", nil, token)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		data := result["data"].(map[string]interface{})
		assert.NotNil(t, data["logs"])
	})
}

// Helper functions

func loadTestConfig() *config.Config {
	// Load test configuration
	// In practice, this would load from .env.test
	return &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "myerp",
			Password: "test_password",
			DBName:   "myerp_v2_test",
		},
		Redis: config.RedisConfig{
			Host:     "localhost",
			Port:     6379,
			Password: "",
		},
		JWT: config.JWTConfig{
			Secret: "test-secret-key-minimum-32-characters-long",
		},
		Server: config.ServerConfig{
			Port:        8080,
			Environment: "test",
		},
	}
}

func makeRequest(url, method string, payload interface{}, token string) (*http.Response, error) {
	var body *bytes.Buffer
	if payload != nil {
		jsonData, _ := json.Marshal(payload)
		body = bytes.NewBuffer(jsonData)
	} else {
		body = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{}
	return client.Do(req)
}

func createTenantAndLogin(t *testing.T, baseURL string) string {
	// Register
	tenantSlug := "test-" + randomString(8)
	registerPayload := map[string]interface{}{
		"company_name": "Test Company",
		"slug":         tenantSlug,
		"email":        "admin@test.com",
		"first_name":   "Admin",
		"last_name":    "User",
		"password":     "Admin@123456",
	}

	makeRequest(baseURL+"/api/auth/register", "POST", registerPayload, "")

	// Login
	loginPayload := map[string]interface{}{
		"email":       "admin@test.com",
		"password":    "Admin@123456",
		"tenant_slug": tenantSlug,
	}

	resp, _ := makeRequest(baseURL+"/api/auth/login", "POST", loginPayload, "")
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	data := result["data"].(map[string]interface{})

	return data["access_token"].(string)
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[i%len(charset)]
	}
	return string(b)
}
