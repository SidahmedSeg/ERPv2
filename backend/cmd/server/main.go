package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"myerp-v2/internal/config"
	"myerp-v2/internal/database"
	"myerp-v2/internal/server"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize PostgreSQL connection
	db, err := database.NewPostgresDB(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	log.Println("✅ Connected to PostgreSQL")

	// Initialize Redis connection
	redisClient, err := database.NewRedisClient(&cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	log.Println("✅ Connected to Redis")

	// Initialize HTTP router with all dependencies
	router := server.NewRouter(db, redisClient, cfg)
	handler := router.Setup() // Call Setup() to configure routes

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("🚀 Server starting on http://localhost:%d", cfg.Server.Port)
		log.Printf("📧 Mailpit UI: http://localhost:18025")
		log.Printf("🗄️  PostgreSQL: localhost:15433")
		log.Printf("💾 Redis: localhost:26379")
		log.Println("---")
		log.Println("API Endpoints:")
		log.Println("  Health:        GET  /health")
		log.Println("  Auth:          POST /api/auth/register, /login, /logout")
		log.Println("  Users:         GET  /api/users")
		log.Println("  Roles:         GET  /api/roles")
		log.Println("  Permissions:   GET  /api/permissions")
		log.Println("  Sessions:      GET  /api/sessions")
		log.Println("  2FA:           POST /api/2fa/setup, /enable")
		log.Println("---")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited gracefully")
}
