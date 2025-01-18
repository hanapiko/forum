package repository

import (
	"database/sql"
	"errors"
	"time"

	"forum/models"
)

type SessionRepository struct {
	conn *sql.DB
}

func NewSessionRepository(conn *sql.DB) *SessionRepository {
	return &SessionRepository{conn: conn}
}

// CreateSession creates a new session for a user
// Ensures only one active session per user
func (r *SessionRepository) CreateSession(userID int64) (*models.Session, error) {
	// Validate input
	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}

	// Start a transaction
	tx, err := r.conn.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		tx.Commit()
	}()

	// First, invalidate any existing sessions for this user
	_, err = tx.Exec("DELETE FROM sessions WHERE user_id = ?", userID)
	if err != nil {
		return nil, err
	}

	// Create new session
	session := models.NewSession(userID)

	query := `
		INSERT INTO sessions (user_id, uuid, expires_at) 
		VALUES (?, ?, ?)
	`
	_, err = tx.Exec(query, session.UserID, session.UUID, session.ExpiresAt)
	if err != nil {
		return nil, err
	}

	return session, nil
}

// ValidateSession checks if a session is valid
func (r *SessionRepository) ValidateSession(sessionUUID string) (*models.Session, error) {
	// Validate input
	if sessionUUID == "" {
		return nil, errors.New("empty session UUID")
	}

	var session models.Session
	query := `
		SELECT id, user_id, uuid, expires_at 
		FROM sessions 
		WHERE uuid = ? AND expires_at > ?
	`
	err := r.conn.QueryRow(query, sessionUUID, time.Now()).Scan(
		&session.ID,
		&session.UserID,
		&session.UUID,
		&session.ExpiresAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("invalid or expired session")
		}
		return nil, err
	}

	return &session, nil
}

// DeleteSession removes a session from the database
func (r *SessionRepository) DeleteSession(sessionUUID string) error {
	// Validate input
	if sessionUUID == "" {
		return errors.New("empty session UUID")
	}

	_, err := r.conn.Exec("DELETE FROM sessions WHERE uuid = ?", sessionUUID)
	return err
}

// GetActiveSessionByUserID retrieves an active session for a specific user
func (r *SessionRepository) GetActiveSessionByUserID(userID int64) (*models.Session, error) {
	// Validate input
	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}

	var session models.Session
	query := `
		SELECT id, user_id, uuid, expires_at 
		FROM sessions 
		WHERE user_id = ? AND expires_at > ?
		LIMIT 1
	`
	err := r.conn.QueryRow(query, userID, time.Now()).Scan(
		&session.ID,
		&session.UserID,
		&session.UUID,
		&session.ExpiresAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No active session
		}
		return nil, err
	}

	return &session, nil
}

// CleanupExpiredSessions removes all expired sessions
func (r *SessionRepository) CleanupExpiredSessions() error {
	_, err := r.conn.Exec("DELETE FROM sessions WHERE expires_at <= ?", time.Now())
	return err
}
