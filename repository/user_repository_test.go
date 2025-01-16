package repository

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"forum/db"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	// Create a temporary in-memory database for testing
	database, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Run migrations
	_, err = database.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	// Cleanup function to close the database
	cleanup := func() {
		database.Close()
	}

	return database, cleanup
}

func TestUserCreation(t *testing.T) {
	testDB, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(testDB)

	// Test successful user creation
	user := &db.User{
		Username: "testuser",
		Email:    "test@example.com",
	}

	err := repo.Create(user, "StrongPass123")
	if err != nil {
		t.Errorf("Failed to create user: %v", err)
	}

	// Verify user was created
	foundUser, err := repo.FindByEmail("test@example.com")
	if err != nil {
		t.Errorf("Failed to find created user: %v", err)
	}
	if foundUser.Username != "testuser" {
		t.Errorf("Unexpected username: got %s, want testuser", foundUser.Username)
	}
}

func TestUserAuthentication(t *testing.T) {
	testDB, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(testDB)

	// Create a test user
	user := &db.User{
		Username: "authuser",
		Email:    "auth@example.com",
	}

	err := repo.Create(user, "ValidPass123")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Test successful authentication
	authenticatedUser, err := repo.Authenticate("auth@example.com", "ValidPass123")
	if err != nil {
		t.Errorf("Authentication failed: %v", err)
	}
	if authenticatedUser == nil {
		t.Error("Authentication returned nil user")
	}

	// Test failed authentication
	_, err = repo.Authenticate("auth@example.com", "WrongPassword")
	if err == nil {
		t.Error("Authentication should fail with wrong password")
	}
}

func TestUserEmailUniqueness(t *testing.T) {
	testDB, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(testDB)

	// Create first user
	user1 := &db.User{
		Username: "user1",
		Email:    "unique@example.com",
	}
	err := repo.Create(user1, "Password123")
	if err != nil {
		t.Fatalf("Failed to create first user: %v", err)
	}

	// Try to create user with same email
	user2 := &db.User{
		Username: "user2",
		Email:    "unique@example.com",
	}
	err = repo.Create(user2, "AnotherPass123")
	if err == nil {
		t.Error("Should not allow creating user with duplicate email")
	}
}

func TestUserUpdate(t *testing.T) {
	testDB, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(testDB)

	// Create a test user
	user := &db.User{
		Username: "updateuser",
		Email:    "update@example.com",
	}

	err := repo.Create(user, "InitialPass123")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Update user details
	user.Username = "updatedusername"
	user.Email = "newemail@example.com"

	err = repo.Update(user)
	if err != nil {
		t.Errorf("Failed to update user: %v", err)
	}

	// Verify updated user
	updatedUser, err := repo.FindByEmail("newemail@example.com")
	if err != nil {
		t.Errorf("Failed to find updated user: %v", err)
	}
	if updatedUser.Username != "updatedusername" {
		t.Errorf("Username not updated: got %s, want updatedusername", updatedUser.Username)
	}
}

func TestUserDelete(t *testing.T) {
	testDB, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(testDB)

	// Create a test user
	user := &db.User{
		Username: "deleteuser",
		Email:    "delete@example.com",
	}

	err := repo.Create(user, "DeletePass123")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Delete the user
	err = repo.Delete(user.ID)
	if err != nil {
		t.Errorf("Failed to delete user: %v", err)
	}

	// Verify user is deleted
	_, err = repo.FindByEmail("delete@example.com")
	if err == nil {
		t.Error("User should not exist after deletion")
	}
}
