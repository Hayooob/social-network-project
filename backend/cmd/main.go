package main

import (
	"log"
	"path/filepath"
	"social-network/internal/app"
	"social-network/internal/db"
)

func main() {
	dbPath := "social.db"
	migrationsPath := filepath.Join("internal", "db", "migrations")

	database := db.OpenDb(dbPath)
	defer database.Close()

	db.RunMigrations(database, migrationsPath) // Pass the path
	
	server := app.NewServer(database)
	addr := ":8080"
	
	log.Println("Starting server on", addr)

	if err := server.Listen(addr); err != nil {
		log.Fatal(err)
	}
}