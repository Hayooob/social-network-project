package main

import (
	"log"
	"os"
	"path/filepath"

	"social-network/internal/app"
	"social-network/internal/db"
)

// env returns the value of the environment variable key, or fallback when it
// is unset or empty. Every deployment knob is read this way so the same binary
// runs under `go run` and under docker-compose without a rebuild.
func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	dbPath := env("DB_PATH", "social.db")
	migrationsPath := env("MIGRATIONS_PATH", filepath.Join("internal", "db", "migrations"))
	addr := ":" + env("PORT", "8080")

	// Make sure the directory for the database exists; under docker-compose
	// DB_PATH points at a mounted volume such as /app/data/social.db.
	if dir := filepath.Dir(dbPath); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			log.Fatalf("failed to create database directory %s: %v", dir, err)
		}
	}

	database := db.OpenDb(dbPath)
	defer database.Close()

	db.RunMigrations(database, migrationsPath)

	server := app.NewServer(database)

	log.Println("Starting server on", addr)

	if err := server.Listen(addr); err != nil {
		log.Fatal(err)
	}
}
