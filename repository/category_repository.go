package repository

import (
	"database/sql"
	"errors"

	"forum/models"
)

type CategoryRepositoryInterface interface {
	Create(category *models.Category) error
	GetByID(categoryID int64) (*models.Category, error)
	ListCategories() ([]models.Category, error)
	Update(category *models.Category) error
	Delete(categoryID int64) error
}

type CategoryRepository struct {
	conn *sql.DB
}

var _ CategoryRepositoryInterface = &CategoryRepository{}

func NewCategoryRepository(conn *sql.DB) *CategoryRepository {
	return &CategoryRepository{conn: conn}
}

// Create a new category
func (r *CategoryRepository) Create(category *models.Category) error {
	// Validate input
	if category.Name == "" {
		return errors.New("category name cannot be empty")
	}

	query := `INSERT INTO categories (name, description) VALUES (?, ?)`
	result, err := r.conn.Exec(query, category.Name, category.Description)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	category.ID = id
	return nil
}

// Get a category by ID
func (r *CategoryRepository) GetByID(categoryID int64) (*models.Category, error) {
	query := `SELECT id, name, description FROM categories WHERE id = ?`
	category := &models.Category{}
	err := r.conn.QueryRow(query, categoryID).Scan(
		&category.ID,
		&category.Name,
		&category.Description,
	)
	if err != nil {
		return nil, err
	}
	return category, nil
}

// List all categories
func (r *CategoryRepository) ListCategories() ([]models.Category, error) {
	query := `SELECT id, name, description FROM categories ORDER BY name`
	rows, err := r.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var category models.Category
		if err := rows.Scan(&category.ID, &category.Name, &category.Description); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	// Check for any errors encountered during iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

// Update an existing category
func (r *CategoryRepository) Update(category *models.Category) error {
	if category.ID == 0 {
		return errors.New("category ID is required for update")
	}

	query := `UPDATE categories SET name = ?, description = ? WHERE id = ?`
	_, err := r.conn.Exec(query, category.Name, category.Description, category.ID)
	return err
}

// Delete a category
func (r *CategoryRepository) Delete(categoryID int64) error {
	// First, check if the category is used in any posts
	checkQuery := `SELECT COUNT(*) FROM post_categories WHERE category_id = ?`
	var count int
	err := r.conn.QueryRow(checkQuery, categoryID).Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return errors.New("cannot delete category: it is used in existing posts")
	}

	// If no posts use this category, proceed with deletion
	query := `DELETE FROM categories WHERE id = ?`
	_, err = r.conn.Exec(query, categoryID)
	return err
}
