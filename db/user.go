package db

import (
	"fmt"
	"regexp"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Stored hashed password
	CreatedAt    time.Time `json:"created_at"`
}

// Validate performs comprehensive validation for user data
func (u *User) Validate() error {
	// Check username length and format
	if len(u.Username) < 3 || len(u.Username) > 50 {
		return fmt.Errorf("username must be between 3 and 50 characters")
	}
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !usernameRegex.MatchString(u.Username) {
		return fmt.Errorf("username can only contain letters, numbers, and underscores")
	}

	// Email validation with more robust regex
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(u.Email) {
		return fmt.Errorf("invalid email format")
	}

	return nil
}

// HashPassword generates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	// Validate password complexity
	if len(password) < 8 {
		return "", fmt.Errorf("password must be at least 8 characters long")
	}

	// Check for at least one uppercase, one lowercase, and one number
	hasUppercase := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLowercase := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)

	if !hasUppercase || !hasLowercase || !hasNumber {
		return "", fmt.Errorf("password must contain at least one uppercase letter, one lowercase letter, and one number")
	}

	// Generate bcrypt hash
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPasswordHash compares a password with its hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
