package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Email    EmailConfig
	Security SecurityConfig
	App      AppConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host            string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	Environment     string // development | staging | production
}

// DatabaseConfig holds PostgreSQL configuration
type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// JWTConfig holds JWT authentication configuration
type JWTConfig struct {
	Secret              string
	RefreshSecret       string
	AccessTokenExpiry   time.Duration
	RefreshTokenExpiry  time.Duration
	Issuer              string
	RememberMeExpiry    time.Duration // Extended expiry for "remember me"
	TrustedDeviceExpiry time.Duration // How long to remember trusted devices
}

// EmailConfig holds SMTP email configuration
type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	FromEmail    string
	FromName     string
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	EncryptionKey          string // AES-256 key for encrypting sensitive data (2FA secrets, etc.)
	BcryptCost             int    // bcrypt cost factor (10-12 recommended)
	PasswordResetExpiry    time.Duration
	VerificationExpiry     time.Duration
	InvitationExpiry       time.Duration
	MaxLoginAttempts       int
	LoginRateLimitWindow   time.Duration
	Max2FAAttempts         int
	TwoFARateLimitWindow   time.Duration
	SessionInactivityLimit time.Duration
}

// AppConfig holds general application configuration
type AppConfig struct {
	Name            string
	BaseURL         string // Base URL for the application (used in emails, etc.)
	FrontendURL     string // Frontend URL for CORS and redirects
	LogLevel        string // debug | info | warn | error
	EnableSwagger   bool   // Enable Swagger API documentation
	EnableProfiling bool   // Enable pprof profiling endpoints
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if exists (ignore errors in production)
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Host:            getEnv("SERVER_HOST", "0.0.0.0"),
			Port:            getEnvAsInt("SERVER_PORT", 8080),
			ReadTimeout:     getEnvAsDuration("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout:    getEnvAsDuration("SERVER_WRITE_TIMEOUT", 15*time.Second),
			ShutdownTimeout: getEnvAsDuration("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second),
			Environment:     getEnv("ENVIRONMENT", "development"),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvAsInt("DB_PORT", 5432),
			User:            getEnv("DB_USER", "myerp"),
			Password:        getEnv("DB_PASSWORD", "myerp_password"),
			Database:        getEnv("DB_NAME", "myerp_v2"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 100),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 25),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 1*time.Hour),
			ConnMaxIdleTime: getEnvAsDuration("DB_CONN_MAX_IDLE_TIME", 10*time.Minute),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", "redis_password"),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret:              getEnv("JWT_SECRET", "your-jwt-secret-key-change-in-production"),
			RefreshSecret:       getEnv("JWT_REFRESH_SECRET", "your-jwt-refresh-secret-key-change-in-production"),
			AccessTokenExpiry:   getEnvAsDuration("JWT_ACCESS_EXPIRY", 15*time.Minute),
			RefreshTokenExpiry:  getEnvAsDuration("JWT_REFRESH_EXPIRY", 7*24*time.Hour),
			Issuer:              getEnv("JWT_ISSUER", "myerp-v2"),
			RememberMeExpiry:    getEnvAsDuration("JWT_REMEMBER_ME_EXPIRY", 30*24*time.Hour),
			TrustedDeviceExpiry: getEnvAsDuration("TRUSTED_DEVICE_EXPIRY", 30*24*time.Hour),
		},
		Email: EmailConfig{
			SMTPHost:     getEnv("SMTP_HOST", "localhost"),
			SMTPPort:     getEnvAsInt("SMTP_PORT", 1025),
			SMTPUser:     getEnv("SMTP_USER", ""),
			SMTPPassword: getEnv("SMTP_PASSWORD", ""),
			FromEmail:    getEnv("EMAIL_FROM", "noreply@myerp.local"),
			FromName:     getEnv("EMAIL_FROM_NAME", "MyERP v2"),
		},
		Security: SecurityConfig{
			EncryptionKey:          getEnv("ENCRYPTION_KEY", "change-this-to-a-32-byte-key!!"),
			BcryptCost:             getEnvAsInt("BCRYPT_COST", 10),
			PasswordResetExpiry:    getEnvAsDuration("PASSWORD_RESET_EXPIRY", 1*time.Hour),
			VerificationExpiry:     getEnvAsDuration("VERIFICATION_EXPIRY", 24*time.Hour),
			InvitationExpiry:       getEnvAsDuration("INVITATION_EXPIRY", 7*24*time.Hour),
			MaxLoginAttempts:       getEnvAsInt("MAX_LOGIN_ATTEMPTS", 5),
			LoginRateLimitWindow:   getEnvAsDuration("LOGIN_RATE_LIMIT_WINDOW", 5*time.Minute),
			Max2FAAttempts:         getEnvAsInt("MAX_2FA_ATTEMPTS", 5),
			TwoFARateLimitWindow:   getEnvAsDuration("2FA_RATE_LIMIT_WINDOW", 15*time.Minute),
			SessionInactivityLimit: getEnvAsDuration("SESSION_INACTIVITY_LIMIT", 30*time.Minute),
		},
		App: AppConfig{
			Name:            getEnv("APP_NAME", "MyERP v2"),
			BaseURL:         getEnv("APP_BASE_URL", "http://localhost:8080"),
			FrontendURL:     getEnv("FRONTEND_URL", "http://localhost:3000"),
			LogLevel:        getEnv("LOG_LEVEL", "info"),
			EnableSwagger:   getEnvAsBool("ENABLE_SWAGGER", true),
			EnableProfiling: getEnvAsBool("ENABLE_PROFILING", false),
		},
	}

	// Validate critical configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// Validate checks that critical configuration values are properly set
func (c *Config) Validate() error {
	// Validate JWT secrets in production
	if c.Server.Environment == "production" {
		if c.JWT.Secret == "your-jwt-secret-key-change-in-production" {
			return fmt.Errorf("JWT_SECRET must be changed in production")
		}
		if c.JWT.RefreshSecret == "your-jwt-refresh-secret-key-change-in-production" {
			return fmt.Errorf("JWT_REFRESH_SECRET must be changed in production")
		}
		if len(c.Security.EncryptionKey) != 32 {
			return fmt.Errorf("ENCRYPTION_KEY must be exactly 32 bytes for AES-256")
		}
	}

	// Validate database connection
	if c.Database.Host == "" {
		return fmt.Errorf("DB_HOST is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("DB_USER is required")
	}
	if c.Database.Database == "" {
		return fmt.Errorf("DB_NAME is required")
	}

	// Validate Redis connection
	if c.Redis.Host == "" {
		return fmt.Errorf("REDIS_HOST is required")
	}

	return nil
}

// DSN returns the PostgreSQL connection string
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode,
	)
}

// Address returns the Redis connection address
func (c *RedisConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Server.Environment == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Server.Environment == "production"
}

// Helper functions to read environment variables

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := getEnv(key, "")
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultValue
}
