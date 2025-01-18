package repository

import (
	"database/sql"
	"errors"
	"time"
)

type InteractionType int

const (
	Like InteractionType = iota + 1
	Dislike
)

type Interaction struct {
	ID         int64           `json:"id"`
	UserID     int64           `json:"user_id"`
	EntityID   int64           `json:"entity_id"`
	EntityType string          `json:"entity_type"` // 'post' or 'comment'
	Type       InteractionType `json:"type"`
	CreatedAt  time.Time       `json:"created_at"`
}

type InteractionRepository struct {
	conn *sql.DB
}

func NewInteractionRepository(conn *sql.DB) *InteractionRepository {
	return &InteractionRepository{conn: conn}
}

// AddInteraction adds a like or dislike to a post or comment
func (r *InteractionRepository) AddInteraction(userID, entityID int64, entityType string, interactionType InteractionType) error {
	// Validate input
	if userID <= 0 || entityID <= 0 || entityType == "" {
		return errors.New("invalid interaction parameters")
	}

	// Start a transaction
	tx, err := r.conn.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		tx.Commit()
	}()

	// Check if the entity exists (basic validation)
	var exists int
	var checkQuery string
	switch entityType {
	case "post":
		checkQuery = "SELECT COUNT(*) FROM posts WHERE id = ?"
	case "comment":
		checkQuery = "SELECT COUNT(*) FROM comments WHERE id = ?"
	default:
		return errors.New("invalid entity type")
	}

	err = tx.QueryRow(checkQuery, entityID).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		return errors.New("entity not found")
	}

	// Check if user has already interacted with this entity
	var currentInteraction InteractionType
	err = tx.QueryRow(`
		SELECT type FROM interactions 
		WHERE user_id = ? AND entity_id = ? AND entity_type = ?
	`, userID, entityID, entityType).Scan(&currentInteraction)

	if err != nil && err != sql.ErrNoRows {
		return err
	}

	// If interaction exists, update it
	if err == nil {
		_, err = tx.Exec(`
			UPDATE interactions 
			SET type = ?, created_at = ? 
			WHERE user_id = ? AND entity_id = ? AND entity_type = ?
		`, interactionType, time.Now(), userID, entityID, entityType)
		return err
	}

	// Insert new interaction
	_, err = tx.Exec(`
		INSERT INTO interactions (user_id, entity_id, entity_type, type, created_at) 
		VALUES (?, ?, ?, ?, ?)
	`, userID, entityID, entityType, interactionType, time.Now())
	return err
}

// GetInteractionCounts retrieves like and dislike counts for a specific entity
func (r *InteractionRepository) GetInteractionCounts(entityID int64, entityType string) (likes, dislikes int, err error) {
	// Validate input
	if entityID <= 0 || entityType == "" {
		return 0, 0, errors.New("invalid input parameters")
	}

	query := `
		SELECT 
			SUM(CASE WHEN type = ? THEN 1 ELSE 0 END) as likes,
			SUM(CASE WHEN type = ? THEN 1 ELSE 0 END) as dislikes
		FROM interactions 
		WHERE entity_id = ? AND entity_type = ?
	`
	err = r.conn.QueryRow(query, Like, Dislike, entityID, entityType).Scan(&likes, &dislikes)
	if err == sql.ErrNoRows {
		return 0, 0, nil
	}
	if err != nil {
		return 0, 0, err
	}
	return
}

// RemoveInteraction removes a user's interaction with an entity
func (r *InteractionRepository) RemoveInteraction(userID, entityID int64, entityType string) error {
	// Validate input
	if userID <= 0 || entityID <= 0 || entityType == "" {
		return errors.New("invalid input parameters")
	}

	query := `
		DELETE FROM interactions 
		WHERE user_id = ? AND entity_id = ? AND entity_type = ?
	`
	_, err := r.conn.Exec(query, userID, entityID, entityType)
	return err
}
