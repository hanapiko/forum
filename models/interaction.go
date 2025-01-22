package models

import "time"

type InteractionType int

const (
	Like InteractionType = iota + 1
	Dislike
)

func (it InteractionType) String() string {
	switch it {
	case Like:
		return "like"
	case Dislike:
		return "dislike"
	default:
		return "unknown"
	}
}

type Interaction struct {
	ID         int64           `json:"id" db:"id"`
	UserID     int64           `json:"user_id" db:"user_id"`
	EntityType string          `json:"entity_type" db:"entity_type"` // 'post' or 'comment'
	EntityID   int64           `json:"entity_id" db:"entity_id"`
	Type       InteractionType `json:"type" db:"type"`
	CreatedAt  time.Time       `json:"created_at" db:"created_at"`
}
