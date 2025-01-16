package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"forum/middleware"
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

func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from token
	tokenString := r.Header.Get("Authorization")[7:] // Remove "Bearer "
	userID, err := h.authMiddleware.ExtractUserIDFromToken(tokenString)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var postData struct {
		CategoryID int64  `json:"category_id"`
		Title      string `json:"title"`
		Content    string `json:"content"`
	}
	err = json.NewDecoder(r.Body).Decode(&postData)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create post
	post := &repository.Post{
		UserID:     userID,
		CategoryID: postData.CategoryID,
		Title:      postData.Title,
		Content:    postData.Content,
	}
	err = h.postRepo.Create(post)
	if err != nil {
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}

	// Respond with created post
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(post)
}

func (h *PostHandler) ListPosts(w http.ResponseWriter, r *http.Request) {
	posts, err := h.postRepo.ListPosts()
	if err != nil {
		http.Error(w, "Failed to retrieve posts", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(posts)
}

func (h *PostHandler) FilterPosts(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	categoryID, _ := strconv.ParseInt(r.URL.Query().Get("category_id"), 10, 64)
	userID, _ := strconv.ParseInt(r.URL.Query().Get("user_id"), 10, 64)

	posts, err := h.postRepo.FilterPosts(categoryID, userID)
	if err != nil {
		http.Error(w, "Failed to filter posts", http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(posts)
}