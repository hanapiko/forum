package repository

import (
	"database/sql"
	"errors"
	"time"

	"forum/models"

	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	Create(user *models.User) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	Authenticate(email, password string) (*models.User, error)
	Update(user *models.User) error
	UpdatePassword(userID int, newPassword string) error
	FindByID(id int) (*models.User, error)
	Delete(userID int) error
}

type UserRepositoryImpl struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepositoryImpl {
	return &UserRepositoryImpl{db: db}
}

// Create creates a new user in the database
func (r *UserRepositoryImpl) Create(user *models.User) (*models.User, error) {
	// Validate user data
	if err := user.Validate(); err != nil {
		return nil, err
	}

	// Check if email already exists
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", user.Email).Scan(&count)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("email already exists")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	query := `INSERT INTO users (username, email, password, created_at) VALUES (?, ?, ?, ?)`
	now := time.Now()
	result, err := r.db.Exec(query, user.Username, user.Email, string(hashedPassword), now)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	user.ID = int(id)
	user.CreatedAt = now
	user.Password = ""
	return user, nil
}

// FindByEmail finds a user by their email
func (r *UserRepositoryImpl) FindByEmail(email string) (*models.User, error) {
	query := `SELECT id, username, email, password, created_at FROM users WHERE email = ?`
	user := &models.User{}
	err := r.db.QueryRow(query, email).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}

// FindByID finds a user by their ID
func (r *UserRepositoryImpl) FindByID(id int) (*models.User, error) {
	query := `SELECT id, username, email, created_at FROM users WHERE id = ?`
	user := &models.User{}
	err := r.db.QueryRow(query, id).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}

// UpdatePassword updates a user's password
func (r *UserRepositoryImpl) UpdatePassword(userID int, newPassword string) error {
	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	query := `UPDATE users SET password = ? WHERE id = ?`
	_, err = r.db.Exec(query, string(hashedPassword), userID)
	return err
}

// Authenticate checks user credentials
func (r *UserRepositoryImpl) Authenticate(email, password string) (*models.User, error) {
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
func (r *UserRepositoryImpl) Update(user *models.User) error {
	query := `UPDATE users 
			  SET username = ?, email = ? 
			  WHERE id = ?`

	_, err := r.db.Exec(query, user.Username, user.Email, user.ID)
	return err
}

// Delete removes a user from the database
func (r *UserRepositoryImpl) Delete(userID int) error {
	query := `DELETE FROM users WHERE id = ?`
	_, err := r.db.Exec(query, userID)
	return err
}

var _ UserRepository = &UserRepositoryImpl{}
