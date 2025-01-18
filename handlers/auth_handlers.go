package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"regexp"
	"strings"

	"forum/middleware"
	"forum/models"
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
	// Trim whitespace
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)
	req.Password = strings.TrimSpace(req.Password)

	// Username validation
	if req.Username == "" {
		return errors.New("username is required")
	}
	if len(req.Username) < 3 || len(req.Username) > 50 {
		return errors.New("username must be between 3 and 50 characters")
	}
	if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(req.Username) {
		return errors.New("username can only contain letters, numbers, and underscores")
	}

	// Email validation
	if req.Email == "" {
		return errors.New("email is required")
	}
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
	if !emailRegex.MatchString(req.Email) {
		return errors.New("invalid email format")
	}

	// Password validation
	if req.Password == "" {
		return errors.New("password is required")
	}
	if len(req.Password) < 8 || len(req.Password) > 72 {
		return errors.New("password must be between 8 and 72 characters")
	}
	if !regexp.MustCompile(`[A-Z]`).MatchString(req.Password) {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !regexp.MustCompile(`[a-z]`).MatchString(req.Password) {
		return errors.New("password must contain at least one lowercase letter")
	}
	if !regexp.MustCompile(`[0-9]`).MatchString(req.Password) {
		return errors.New("password must contain at least one number")
	}
	if !regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(req.Password) {
		return errors.New("password must contain at least one special character")
	}

	return nil
}

// Input Validation for Login Request
func validateLoginRequest(req *LoginRequest) error {
	// Trim whitespace
	req.Email = strings.TrimSpace(req.Email)
	req.Password = strings.TrimSpace(req.Password)

	// Email validation
	if req.Email == "" {
		return errors.New("email is required")
	}
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
	if !emailRegex.MatchString(req.Email) {
		return errors.New("invalid email format")
	}

	// Password validation
	if req.Password == "" {
		return errors.New("password is required")
	}
	if len(req.Password) < 8 || len(req.Password) > 72 {
		return errors.New("password must be between 8 and 72 characters")
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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validate input
	if err := validateRegisterRequest(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create user
	user := &models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	}

	err := h.userRepo.Create(user)
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
	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode registration response: %v", err)
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// Validate HTTP method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validate input
	if err := validateLoginRequest(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode login response: %v", err)
	}
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Validate HTTP method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is authenticated
	_, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// For JWT, logout is typically handled client-side by removing the token
	// Here we can add additional logout logic if needed (e.g., token blacklisting)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"message": "Logout successful",
	}); err != nil {
		log.Printf("Failed to encode logout response: %v", err)
	}
}
