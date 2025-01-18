package models

import "time"

type Interaction struct {
	ID         int64     `json:"id" db:"id"`
	UserID     int64     `json:"user_id" db:"user_id"`
	EntityType string    `json:"entity_type" db:"entity_type"` // 'post' or 'comment'
	EntityID   int64     `json:"entity_id" db:"entity_id"`
	Type       string    `json:"type" db:"type"` // 'like' or 'dislike'
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}
