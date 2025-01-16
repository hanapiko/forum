package repository

import (
	"database/sql"
	"errors"
	"time"
)

type Post struct {
	ID         int64
	UserID     int64
	CategoryID int64
	Title      string
	Content    string
	CreatedAt  time.Time
}

type PostRepository struct {
	conn *sql.DB
}

func NewPostRepository(conn *sql.DB) *PostRepository {
	return &PostRepository{conn: conn}
}

func (r *PostRepository) Create(post *Post) error {
	query := `INSERT INTO posts (user_id, category_id, title, content, created_at) 
			  VALUES (?, ?, ?, ?, ?)`
	
	result, err := r.conn.Exec(query, post.UserID, post.CategoryID, post.Title, post.Content, time.Now())
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	post.ID = id
	post.CreatedAt = time.Now()

	return nil
}

func (r *PostRepository) GetByID(postID int64) (*Post, error) {
	query := `SELECT id, user_id, category_id, title, content, created_at 
			  FROM posts WHERE id = ?`
	
	post := &Post{}
	err := r.conn.QueryRow(query, postID).Scan(
		&post.ID, 
		&post.UserID, 
		&post.CategoryID, 
		&post.Title, 
		&post.Content, 
		&post.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return post, nil
}

func (r *PostRepository) ListPosts() ([]Post, error) {
	query := `SELECT id, user_id, category_id, title, content, created_at 
			  FROM posts ORDER BY created_at DESC`
	
	rows, err := r.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		err := rows.Scan(
			&post.ID, 
			&post.UserID, 
			&post.CategoryID, 
			&post.Title, 
			&post.Content, 
			&post.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}

func (r *PostRepository) FilterPosts(categoryID, userID int64) ([]Post, error) {
	var query string
	var args []interface{}

	// Build dynamic query based on filter parameters
	switch {
	case categoryID > 0 && userID > 0:
		query = `SELECT id, user_id, category_id, title, content, created_at 
				 FROM posts WHERE category_id = ? AND user_id = ?`
		args = []interface{}{categoryID, userID}
	case categoryID > 0:
		query = `SELECT id, user_id, category_id, title, content, created_at 
				 FROM posts WHERE category_id = ?`
		args = []interface{}{categoryID}
	case userID > 0:
		query = `SELECT id, user_id, category_id, title, content, created_at 
				 FROM posts WHERE user_id = ?`
		args = []interface{}{userID}
	default:
		return nil, errors.New("no filter criteria provided")
	}

	rows, err := r.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		err := rows.Scan(
			&post.ID, 
			&post.UserID, 
			&post.CategoryID, 
			&post.Title, 
			&post.Content, 
			&post.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}