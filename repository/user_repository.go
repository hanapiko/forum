package repository

import (
	"database/sql"
	"errors"
	"time"

	"forum/models"

	"golang.org/x/crypto/bcrypt"
)

type UserRepository struct {
	conn *sql.DB
}

func NewUserRepository(conn *sql.DB) *UserRepository {
	return &UserRepository{conn: conn}
}

// Create creates a new user in the database
func (r *UserRepository) Create(user *models.User) error {
	// Validate user data
	if err := user.Validate(); err != nil {
		return err
	}

	// Check if email already exists
	var count int
	err := r.conn.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", user.Email).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("email already exists")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Prepare SQL insert statement
	query := `INSERT INTO users (username, email, password, created_at) 
			  VALUES (?, ?, ?, ?)`

	// Execute the query
	result, err := r.conn.Exec(query, user.Username, user.Email, string(hashedPassword), time.Now())
	if err != nil {
		return err
	}

	// Get the ID of the newly inserted user
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	user.ID = id

	// Clear password for security
	user.Password = ""

	return nil
}

// FindByEmail finds a user by their email
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	query := `SELECT id, username, email, password, created_at 
			  FROM users WHERE email = ?`

	user := &models.User{}
	err := r.conn.QueryRow(query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
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
func (r *UserRepository) Authenticate(email, password string) (*models.User, error) {
	// Find user by email
	user, err := r.FindByEmail(email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Clear password for security
	user.Password = ""

	return user, nil
}

// Update updates user information
func (r *UserRepository) Update(user *models.User) error {
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
