package models

import (
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID            int       `json:"id"`
	Username      string    `json:"username"`
	Email         string    `json:"email"`
	Password      string    `json:"password,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	LastLogin     time.Time `json:"last_login"`
	SessionToken  string    `json:"session_token,omitempty"`
	SessionExpiry time.Time `json:"session_expiry"`
}

// Validate checks if user data is valid
func (u *User) Validate() error {
	if u.Username == "" {
		return fmt.Errorf("username is required")
	}
	if len(u.Username) < 3 {
		return fmt.Errorf("username must be at least 3 characters long")
	}
	if u.Email == "" {
		return fmt.Errorf("email is required")
	}
	if !isValidEmail(u.Email) {
		return fmt.Errorf("invalid email format")
	}
	if u.Password == "" {
		return fmt.Errorf("password is required")
	}
	if len(u.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}
	return nil
}

// isValidEmail checks if the email format is valid
func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return emailRegex.MatchString(email)
}

// GenerateSessionToken creates a new session token and sets expiry
func (u *User) GenerateSessionToken() {
	u.SessionToken = uuid.New().String()
	u.SessionExpiry = time.Now().Add(24 * time.Hour) // Session valid for 24 hours
	u.LastLogin = time.Now()
}

// IsSessionValid checks if the current session is still valid
func (u *User) IsSessionValid() bool {
	return u.SessionExpiry.After(time.Now())
}

// InvalidateSession clears the session token and expiry
func (u *User) InvalidateSession() {
	u.SessionToken = ""
	u.SessionExpiry = time.Time{}
}

// RegisterRequest represents the data needed for user registration
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
