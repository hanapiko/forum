package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"forum/db"
	"forum/middleware"
	"forum/repository"
)

type AuthHandler struct {
	userRepo       *repository.UserRepository
	authMiddleware *middleware.AuthMiddleware
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token   string `json:"token"`
	UserID  int64  `json:"user_id"`
	Message string `json:"message"`
}

func NewAuthHandler(userRepo *repository.UserRepository, authMiddleware *middleware.AuthMiddleware) *AuthHandler {
	return &AuthHandler{
		userRepo:       userRepo,
		authMiddleware: authMiddleware,
	}
}

// Input Validation
func validateRegisterRequest(req *RegisterRequest) error {
	// Basic validation
	if len(strings.TrimSpace(req.Username)) < 3 {
		return errors.New("username must be at least 3 characters long")
	}
	if len(strings.TrimSpace(req.Password)) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	if !strings.Contains(req.Email, "@") {
		return errors.New("invalid email format")
	}
	return nil
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// Validate HTTP method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if err := validateRegisterRequest(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create user
	user := &db.User{
		Username: req.Username,
		Email:    req.Email,
	}

	err = h.userRepo.Create(user, req.Password)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			http.Error(w, "Username or email already exists", http.StatusConflict)
			return
		}
		log.Printf("User creation error: %v", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Generate JWT token
	token, err := h.authMiddleware.GenerateToken(user)
	if err != nil {
		log.Printf("Token generation error: %v", err)
		http.Error(w, "Failed to generate authentication token", http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := AuthResponse{
		Token:   token,
		UserID:  user.ID,
		Message: "User registered successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// Validate HTTP method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Authenticate user
	user, err := h.userRepo.Authenticate(req.Email, req.Password)
	if err != nil {
		log.Printf("Authentication error: %v", err)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate JWT token
	token, err := h.authMiddleware.GenerateToken(user)
	if err != nil {
		log.Printf("Token generation error: %v", err)
		http.Error(w, "Failed to generate authentication token", http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := AuthResponse{
		Token:   token,
		UserID:  user.ID,
		Message: "Login successful",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// For JWT, logout is typically handled client-side by removing the token
	// Here we can add additional logout logic if needed (e.g., token blacklisting)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logout successful",
	})
}
