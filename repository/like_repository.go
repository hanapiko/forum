package repository

import (
	"database/sql"
	"fmt"
	"forum/models"
)

// LikeRepositoryInterface defines the methods for interacting with likes in the database
type LikeRepositoryInterface interface {
	CreateLike(like *models.Interaction) error
	DeleteLike(userID, postID int) error
	GetLikesByPostID(postID int) ([]models.Interaction, error)
	UserHasLikedPost(userID, postID int) (bool, error)
	AddInteraction(userID, entityID int64, entityType string, interactionType models.InteractionType) error
	GetInteractionCounts(entityID int64, entityType string) (likes, dislikes int, err error)
	RemoveInteraction(userID, entityID int64, entityType string) error
}

// LikeRepository implements the LikeRepositoryInterface
type LikeRepository struct {
	db *sql.DB
}

// NewLikeRepository creates a new instance of LikeRepository
func NewLikeRepository(db *sql.DB) LikeRepositoryInterface {
	return &LikeRepository{db: db}
}

// CreateLike adds a new like to the database
func (r *LikeRepository) CreateLike(like *models.Interaction) error {
	query := `INSERT INTO likes (user_id, post_id) VALUES (?, ?)`
	_, err := r.db.Exec(query, like.UserID, like.EntityID)
	return err
}

// DeleteLike removes a like from the database
func (r *LikeRepository) DeleteLike(userID, postID int) error {
	query := `DELETE FROM likes WHERE user_id = ? AND post_id = ?`
	_, err := r.db.Exec(query, userID, postID)
	return err
}

// GetLikesByPostID retrieves all likes for a specific post
func (r *LikeRepository) GetLikesByPostID(postID int) ([]models.Interaction, error) {
	query := `SELECT id, user_id, post_id FROM likes WHERE post_id = ?`
	rows, err := r.db.Query(query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var likes []models.Interaction
	for rows.Next() {
		var like models.Interaction
		err := rows.Scan(&like.ID, &like.UserID, &like.EntityID)
		if err != nil {
			return nil, err
		}
		likes = append(likes, like)
	}
	return likes, nil
}

// UserHasLikedPost checks if a user has already liked a post
func (r *LikeRepository) UserHasLikedPost(userID, postID int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM likes WHERE user_id = ? AND post_id = ?)`
	var exists bool
	err := r.db.QueryRow(query, userID, postID).Scan(&exists)
	return exists, err
}

// AddInteraction adds a new interaction to the database
func (r *LikeRepository) AddInteraction(userID, entityID int64, entityType string, interactionType models.InteractionType) error {
	query := `
		INSERT INTO interactions (user_id, entity_id, entity_type, type)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE type = ?
	`
	_, err := r.db.Exec(query, userID, entityID, entityType, interactionType, interactionType)
	return err
}

// GetInteractionCounts retrieves the interaction counts for a specific entity
func (r *LikeRepository) GetInteractionCounts(entityID int64, entityType string) (likes, dislikes int, err error) {
	query := `SELECT COUNT(*) FROM likes WHERE post_id = ?`
	err = r.db.QueryRow(query, entityID).Scan(&likes)
	if err != nil {
		return 0, 0, fmt.Errorf("error getting interaction counts: %v", err)
	}
	return likes, 0, nil
}

// RemoveInteraction removes an interaction from the database
func (r *LikeRepository) RemoveInteraction(userID, entityID int64, entityType string) error {
	query := `DELETE FROM likes WHERE user_id = ? AND post_id = ?`
	_, err := r.db.Exec(query, userID, entityID)
	if err != nil {
		return fmt.Errorf("error removing interaction: %v", err)
	}
	return nil
}
