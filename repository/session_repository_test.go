package repository

import (
	"database/sql"
	"testing"
	"time"

	"forum/models"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func setupSessionTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	assert.NoError(t, err, "Failed to open test database")

	// Create necessary tables for testing
	_, err = db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			created_at DATETIME,
			last_login DATETIME
		);
		CREATE TABLE sessions (
			user_id INTEGER PRIMARY KEY,
			session_token TEXT NOT NULL,
			created_at DATETIME,
			expires_at DATETIME,
			FOREIGN KEY(user_id) REFERENCES users(id)
		);
	`)
	assert.NoError(t, err, "Failed to create test tables")

	return db
}

func TestCreateSession(t *testing.T) {
	db := setupSessionTestDB(t)
	defer db.Close()

	// Insert a test user first
	_, err := db.Exec(`
		INSERT INTO users (username, email, password, created_at) 
		VALUES (?, ?, ?, ?)
	`, "testuser", "test@example.com", "hashedpassword", time.Now())
	assert.NoError(t, err)

	// Get the user to create a session for
	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
	}

	sessionRepo := NewSessionRepository(db)
	err = sessionRepo.CreateSession(user)
	assert.NoError(t, err)

	// Verify session was created
	var sessionToken string
	var expiresAt time.Time
	err = db.QueryRow(`
		SELECT session_token, expires_at 
		FROM sessions 
		WHERE user_id = ?
	`, user.ID).Scan(&sessionToken, &expiresAt)

	assert.NoError(t, err)
	assert.NotEmpty(t, sessionToken)
	assert.False(t, expiresAt.IsZero())
}

func TestValidateSession(t *testing.T) {
	db := setupSessionTestDB(t)
	defer db.Close()

	// Insert a test user
	_, err := db.Exec(`
		INSERT INTO users (id, username, email, password, created_at) 
		VALUES (?, ?, ?, ?, ?)
	`, 1, "testuser", "test@example.com", "hashedpassword", time.Now())
	assert.NoError(t, err)

	// Create a session
	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
	}
	user.GenerateSessionToken()

	sessionRepo := NewSessionRepository(db)
	err = sessionRepo.CreateSession(user)
	assert.NoError(t, err)

	// Test valid session validation
	validatedUser, err := sessionRepo.ValidateSession(user.SessionToken)
	assert.NoError(t, err)
	assert.NotNil(t, validatedUser)
	assert.Equal(t, user.ID, validatedUser.ID)
}

func TestInvalidateSession(t *testing.T) {
	db := setupSessionTestDB(t)
	defer db.Close()

	// Insert a test user
	_, err := db.Exec(`
		INSERT INTO users (id, username, email, password, created_at) 
		VALUES (?, ?, ?, ?, ?)
	`, 1, "testuser", "test@example.com", "hashedpassword", time.Now())
	assert.NoError(t, err)

	// Create a session
	sessionRepo := NewSessionRepository(db)
	err = sessionRepo.InvalidateSession(1)
	assert.NoError(t, err)

	// Verify session is invalidated
	var count int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM sessions WHERE user_id = ?
	`, 1).Scan(&count)

	assert.NoError(t, err)
	assert.Zero(t, count)
}
