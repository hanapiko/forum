package repository

import (
	"database/sql"
	"errors"
	"time"

	"forum/models"
)

type SessionRepositoryInterface interface {
	CreateSession(user *models.User) error
	ValidateSession(sessionToken string) (*models.User, error)
	InvalidateSession(userID int) error
	Validate(token string) (int64, error)
}

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// CreateSession stores a new session for a user
func (r *SessionRepository) CreateSession(user *models.User) error {
	// Generate a new session token
	user.GenerateSessionToken()

	// Prepare SQL to insert or update session
	query := `
		INSERT INTO sessions (user_id, session_token, created_at, expires_at) 
		VALUES (?, ?, ?, ?) 
		ON CONFLICT(user_id) DO UPDATE SET 
		session_token = ?, 
		created_at = ?, 
		expires_at = ?
	`

	_, err := r.db.Exec(query,
		user.ID,
		user.SessionToken,
		time.Now(),
		user.SessionExpiry,
		user.SessionToken,
		time.Now(),
		user.SessionExpiry,
	)

	return err
}

// ValidateSession checks if a session token is valid
func (r *SessionRepository) ValidateSession(sessionToken string) (*models.User, error) {
	query := `
		SELECT u.id, u.username, u.email, s.session_token, s.expires_at
		FROM users u
		JOIN sessions s ON u.id = s.user_id
		WHERE s.session_token = ? AND s.expires_at > ?
	`

	user := &models.User{}
	err := r.db.QueryRow(query, sessionToken, time.Now()).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.SessionToken,
		&user.SessionExpiry,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("invalid or expired session")
		}
		return nil, err
	}

	return user, nil
}

// InvalidateSession removes or expires a user's session
func (r *SessionRepository) InvalidateSession(userID int) error {
	query := `
		DELETE FROM sessions 
		WHERE user_id = ?
	`

	_, err := r.db.Exec(query, userID)
	return err
}

// Validate checks if a session token is valid and returns the user ID
func (r *SessionRepository) Validate(token string) (int64, error) {
	// Implement the Validate method to match the interface
	user, err := r.ValidateSession(token)
	if err != nil {
		return 0, err
	}
	return int64(user.ID), nil
}
