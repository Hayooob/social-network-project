package db

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
		_ "github.com/mattn/go-sqlite3"
)

// opens a sql database at the given path and pings it
func OpenDb(path string) *sql.DB {
	//will open .db file in the req path
	database, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
//if db doesnt open 
	if err := database.Ping(); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}
	// otherwise return db connection 
	return database
}

// reads all .sql mig files and runs them
func RunMigrations(db *sql.DB) {
	migrationsDir := "internal/db/migrations" 

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		log.Printf("could not read migrations directory %s: %v", migrationsDir, err)
		return
	}
// loop through each entry 
	for _, entry := range entries {
		//only run sql files not directories incase we add folders in the future 
		if entry.IsDir() {
			continue
		}

		path := filepath.Join(migrationsDir, entry.Name())
		sqlBytes, err := os.ReadFile(path)
		if err != nil {
			log.Fatalf("failed to read migration %s: %v", path, err)
		}
//execute migration
		log.Printf("running migration %s", entry.Name())
		if _, err := db.Exec(string(sqlBytes)); err != nil {
			log.Fatalf("failed to execute migration %s: %v", path, err)
		}
	}
}
