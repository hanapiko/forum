package repository

import (
	"database/sql"
	"strings"
	"time"

	"forum/models"
)

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

func (r *PostRepository) GetByID(postID int64) (*models.Post, error) {
	// Query to get post details
	query := `SELECT id, user_id, title, content, created_at 
			  FROM posts WHERE id = ?`

	post := &models.Post{}
	err := r.conn.QueryRow(query, postID).Scan(
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
	rows, err := r.conn.Query(categoryQuery, postID)
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

func (r *PostRepository) ListPosts() ([]models.Post, error) {
	query := `SELECT id, user_id, title, content, created_at FROM posts ORDER BY created_at DESC`

	rows, err := r.conn.Query(query)
	if err != nil {
		return nil, err
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
			return nil, err
		}

		// Fetch categories for each post
		categoryQuery := `SELECT category_id FROM post_categories WHERE post_id = ?`
		categoryRows, err := r.conn.Query(categoryQuery, post.ID)
		if err != nil {
			return nil, err
		}
		defer categoryRows.Close()

		var categoryIDs []int64
		for categoryRows.Next() {
			var categoryID int64
			if err := categoryRows.Scan(&categoryID); err != nil {
				return nil, err
			}
			categoryIDs = append(categoryIDs, categoryID)
		}
		post.Categories = categoryIDs

		posts = append(posts, post)
	}

	return posts, nil
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
		JOIN categories c ON p.category_id = c.id
		WHERE 1=1
	`

	// Add conditions based on filter parameters
	if filters.CategoryID != nil {
		conditions = append(conditions, "p.category_id = ?")
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
