package main

import (
	"log"
	"social-network/internal/app"
	"social-network/internal/db"
)

func main() {
		dbPath := "social.db" 

	// Open the SQLite DB
	database := db.OpenDb(dbPath)
	defer database.Close()

	// run all migrations
	db.RunMigrations(database)

	// create server with db
	server := app.NewServer(database)
	
		addr := ":8080"
	

	log.Println("Starting server on", addr)

	// start the HTTP server
	if err := server.Listen(addr); err != nil {
		log.Fatal(err)
	}
}
