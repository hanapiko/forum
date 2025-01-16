package repository

import (
	"database/sql"
	"errors"
	"time"

	"forum/db"
)

type UserRepository struct {
	conn *sql.DB
}

func NewUserRepository(conn *sql.DB) *UserRepository {
	return &UserRepository{conn: conn}
}

// Create creates a new user in the database
func (r *UserRepository) Create(user *db.User, password string) error {
	// Validate user data
	if err := user.Validate(); err != nil {
		return err
	}

	// Hash the password
	passwordHash, err := db.HashPassword(password)
	if err != nil {
		return err
	}

	// Prepare SQL insert statement
	query := `INSERT INTO users (username, email, password_hash, created_at) 
			  VALUES (?, ?, ?, ?)`

	// Execute the query
	result, err := r.conn.Exec(query, user.Username, user.Email, passwordHash, time.Now())
	if err != nil {
		// Check for unique constraint violation
		return err
	}

	// Get the ID of the newly inserted user
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	user.ID = id

	return nil
}

// FindByEmail finds a user by their email
func (r *UserRepository) FindByEmail(email string) (*db.User, error) {
	query := `SELECT id, username, email, password_hash, created_at 
			  FROM users WHERE email = ?`

	user := &db.User{}
	err := r.conn.QueryRow(query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Authenticate checks user credentials
func (r *UserRepository) Authenticate(email, password string) (*db.User, error) {
	// Find user by email
	user, err := r.FindByEmail(email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Check password
	if !db.CheckPasswordHash(password, user.PasswordHash) {
		return nil, errors.New("invalid email or password")
	}

	return user, nil
}

// Update updates user information
func (r *UserRepository) Update(user *db.User) error {
	query := `UPDATE users 
			  SET username = ?, email = ? 
			  WHERE id = ?`

	_, err := r.conn.Exec(query, user.Username, user.Email, user.ID)
	return err
}

// Delete removes a user from the database
func (r *UserRepository) Delete(userID int64) error {
	query := `DELETE FROM users WHERE id = ?`

	_, err := r.conn.Exec(query, userID)
	return err
}
