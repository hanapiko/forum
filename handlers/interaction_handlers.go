package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"forum/middleware"
	"forum/models"
	"forum/repository"
)

type InteractionHandler struct {
	interactionRepo repository.InteractionRepositoryInterface
	authMiddleware  *middleware.AuthMiddleware
}

func NewInteractionHandler(
	interactionRepo repository.InteractionRepositoryInterface,
	authMiddleware *middleware.AuthMiddleware,
) *InteractionHandler {
	return &InteractionHandler{
		interactionRepo: interactionRepo,
		authMiddleware:  authMiddleware,
	}
}

type InteractionRequest struct {
	EntityID   int64  `json:"entity_id"`
	EntityType string `json:"entity_type"`
}

type InteractionResponse struct {
	Likes    int `json:"likes"`
	Dislikes int `json:"dislikes"`
}

func (h *InteractionHandler) LikeEntity(w http.ResponseWriter, r *http.Request) {
	// Check if user is logged in
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized: Login required", http.StatusUnauthorized)
		return
	}

	var request InteractionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validate request
	if request.EntityID <= 0 || request.EntityType == "" {
		http.Error(w, "Invalid entity details", http.StatusBadRequest)
		return
	}

	// Validate entity type
	if request.EntityType != "post" && request.EntityType != "comment" {
		http.Error(w, "Invalid entity type", http.StatusBadRequest)
		return
	}

	// Add like interaction
	err := h.interactionRepo.AddInteraction(
		int64(userID),
		request.EntityID,
		request.EntityType,
		models.Like,
	)
	if err != nil {
		log.Printf("Failed to add like: %v", err)
		http.Error(w, "Failed to add like", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *InteractionHandler) DislikeEntity(w http.ResponseWriter, r *http.Request) {
	// Check if user is logged in
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized: Login required", http.StatusUnauthorized)
		return
	}

	var request InteractionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validate request
	if request.EntityID <= 0 || request.EntityType == "" {
		http.Error(w, "Invalid entity details", http.StatusBadRequest)
		return
	}

	// Validate entity type
	if request.EntityType != "post" && request.EntityType != "comment" {
		http.Error(w, "Invalid entity type", http.StatusBadRequest)
		return
	}

	// Add dislike interaction
	err := h.interactionRepo.AddInteraction(
		int64(userID),
		request.EntityID,
		request.EntityType,
		models.Dislike,
	)
	if err != nil {
		log.Printf("Failed to add dislike: %v", err)
		http.Error(w, "Failed to add dislike", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *InteractionHandler) GetInteractionCounts(w http.ResponseWriter, r *http.Request) {
	// Parse entity details from query parameters
	entityID := r.URL.Query().Get("entity_id")
	entityType := r.URL.Query().Get("entity_type")

	// Validate input
	if entityID == "" || entityType == "" {
		http.Error(w, "Missing entity details", http.StatusBadRequest)
		return
	}

	// Validate entity type
	if entityType != "post" && entityType != "comment" {
		http.Error(w, "Invalid entity type", http.StatusBadRequest)
		return
	}

	// Convert entityID to int64
	parsedEntityID, err := strconv.ParseInt(entityID, 10, 64)
	if err != nil {
		http.Error(w, "Invalid entity ID", http.StatusBadRequest)
		return
	}

	// Get interaction counts
	likes, dislikes, err := h.interactionRepo.GetInteractionCounts(parsedEntityID, entityType)
	if err != nil {
		log.Printf("Failed to retrieve interaction counts: %v", err)
		http.Error(w, "Failed to retrieve interaction counts", http.StatusInternalServerError)
		return
	}

	// Respond with counts
	w.Header().Set("Content-Type", "application/json")
	response := InteractionResponse{
		Likes:    likes,
		Dislikes: dislikes,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
