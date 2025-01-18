package models

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	UUID      string    `json:"uuid" db:"uuid"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
}

// NewSession creates a new session with a UUID and expiration
func NewSession(userID int64) *Session {
	return &Session{
		UserID:    userID,
		UUID:      uuid.New().String(),
		ExpiresAt: time.Now().Add(24 * time.Hour), // Session valid for 24 hours
	}
}

// IsExpired checks if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}
