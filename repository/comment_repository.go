package repository

import (
	"database/sql"
	"time"
)

type Comment struct {
	ID        int64
	PostID    int64
	UserID    int64
	Content   string
	CreatedAt time.Time
}

// CommentRepositoryInterface defines the methods that a comment repository must implement
type CommentRepositoryInterface interface {
	Create(comment *Comment) error
	GetByPostID(postID int64) ([]Comment, error)
}

type CommentRepository struct {
	conn *sql.DB
}

func NewCommentRepository(conn *sql.DB) *CommentRepository {
	return &CommentRepository{conn: conn}
}

func (r *CommentRepository) Create(comment *Comment) error {
	query := `INSERT INTO comments (post_id, user_id, content, created_at) 
			  VALUES (?, ?, ?, ?)`
	
	result, err := r.conn.Exec(query, comment.PostID, comment.UserID, comment.Content, time.Now())
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	comment.ID = id
	comment.CreatedAt = time.Now()

	return nil
}

func (r *CommentRepository) GetByPostID(postID int64) ([]Comment, error) {
	query := `SELECT id, post_id, user_id, content, created_at 
			  FROM comments WHERE post_id = ? ORDER BY created_at DESC`
	
	rows, err := r.conn.Query(query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		err := rows.Scan(
			&comment.ID, 
			&comment.PostID, 
			&comment.UserID, 
			&comment.Content, 
			&comment.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

// Ensure CommentRepository implements the interface
var _ CommentRepositoryInterface = &CommentRepository{}