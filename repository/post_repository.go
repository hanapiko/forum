package repository

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"forum/models"
)

type PostRepositoryInterface interface {
	Create(post *models.Post, categoryIDs []int64) error
	GetByID(postID string) (*models.Post, error)
	ListPosts(page, limit int) ([]models.Post, int, error)
	GetPostsByCategory(categoryID int64) ([]models.Post, error)
	GetUserPosts(userID int64) ([]models.Post, error)
	GetLikedPosts(userID int64) ([]models.Post, error)
	UpdatePost(post *models.Post, categoryIDs []int64, userID int64) error
	DeletePost(postID string, userID int64) error
	FilterPosts(filters struct {
		CategoryID  *int64
		UserID      *int64
		LikedByUser *int64
	}) ([]models.Post, error)
}

type PostRepository struct {
	conn *sql.DB
}

func NewPostRepository(conn *sql.DB) *PostRepository {
	return &PostRepository{conn: conn}
}

func (r *PostRepository) Create(post *models.Post, categoryIDs []int64) error {
	// Start a transaction
	tx, err := r.conn.Begin()
	if err != nil {
		return err
	}

	// Insert post
	query := `INSERT INTO posts (user_id, title, content, created_at) 
			  VALUES (?, ?, ?, ?)`

	result, err := tx.Exec(query, post.UserID, post.Title, post.Content, time.Now())
	if err != nil {
		tx.Rollback()
		return err
	}

	// Get the ID of the newly inserted post
	postID, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return err
	}
	post.ID = postID

	// Insert post categories
	categoryQuery := `INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)`
	for _, categoryID := range categoryIDs {
		_, err = tx.Exec(categoryQuery, postID, categoryID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	// Commit the transaction
	return tx.Commit()
}

func (r *PostRepository) GetByID(postID string) (*models.Post, error) {
	// Convert string ID to int64
	id, err := strconv.ParseInt(postID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid post ID: %v", err)
	}

	// Prepare SQL query
	query := `SELECT id, user_id, title, content, created_at 
			  FROM posts WHERE id = ?`

	post := &models.Post{}
	err = r.conn.QueryRow(query, id).Scan(
		&post.ID,
		&post.UserID,
		&post.Title,
		&post.Content,
		&post.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Fetch associated categories
	categoryQuery := `SELECT category_id FROM post_categories WHERE post_id = ?`
	rows, err := r.conn.Query(categoryQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categoryIDs []int64
	for rows.Next() {
		var categoryID int64
		if err := rows.Scan(&categoryID); err != nil {
			return nil, err
		}
		categoryIDs = append(categoryIDs, categoryID)
	}
	post.Categories = categoryIDs

	return post, nil
}

func (r *PostRepository) ListPosts(page, limit int) ([]models.Post, int, error) {
	offset := (page - 1) * limit

	// First, get total count of posts
	var totalCount int
	countQuery := `SELECT COUNT(*) FROM posts`
	err := r.conn.QueryRow(countQuery).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	// Then, get paginated posts
	query := `SELECT id, user_id, title, content, created_at 
			  FROM posts 
			  ORDER BY created_at DESC 
			  LIMIT ? OFFSET ?`

	rows, err := r.conn.Query(query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Title,
			&post.Content,
			&post.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		// Fetch categories for each post
		categoryQuery := `SELECT category_id FROM post_categories WHERE post_id = ?`
		categoryRows, err := r.conn.Query(categoryQuery, post.ID)
		if err != nil {
			return nil, 0, err
		}
		defer categoryRows.Close()

		var categoryIDs []int64
		for categoryRows.Next() {
			var categoryID int64
			if err := categoryRows.Scan(&categoryID); err != nil {
				return nil, 0, err
			}
			categoryIDs = append(categoryIDs, categoryID)
		}
		post.Categories = categoryIDs

		posts = append(posts, post)
	}

	return posts, totalCount, nil
}

func (r *PostRepository) FilterPosts(filters struct {
	CategoryID  *int64
	UserID      *int64
	LikedByUser *int64
},
) ([]models.Post, error) {
	// nolint:SA4006 // Conditions slice is used later in the function
	var conditions []string
	var args []interface{}

	// Base query with a more specific initial condition
	query := `
		SELECT p.id, p.title, p.content, p.user_id, p.created_at, 
		       u.username, c.name as category_name
		FROM posts p
		JOIN users u ON p.user_id = u.id
		JOIN post_categories pc ON p.id = pc.post_id
		JOIN categories c ON pc.category_id = c.id
		WHERE 1=1
	`

	// Add conditions based on filter parameters
	if filters.CategoryID != nil {
		conditions = append(conditions, "pc.category_id = ?")
		args = append(args, *filters.CategoryID)
	}

	if filters.UserID != nil {
		conditions = append(conditions, "p.user_id = ?")
		args = append(args, *filters.UserID)
	}

	if filters.LikedByUser != nil {
		query += ` JOIN likes l ON p.id = l.post_id`
		conditions = append(conditions, "l.user_id = ?")
		args = append(args, *filters.LikedByUser)
	}

	// Modify query with conditions if any exist
	if len(conditions) > 0 {
		conditionString := strings.Join(conditions, " AND ")
		query = strings.Replace(query, "WHERE 1=1", "WHERE "+conditionString, 1)
	}

	// Add ORDER BY to ensure consistent results
	query += " ORDER BY p.created_at DESC"

	// Execute query
	rows, err := r.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		err = rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.UserID,
			&post.CreatedAt,
			&post.Username,
			&post.CategoryName,
		)
		if err != nil {
			return nil, err
		}

		posts = append(posts, post)
	}

	return posts, nil
}

func (r *PostRepository) GetPostsByCategory(categoryID int64) ([]models.Post, error) {
	return r.FilterPosts(struct {
		CategoryID  *int64
		UserID      *int64
		LikedByUser *int64
	}{
		CategoryID: &categoryID,
	})
}

func (r *PostRepository) GetUserPosts(userID int64) ([]models.Post, error) {
	return r.FilterPosts(struct {
		CategoryID  *int64
		UserID      *int64
		LikedByUser *int64
	}{
		UserID: &userID,
	})
}

func (r *PostRepository) GetLikedPosts(userID int64) ([]models.Post, error) {
	return r.FilterPosts(struct {
		CategoryID  *int64
		UserID      *int64
		LikedByUser *int64
	}{
		LikedByUser: &userID,
	})
}

func (r *PostRepository) UpdatePost(post *models.Post, categoryIDs []int64, userID int64) error {
	// Start a transaction
	tx, err := r.conn.Begin()
	if err != nil {
		return err
	}

	// Update post query
	query := `UPDATE posts 
			  SET title = ?, content = ?, updated_at = ?
			  WHERE id = ? AND user_id = ?`

	// Execute update
	result, err := tx.Exec(query,
		post.Title,
		post.Content,
		post.UpdatedAt,
		post.ID,
		userID,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}

	if rowsAffected == 0 {
		tx.Rollback()
		return fmt.Errorf("no rows updated: post not found or unauthorized")
	}

	// Delete existing post categories
	_, err = tx.Exec(`DELETE FROM post_categories WHERE post_id = ?`, post.ID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Insert new post categories
	categoryQuery := `INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)`
	for _, categoryID := range categoryIDs {
		_, err = tx.Exec(categoryQuery, post.ID, categoryID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (r *PostRepository) DeletePost(postID string, userID int64) error {
	// Convert string ID to int64
	id, err := strconv.ParseInt(postID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid post ID: %v", err)
	}

	// Start a transaction to ensure atomicity
	tx, err := r.conn.Begin()
	if err != nil {
		return err
	}

	// First, delete associated post categories
	_, err = tx.Exec(`DELETE FROM post_categories WHERE post_id = ?`, id)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Delete the post, but only if it belongs to the specified user
	result, err := tx.Exec(`DELETE FROM posts WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Check if any rows were actually deleted
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}

	if rowsAffected == 0 {
		tx.Rollback()
		return fmt.Errorf("no post found or unauthorized to delete")
	}

	// Commit the transaction
	return tx.Commit()
}

var _ PostRepositoryInterface = &PostRepository{}
