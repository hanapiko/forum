package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"forum/middleware"
	"forum/repository"

	"github.com/go-chi/chi/v5"
)

type CommentHandler struct {
	commentRepo    repository.CommentRepositoryInterface
	authMiddleware *middleware.AuthMiddleware
}

func NewCommentHandler(commentRepo repository.CommentRepositoryInterface, authMiddleware *middleware.AuthMiddleware) *CommentHandler {
	return &CommentHandler{
		commentRepo:    commentRepo,
		authMiddleware: authMiddleware,
	}
}

func (h *CommentHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	// Check if user is authenticated
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get post ID from URL
	postIDStr := chi.URLParam(r, "postId")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req struct {
		Content string `json:"content"`
	}
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validate input
	if req.Content == "" {
		http.Error(w, "Comment content is required", http.StatusBadRequest)
		return
	}

	// Create comment
	newComment := &repository.Comment{
		PostID:  postID,
		UserID:  int64(userID),
		Content: req.Content,
	}

	// Save comment to database
	err = h.commentRepo.Create(newComment)
	if err != nil {
		log.Printf("Error creating comment: %v", err)
		http.Error(w, "Failed to create comment", http.StatusInternalServerError)
		return
	}

	// Prepare response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newComment)
}

func (h *CommentHandler) GetComments(w http.ResponseWriter, r *http.Request) {
	// Get post ID from URL
	postIDStr := chi.URLParam(r, "postId")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Retrieve comments for the post
	comments, err := h.commentRepo.GetByPostID(postID)
	if err != nil {
		log.Printf("Error retrieving comments: %v", err)
		http.Error(w, "Failed to retrieve comments", http.StatusInternalServerError)
		return
	}

	// Prepare response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(comments)
}
