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

	"github.com/go-chi/chi/v5"
)

type PostHandler struct {
	postRepo       repository.PostRepositoryInterface
	authMiddleware middleware.AuthMiddlewareInterface
	renderer       *TemplateRenderer
}

func NewPostHandler(postRepo repository.PostRepositoryInterface, authMiddleware middleware.AuthMiddlewareInterface, renderer *TemplateRenderer) *PostHandler {
	return &PostHandler{
		postRepo:       postRepo,
		authMiddleware: authMiddleware,
		renderer:       renderer,
	}
}

type PostRequest struct {
	Categories []int64 `json:"categories"`
	Title      string  `json:"title"`
	Content    string  `json:"content"`
}

type CreatePostRequest struct {
	Title      string `json:"title"`
	Content    string `json:"content"`
	CategoryID int    `json:"category_id"`
}

func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	// Check if user is authenticated
	userID, ok := h.authMiddleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req CreatePostRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validate input
	if req.Title == "" || req.Content == "" {
		http.Error(w, "Title and content are required", http.StatusBadRequest)
		return
	}

	// Create post
	newPost := &models.Post{
		UserID:     int64(userID),
		Title:      req.Title,
		Content:    req.Content,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Categories: []int64{int64(req.CategoryID)},
	}

	// Save post to database
	err = h.postRepo.Create(newPost, []int64{int64(req.CategoryID)})
	if err != nil {
		log.Printf("Error creating post: %v", err)
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}

	// Prepare response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newPost)
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

	// Calculate total pages: ceiling division of total count by limit
	totalPages := (totalCount + limit - 1) / limit

	// Convert posts slice to slice of pointers
	postPtrs := make([]*models.Post, len(posts))
	for i := range posts {
		postPtrs[i] = &posts[i]
	}

	// Prepare paginated response
	response := struct {
		Posts       []*models.Post `json:"posts"`
		TotalCount  int            `json:"total_count"`
		CurrentPage int            `json:"current_page"`
		TotalPages  int            `json:"total_pages"`
	}{
		Posts:       postPtrs,
		TotalCount:  totalCount,
		CurrentPage: page,
		TotalPages:  totalPages,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *PostHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	// Extract post ID
	postIDStr := chi.URLParam(r, "id")
	if postIDStr == "" {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Retrieve post with details
	post, err := h.postRepo.GetByID(strconv.FormatInt(postID, 10))
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
	userID, ok := h.authMiddleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req CreatePostRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validate input
	if req.Title == "" || req.Content == "" {
		http.Error(w, "Title and content are required", http.StatusBadRequest)
		return
	}

	// Extract post ID
	postIDStr := chi.URLParam(r, "id")
	if postIDStr == "" {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Retrieve post with details
	post, err := h.postRepo.GetByID(strconv.FormatInt(postID, 10))
	if err != nil {
		log.Printf("Error retrieving post: %v", err)
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	// Check if user is authorized
	if post.UserID != int64(userID) {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Update post
	updatedPost := &models.Post{
		ID:         postID,
		Title:      req.Title,
		Content:    req.Content,
		Categories: []int64{int64(req.CategoryID)},
		UpdatedAt:  time.Now(),
		UserID:     post.UserID, // Add this line to ensure user ownership check
	}

	err = h.postRepo.UpdatePost(updatedPost, []int64{int64(req.CategoryID)}, post.UserID)
	if err != nil {
		log.Printf("Error updating post: %v", err)
		http.Error(w, "Failed to update post", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Post updated successfully"})
}

func (h *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	// Check user authentication
	userID, ok := h.authMiddleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract post ID
	postIDStr := chi.URLParam(r, "id")
	if postIDStr == "" {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Retrieve post with details
	post, err := h.postRepo.GetByID(strconv.FormatInt(postID, 10))
	if err != nil || post.UserID != int64(userID) {
		http.Error(w, "Unauthorized or post not found", http.StatusForbidden)
		return
	}

	// Delete post
	err = h.postRepo.DeletePost(postIDStr, int64(userID))
	if err != nil {
		log.Printf("Error deleting post: %v", err)
		http.Error(w, "Failed to delete post", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Post deleted successfully"})
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

func (h *PostHandler) CreatePostPage(w http.ResponseWriter, r *http.Request) {
	// Check if user is authenticated
	userID, ok := h.authMiddleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Prepare data for the create post page
	data := TemplateData{
		PageTitle: "Create Post",
		User:      userID,
		Data:      map[string]interface{}{"UserID": userID},
	}

	// Render the create post page template
	err := h.renderer.Render(w, r, "posts/create", data)
	if err != nil {
		log.Printf("Error rendering create post page: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
