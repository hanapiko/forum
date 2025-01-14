package main

import (
	"fmt"
	"log"
	"os"

	"forum/db"
	"forum/repository"
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

	// Create user repository
	userRepo := repository.NewUserRepository(database.Conn)

	// Diagnostic: Check database connectivity
	log.Println("🔍 Checking database connectivity...")
	pingErr := database.Conn.Ping()
	if pingErr != nil {
		log.Fatalf("❌ Database connection failed: %v", pingErr)
	}
	log.Println("✅ Database connection successful")

	// Example: Create multiple test users
	testUsers := []db.User{
		{
			Username: "johndoe",
			Email:    "john@example.com",
			Password: "password123",
		},
		{
			Username: "janedoe",
			Email:    "jane@example.com",
			Password: "securepass456",
		},
	}

	// Create and authenticate test users
	for _, testUser := range testUsers {
		log.Printf("🧪 Testing user creation: %s", testUser.Username)

		// Try to create user
		err = userRepo.Create(&testUser)
		if err != nil {
			log.Printf("❌ Error creating user %s: %v", testUser.Username, err)
			continue
		}
		log.Printf("✅ User %s created successfully", testUser.Username)

		// Try to authenticate
		log.Printf("🔐 Authenticating user: %s", testUser.Username)
		authenticatedUser, authErr := userRepo.Authenticate(testUser.Email, testUser.Password)
		if authErr != nil {
			log.Printf("❌ Authentication failed for %s: %v", testUser.Username, authErr)
			continue
		}
		log.Printf("✅ User %s authenticated successfully", authenticatedUser.Username)
	}

	// Diagnostic: List users in the database
	log.Println("📋 Listing users in the database:")
	rows, err := database.Conn.Query("SELECT id, username, email FROM users")
	if err != nil {
		log.Fatalf("❌ Failed to query users: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var username, email string
		if err := rows.Scan(&id, &username, &email); err != nil {
			log.Printf("❌ Error scanning row: %v", err)
			continue
		}
		fmt.Printf("👤 User: ID=%d, Username=%s, Email=%s\n", id, username, email)
	}

	log.Println("🎉 All tests completed!")
}
