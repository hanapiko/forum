package main

import (
	"log"
	"os"

	"forum/db"
	"forum/routes"
)

func main() {
	// Set up logging
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// Initialize database
	database, err := db.NewDatabase()
	if err != nil {
		log.Fatalf("❌ Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Run migrations
	log.Println("🚀 Running database migrations...")
	if err := database.Migrate(); err != nil {
		log.Fatalf("❌ Failed to run migrations: %v", err)
	}
	log.Println("✅ Migrations completed successfully")

	// Create router and start server
	router := routes.NewRouter(database)
	router.Start("8080")
}