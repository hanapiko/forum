package models

import (
	"fmt"
	"time"
)

type Post struct {
	ID         int64     `json:"id" db:"id"`
	UserID     int64     `json:"user_id" db:"user_id"`
	Title      string    `json:"title" db:"title"`
	Content    string    `json:"content" db:"content"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
	Categories []int64   `json:"categories" db:"categories"` // Category IDs
	Username   string    `json:"username" db:"username"`     // Add this line
	CategoryName string  `json:"category_name" db:"category_name"` // Add this line
}

// Validate checks if the post data is valid
func (p *Post) Validate() error {
	if p.Title == "" {
		return fmt.Errorf("title is required")
	}
	if p.Content == "" {
		return fmt.Errorf("content is required")
	}
	return nil
}
