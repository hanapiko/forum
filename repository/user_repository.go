package repository

import (
	"database/sql"
	"errors"
	"time"

	"forum/db"

	"golang.org/x/crypto/bcrypt"
)

// UserRepository handles database operations for users
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new instance of UserRepository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create adds a new user to the database
func (r *UserRepository) Create(user *db.User) error {
	// Validate user data
	if err := user.Validate(); err != nil {
		return err
	}

	// Check if email already exists
	existingUser, _ := r.FindByEmail(user.Email)
	if existingUser != nil {
		return errors.New("email already registered")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Insert user into database
	query := `INSERT INTO users (username, email, password_hash, created_at) VALUES (?, ?, ?, ?)`
	result, err := r.db.Exec(query, user.Username, user.Email, string(hashedPassword), time.Now())
	if err != nil {
		return err
	}

	// Set the ID of the created user
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	user.ID = id
	user.CreatedAt = time.Now()

	return nil
}

// FindByEmail retrieves a user by their email
func (r *UserRepository) FindByEmail(email string) (*db.User, error) {
	user := &db.User{}
	query := `SELECT id, username, email, password_hash, created_at FROM users WHERE email = ?`
	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}

// Authenticate checks user credentials
func (r *UserRepository) Authenticate(email, password string) (*db.User, error) {
	user, err := r.FindByEmail(email)
	if err != nil {
		return nil, err
	}

	// Compare passwords
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	return user, nil
}

// UpdatePassword allows a user to change their password
func (r *UserRepository) UpdatePassword(userID int64, newPassword string) error {
	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update password in database
	query := `UPDATE users SET password_hash = ? WHERE id = ?`
	_, err = r.db.Exec(query, string(hashedPassword), userID)
	return err
}
