package main

import (
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	// Create migration instance
	m, err := migrate.New(
		"file://migrations",
		dbURL,
	)
	if err != nil {
		log.Fatalf("Failed to create migration instance: %v", err)
	}
	defer m.Close()

	// Parse command line arguments
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "up":
		if err := m.Up(); err != nil {
			if err == migrate.ErrNoChange {
				log.Println("✅ No migrations to apply")
				return
			}
			log.Fatalf("Failed to apply migrations: %v", err)
		}
		log.Println("✅ Migrations applied successfully")

	case "down":
		steps := 1
		if len(os.Args) > 2 {
			fmt.Sscanf(os.Args[2], "%d", &steps)
		}

		if err := m.Steps(-steps); err != nil {
			if err == migrate.ErrNoChange {
				log.Println("✅ No migrations to rollback")
				return
			}
			log.Fatalf("Failed to rollback migrations: %v", err)
		}
		log.Printf("✅ Rolled back %d migration(s)\n", steps)

	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			log.Fatalf("Failed to get version: %v", err)
		}
		if dirty {
			log.Printf("Current version: %d (dirty)", version)
		} else {
			log.Printf("Current version: %d", version)
		}

	case "force":
		if len(os.Args) < 3 {
			log.Fatal("Usage: migrate force VERSION")
		}
		var version int
		fmt.Sscanf(os.Args[2], "%d", &version)

		if err := m.Force(version); err != nil {
			log.Fatalf("Failed to force version: %v", err)
		}
		log.Printf("✅ Forced version to %d\n", version)

	default:
		log.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: migrate <command> [arguments]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  up              Apply all pending migrations")
	fmt.Println("  down [N]        Rollback N migrations (default: 1)")
	fmt.Println("  version         Show current migration version")
	fmt.Println("  force VERSION   Force set migration version (use with caution)")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  migrate up")
	fmt.Println("  migrate down")
	fmt.Println("  migrate down 3")
	fmt.Println("  migrate version")
	fmt.Println("  migrate force 5")
}
