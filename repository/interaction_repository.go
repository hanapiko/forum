package repository

import (
	"database/sql"
	"time"
)

type InteractionType string

const (
	LikeInteraction    InteractionType = "like"
	DislikeInteraction InteractionType = "dislike"
)

type Interaction struct {
	ID        int64
	UserID    int64
	PostID    *int64
	CommentID *int64
	Type      InteractionType
	CreatedAt time.Time
}

type InteractionRepository struct {
	conn *sql.DB
}

func NewInteractionRepository(conn *sql.DB) *InteractionRepository {
	return &InteractionRepository{conn: conn}
}

func (r *InteractionRepository) CreatePostInteraction(userID, postID int64, interactionType InteractionType) error {
	// First, remove any existing interaction of the opposite type
	var oppositeType InteractionType
	if interactionType == LikeInteraction {
		oppositeType = DislikeInteraction
	} else {
		oppositeType = LikeInteraction
	}

	tx, err := r.conn.Begin()
	if err != nil {
		return err
	}

	// Remove opposite interaction if exists
	_, err = tx.Exec(`
		DELETE FROM interactions 
		WHERE user_id = ? AND post_id = ? AND type = ?
	`, userID, postID, oppositeType)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Insert new interaction
	_, err = tx.Exec(`
		INSERT OR REPLACE INTO interactions (user_id, post_id, type, created_at)
		VALUES (?, ?, ?, ?)
	`, userID, postID, interactionType, time.Now())
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (r *InteractionRepository) CreateCommentInteraction(userID, commentID int64, interactionType InteractionType) error {
	// First, remove any existing interaction of the opposite type
	var oppositeType InteractionType
	if interactionType == LikeInteraction {
		oppositeType = DislikeInteraction
	} else {
		oppositeType = LikeInteraction
	}

	tx, err := r.conn.Begin()
	if err != nil {
		return err
	}

	// Remove opposite interaction if exists
	_, err = tx.Exec(`
		DELETE FROM interactions 
		WHERE user_id = ? AND comment_id = ? AND type = ?
	`, userID, commentID, oppositeType)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Insert new interaction
	_, err = tx.Exec(`
		INSERT OR REPLACE INTO interactions (user_id, comment_id, type, created_at)
		VALUES (?, ?, ?, ?)
	`, userID, commentID, interactionType, time.Now())
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (r *InteractionRepository) RemovePostInteraction(userID, postID int64, interactionType InteractionType) error {
	_, err := r.conn.Exec(`
		DELETE FROM interactions 
		WHERE user_id = ? AND post_id = ? AND type = ?
	`, userID, postID, interactionType)
	return err
}

func (r *InteractionRepository) RemoveCommentInteraction(userID, commentID int64, interactionType InteractionType) error {
	_, err := r.conn.Exec(`
		DELETE FROM interactions 
		WHERE user_id = ? AND comment_id = ? AND type = ?
	`, userID, commentID, interactionType)
	return err
}

func (r *InteractionRepository) GetPostInteractionCounts(postID int64) (likes, dislikes int64, err error) {
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN type = 'like' THEN 1 ELSE 0 END), 0) as likes,
			COALESCE(SUM(CASE WHEN type = 'dislike' THEN 1 ELSE 0 END), 0) as dislikes
		FROM interactions
		WHERE post_id = ?
	`

	err = r.conn.QueryRow(query, postID).Scan(&likes, &dislikes)
	return
}

func (r *InteractionRepository) GetCommentInteractionCounts(commentID int64) (likes, dislikes int64, err error) {
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN type = 'like' THEN 1 ELSE 0 END), 0) as likes,
			COALESCE(SUM(CASE WHEN type = 'dislike' THEN 1 ELSE 0 END), 0) as dislikes
		FROM interactions
		WHERE comment_id = ?
	`

	err = r.conn.QueryRow(query, commentID).Scan(&likes, &dislikes)
	return
}

func (r *InteractionRepository) HasUserInteractedWithPost(userID, postID int64, interactionType InteractionType) (bool, error) {
	var count int
	err := r.conn.QueryRow(`
		SELECT COUNT(*) 
		FROM interactions 
		WHERE user_id = ? AND post_id = ? AND type = ?
	`, userID, postID, interactionType).Scan(&count)

	return count > 0, err
}

func (r *InteractionRepository) HasUserInteractedWithComment(userID, commentID int64, interactionType InteractionType) (bool, error) {
	var count int
	err := r.conn.QueryRow(`
		SELECT COUNT(*) 
		FROM interactions 
		WHERE user_id = ? AND comment_id = ? AND type = ?
	`, userID, commentID, interactionType).Scan(&count)

	return count > 0, err
}
