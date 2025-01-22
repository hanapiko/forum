package repository

import (
	"database/sql"
	"strconv"
	"testing"

	"forum/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	// Import the sqlite3 driver
	_ "github.com/mattn/go-sqlite3"
)

func setupTestUser(t *testing.T, db *sql.DB) int64 {
	// Insert a test user
	result, err := db.Exec(
		"INSERT INTO users (username, email, password) VALUES (?, ?, ?)",
		"testuser", "test@example.com", "hashedpassword",
	)
	require.NoError(t, err)

	userID, err := result.LastInsertId()
	require.NoError(t, err)

	return userID
}

func setupTestCategories(t *testing.T, db *sql.DB) []int64 {
	// Insert test categories
	categories := []string{"Technology", "Sports", "Music"}
	var categoryIDs []int64

	for _, name := range categories {
		result, err := db.Exec("INSERT INTO categories (name) VALUES (?)", name)
		require.NoError(t, err)

		categoryID, err := result.LastInsertId()
		require.NoError(t, err)

		categoryIDs = append(categoryIDs, categoryID)
	}

	return categoryIDs
}

func TestPostRepository_Create(t *testing.T) {
	// Setup test database
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	repo := NewPostRepository(db)

	// Create a test post
	post := &models.Post{
		UserID:  1,
		Title:   "Test Post",
		Content: "Testing Create method",
	}

	err := repo.Create(post, []int64{1, 2})
	assert.NoError(t, err)
	assert.NotZero(t, post.ID)

	// Verify post was created
	createdPost, err := repo.GetByID(strconv.FormatInt(post.ID, 10))
	assert.NoError(t, err)
	assert.Equal(t, post.Title, createdPost.Title)
	assert.Equal(t, post.Content, createdPost.Content)
	assert.Equal(t, 2, len(createdPost.Categories))
}

func TestPostRepository_GetByID(t *testing.T) {
	// Setup test database
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	repo := NewPostRepository(db)

	// Create a test post
	post := &models.Post{
		UserID:  1,
		Title:   "Test Post",
		Content: "Testing GetByID method",
	}
	err := repo.Create(post, []int64{1})
	assert.NoError(t, err)

	// Convert ID to string for testing
	postIDStr := strconv.FormatInt(post.ID, 10)

	// Retrieve the post
	retrievedPost, err := repo.GetByID(postIDStr)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedPost)
	assert.Equal(t, post.Title, retrievedPost.Title)
	assert.Equal(t, post.Content, retrievedPost.Content)
}

func TestPostRepository_ListPosts(t *testing.T) {
	// Setup test database
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	// Setup test user and categories
	userID := setupTestUser(t, db)
	categoryIDs := setupTestCategories(t, db)

	// Create repository
	repo := NewPostRepository(db)

	// Create multiple test posts
	posts := []*models.Post{
		{
			UserID:  userID,
			Title:   "Test Post 1",
			Content: "Content 1",
		},
		{
			UserID:  userID,
			Title:   "Test Post 2",
			Content: "Content 2",
		},
	}

	for _, post := range posts {
		err := repo.Create(post, []int64{categoryIDs[0]})
		assert.NoError(t, err)
	}

	// List posts
	listedPosts, _, err := repo.ListPosts(1, 10)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(listedPosts), 2)
}

func TestPostRepository_UpdatePost(t *testing.T) {
	// Setup test database
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	repo := NewPostRepository(db)

	// Create a test post
	post := &models.Post{
		UserID:  1,
		Title:   "Original Title",
		Content: "Original Content",
	}

	err := repo.Create(post, []int64{1})
	assert.NoError(t, err)

	// Update the post
	updatedPost := &models.Post{
		ID:      post.ID,
		UserID:  1,
		Title:   "Updated Title",
		Content: "Updated Content",
	}

	err = repo.UpdatePost(updatedPost, []int64{2}, post.UserID)
	assert.NoError(t, err)

	// Retrieve and verify the updated post
	retrievedPost, err := repo.GetByID(strconv.FormatInt(post.ID, 10))
	assert.NoError(t, err)
	assert.Equal(t, "Updated Title", retrievedPost.Title)
	assert.Equal(t, "Updated Content", retrievedPost.Content)
	assert.Equal(t, 1, len(retrievedPost.Categories))
}

func TestPostRepository_DeletePost(t *testing.T) {
	// Setup test database
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	repo := NewPostRepository(db)

	// Create a test post
	post := &models.Post{
		UserID:  1,
		Title:   "Post to Delete",
		Content: "This post will be deleted",
	}

	err := repo.Create(post, []int64{1})
	assert.NoError(t, err)

	// Delete the post
	err = repo.DeletePost(strconv.FormatInt(post.ID, 10), post.UserID)
	assert.NoError(t, err)

	// Try to retrieve the deleted post
	_, err = repo.GetByID(strconv.FormatInt(post.ID, 10))
	assert.Error(t, err)
}

func TestPostRepository_GetUserPosts(t *testing.T) {
	// Setup test database
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	// Setup test user
	userID := setupTestUser(t, db)

	repo := NewPostRepository(db)

	// Create multiple posts for the user
	posts := []*models.Post{
		{
			UserID:  userID,
			Title:   "User Post 1",
			Content: "Content 1",
		},
		{
			UserID:  userID,
			Title:   "User Post 2",
			Content: "Content 2",
		},
	}

	for _, post := range posts {
		err := repo.Create(post, []int64{1})
		assert.NoError(t, err)
	}

	// Get user posts
	userPosts, err := repo.GetUserPosts(userID)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(userPosts))
}

func TestPostRepository_FilterPosts(t *testing.T) {
	// Setup test database
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	// Setup test user and categories
	userID := setupTestUser(t, db)
	categoryIDs := setupTestCategories(t, db)

	repo := NewPostRepository(db)

	// Create multiple posts with different categories
	posts := []*models.Post{
		{
			UserID:  userID,
			Title:   "Tech Post",
			Content: "Technology content",
		},
		{
			UserID:  userID,
			Title:   "Sports Post",
			Content: "Sports content",
		},
	}

	err := repo.Create(posts[0], []int64{categoryIDs[0]}) // Tech category
	assert.NoError(t, err)
	err = repo.Create(posts[1], []int64{categoryIDs[1]}) // Sports category
	assert.NoError(t, err)

	// Filter posts by category
	categoryID, _ := strconv.ParseInt(strconv.FormatInt(categoryIDs[0], 10), 10, 64)
	filteredPosts, err := repo.FilterPosts(struct {
		CategoryID  *int64
		UserID      *int64
		LikedByUser *int64
	}{
		CategoryID: &categoryID,
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(filteredPosts))
	assert.Equal(t, "Tech Post", filteredPosts[0].Title)
}
