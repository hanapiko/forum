package db

import (
	"fmt"
	"regexp"
	"time"
)

type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // Sensitive, won't be JSON encoded
	CreatedAt time.Time `json:"created_at"`
}

// Validate performs basic validation for user data
func (u *User) Validate() error {
	// Check username length
	if len(u.Username) < 3 || len(u.Username) > 50 {
		return fmt.Errorf("username must be between 3 and 50 characters")
	}

	// Basic email validation
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	if !emailRegex.MatchString(u.Email) {
		return fmt.Errorf("invalid email format")
	}

	// Password complexity (optional, can be enhanced)
	if len(u.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	return nil
}
