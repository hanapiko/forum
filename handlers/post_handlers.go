package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"forum/middleware"
	"forum/models"
	"forum/repository"
)

type PostHandler struct {
	postRepo       *repository.PostRepository
	authMiddleware *middleware.AuthMiddleware
}

func NewPostHandler(postRepo *repository.PostRepository, authMiddleware *middleware.AuthMiddleware) *PostHandler {
	return &PostHandler{
		postRepo:       postRepo,
		authMiddleware: authMiddleware,
	}
}

type PostRequest struct {
	CategoryID int64  `json:"category_id"`
	Title      string `json:"title"`
	Content    string `json:"content"`
}

func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	// Check if user is authenticated
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var postData PostRequest
	if err := json.NewDecoder(r.Body).Decode(&postData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validate input
	if postData.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}
	if postData.Content == "" {
		http.Error(w, "Content is required", http.StatusBadRequest)
		return
	}

	// Create post
	newPost := &models.Post{
		UserID:     userID,
		CategoryID: postData.CategoryID,
		Title:      postData.Title,
		Content:    postData.Content,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Save post to database
	createdPost, err := h.postRepo.CreatePost(newPost)
	if err != nil {
		log.Printf("Error creating post: %v", err)
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}

	// Prepare response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdPost)
}

func (h *PostHandler) ListPosts(w http.ResponseWriter, r *http.Request) {
	// Optional query parameters for pagination
	page := 1
	limit := 10

	pageParam := r.URL.Query().Get("page")
	limitParam := r.URL.Query().Get("limit")

	if pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
			page = p
		}
	}

	if limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Retrieve posts
	posts, totalCount, err := h.postRepo.ListPosts(page, limit)
	if err != nil {
		log.Printf("Error listing posts: %v", err)
		http.Error(w, "Failed to retrieve posts", http.StatusInternalServerError)
		return
	}

	// Prepare paginated response
	response := struct {
		Posts       []*models.Post `json:"posts"`
		TotalCount  int            `json:"total_count"`
		CurrentPage int            `json:"current_page"`
		TotalPages  int            `json:"total_pages"`
	}{
		Posts:       posts,
		TotalCount:  totalCount,
		CurrentPage: page,
		TotalPages:  (totalCount + limit - 1) / limit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *PostHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	// Extract post ID from URL
	postIDStr := r.URL.Query().Get("id")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Retrieve post with details
	post, err := h.postRepo.GetPostByID(postID)
	if err != nil {
		log.Printf("Error retrieving post: %v", err)
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

func (h *PostHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	// Check user authentication
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var updateData PostRequest
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Extract post ID
	postIDStr := r.URL.Query().Get("id")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Update post
	updatedPost, err := h.postRepo.UpdatePost(&models.Post{
		ID:         postID,
		UserID:     userID,
		Title:      updateData.Title,
		Content:    updateData.Content,
		CategoryID: updateData.CategoryID,
		UpdatedAt:  time.Now(),
	})
	if err != nil {
		log.Printf("Error updating post: %v", err)
		http.Error(w, "Failed to update post", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedPost)
}

func (h *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	// Check user authentication
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract post ID
	postIDStr := r.URL.Query().Get("id")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Delete post
	err = h.postRepo.DeletePost(postID, userID)
	if err != nil {
		log.Printf("Error deleting post: %v", err)
		http.Error(w, "Failed to delete post", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *PostHandler) FilterPosts(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters with validation
	categoryIDStr := r.URL.Query().Get("category_id")
	userIDStr := r.URL.Query().Get("user_id")
	likedByUserStr := r.URL.Query().Get("liked_by_user")

	var categoryID, userID, likedByUser *int64

	if categoryIDStr != "" {
		parsedCategoryID, err := strconv.ParseInt(categoryIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid category ID", http.StatusBadRequest)
			return
		}
		categoryID = &parsedCategoryID
	}

	if userIDStr != "" {
		parsedUserID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}
		userID = &parsedUserID
	}

	if likedByUserStr != "" {
		parsedLikedByUser, err := strconv.ParseInt(likedByUserStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid liked by user ID", http.StatusBadRequest)
			return
		}
		likedByUser = &parsedLikedByUser
	}

	// Filter posts
	posts, err := h.postRepo.FilterPosts(struct {
		CategoryID  *int64
		UserID      *int64
		LikedByUser *int64
	}{
		CategoryID:  categoryID,
		UserID:      userID,
		LikedByUser: likedByUser,
	})
	if err != nil {
		log.Printf("Failed to filter posts: %v", err)
		http.Error(w, "Failed to filter posts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(posts); err != nil {
		log.Printf("Failed to encode filtered posts: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
