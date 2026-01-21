package server

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"

	"myerp-v2/internal/config"
	"myerp-v2/internal/handlers"
	appMiddleware "myerp-v2/internal/middleware"
	"myerp-v2/internal/repository"
	"myerp-v2/internal/services"
)

// Router creates and configures the HTTP router
type Router struct {
	router *chi.Mux
	db     *sqlx.DB
	redis  *redis.Client
	config *config.Config
}

// NewRouter creates a new router instance
func NewRouter(db *sqlx.DB, redis *redis.Client, cfg *config.Config) *Router {
	return &Router{
		router: chi.NewRouter(),
		db:     db,
		redis:  redis,
		config: cfg,
	}
}

// Setup configures all routes and middleware
func (s *Router) Setup() *chi.Mux {
	// Global middleware
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.Timeout(60 * time.Second))

	// CORS middleware
	allowedOrigins := []string{s.config.App.FrontendURL, s.config.App.BaseURL}
	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Initialize repositories
	tenantRepo := repository.NewTenantRepository(s.db)
	userRepo := repository.NewUserRepository(s.db)
	sessionRepo := repository.NewSessionRepository(s.db)
	roleRepo := repository.NewRoleRepository(s.db)
	permissionRepo := repository.NewPermissionRepository(s.db)
	userRoleRepo := repository.NewUserRoleRepository(s.db)
	companySettingsRepo := repository.NewCompanySettingsRepository(s.db)

	// Initialize services
	jwtService := services.NewJWTService(&s.config.JWT)
	emailService := services.NewEmailService(&s.config.Email, &s.config.App)
	authService := services.NewAuthService(tenantRepo, userRepo, sessionRepo, roleRepo, userRoleRepo, jwtService, emailService, s.config)
	permissionService := services.NewPermissionService(permissionRepo, userRoleRepo, roleRepo, s.redis)
	twoFactorService := services.NewTwoFactorService(s.db, s.redis, s.config)
	sessionService := services.NewSessionService(s.db)
	invitationService := services.NewInvitationService(s.db, userRepo, userRoleRepo, emailService)
	auditService := services.NewAuditService(s.db)
	companySettingsService := services.NewCompanySettingsService(companySettingsRepo, tenantRepo, auditService)

	// Initialize middleware
	tenantMiddleware := appMiddleware.NewTenantMiddleware(tenantRepo)
	authMiddleware := appMiddleware.NewAuthMiddleware(authService)
	permMiddleware := appMiddleware.NewPermissionMiddleware(permissionService)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userRepo, userRoleRepo, permissionService)
	roleHandler := handlers.NewRoleHandler(roleRepo, userRoleRepo, permissionService)
	permissionHandler := handlers.NewPermissionHandler(permissionService)
	twoFactorHandler := handlers.NewTwoFactorHandler(twoFactorService, userRepo)
	sessionHandler := handlers.NewSessionHandler(sessionService)
	invitationHandler := handlers.NewInvitationHandler(invitationService)
	auditHandler := handlers.NewAuditHandler(auditService)
	securityHandler := handlers.NewSecurityHandler(auditService, sessionService, twoFactorService)
	companySettingsHandler := handlers.NewCompanySettingsHandler(companySettingsService)

	// Health check endpoint
	s.router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Apply tenant resolution middleware to all routes (except /health)
	s.router.Group(func(r chi.Router) {
		r.Use(tenantMiddleware.ResolveTenant)

		// Public routes (no authentication required)
		authHandler.RegisterRoutes(r, authMiddleware, tenantMiddleware) // Includes login, register, verify-email, etc.
		invitationHandler.RegisterRoutes(r, authMiddleware, permMiddleware) // Accept invitation is public

		// Protected routes (authentication required)
		// Week 2: Authentication Core
		// (Auth routes are already registered above)

		// Week 3: RBAC System
		userHandler.RegisterRoutes(r, authMiddleware, permMiddleware)
		roleHandler.RegisterRoutes(r, authMiddleware, permMiddleware)
		permissionHandler.RegisterRoutes(r, authMiddleware, permMiddleware)

		// Week 4: Advanced Features
		twoFactorHandler.RegisterRoutes(r, authMiddleware)
		sessionHandler.RegisterRoutes(r, authMiddleware)
		// Invitation routes already registered above
		auditHandler.RegisterRoutes(r, authMiddleware, permMiddleware)
		securityHandler.RegisterRoutes(r, authMiddleware, permMiddleware)

		// Company Settings
		companySettingsHandler.RegisterRoutes(r, authMiddleware)
	})

	return s.router
}

// GetRouter returns the configured router
func (s *Router) GetRouter() *chi.Mux {
	return s.router
}
