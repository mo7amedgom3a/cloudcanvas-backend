package main

import (
	"fmt"
	"log"
	"path/filepath"

	"cloudcanvas-backend/internal/database"
)

func main() {
	// Get the migrations directory path relative to backend root
	migrationsDir := filepath.Join("migrations")

	// Run database migrations
	fmt.Println("Running database migrations...")
	if err := database.RunMigrations(migrationsDir); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	fmt.Println("✅ Migrations completed successfully!")
}
