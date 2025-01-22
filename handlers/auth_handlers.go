package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"forum/middleware"
	"forum/models"
	"forum/repository"
)

type AuthHandler struct {
	userRepo       repository.UserRepository
	authMiddleware middleware.AuthMiddlewareInterface
	renderer       *TemplateRenderer
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token   string `json:"token"`
	UserID  int    `json:"user_id"`
	Message string `json:"message"`
}

func NewAuthHandler(userRepo repository.UserRepository, authMiddleware middleware.AuthMiddlewareInterface, renderer *TemplateRenderer) *AuthHandler {
	return &AuthHandler{
		userRepo:       userRepo,
		authMiddleware: authMiddleware,
		renderer:       renderer,
	}
}

// Input Validation
func validateRegisterRequest(req map[string]string) error {
	// Trim whitespace
	username := strings.TrimSpace(req["username"])
	email := strings.TrimSpace(req["email"])
	password := strings.TrimSpace(req["password"])

	// Username validation
	if username == "" {
		return errors.New("username is required")
	}
	if len(username) < 3 || len(username) > 50 {
		return errors.New("username must be between 3 and 50 characters")
	}
	if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(username) {
		return errors.New("username can only contain letters, numbers, and underscores")
	}

	// Email validation
	if email == "" {
		return errors.New("email is required")
	}
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}

	// Password validation
	if password == "" {
		return errors.New("password is required")
	}
	if len(password) < 8 || len(password) > 72 {
		return errors.New("password must be between 8 and 72 characters")
	}
	if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !regexp.MustCompile(`[a-z]`).MatchString(password) {
		return errors.New("password must contain at least one lowercase letter")
	}
	if !regexp.MustCompile(`[0-9]`).MatchString(password) {
		return errors.New("password must contain at least one number")
	}
	if !regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password) {
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

func (h *AuthHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	// Prepare template data
	data := TemplateData{
		PageTitle: "Login",
	}

	// If there's a login error from previous attempt, pass it
	if loginError := r.URL.Query().Get("error"); loginError != "" {
		data.Error = "Invalid login credentials"
	}

	// Render the login page
	err := h.renderer.Render(w, r, "login.html", data)
	if err != nil {
		log.Printf("Error rendering login page: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	if err := r.ParseForm(); err != nil {
		h.renderer.RenderError(w, r, http.StatusBadRequest, err)
		return
	}

	loginRequest := &LoginRequest{
		Email:    r.Form.Get("email"),
		Password: r.Form.Get("password"),
	}

	// Validate login request
	if err := validateLoginRequest(loginRequest); err != nil {
		http.Redirect(w, r, "/login?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}

	// Attempt authentication
	user, err := h.userRepo.Authenticate(loginRequest.Email, loginRequest.Password)
	if err != nil {
		// Redirect back to login with error
		http.Redirect(w, r, "/login?error=invalid_credentials", http.StatusSeeOther)
		return
	}

	// Generate JWT token
	token, err := h.authMiddleware.GenerateToken(user.ID, user.Username)
	if err != nil {
		log.Printf("Token generation error: %v", err)
		h.renderer.RenderError(w, r, http.StatusInternalServerError, err)
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

func (h *AuthHandler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	// Prepare template data
	data := TemplateData{
		PageTitle: "Register",
	}

	// If there's a registration error from previous attempt, pass it
	if regError := r.URL.Query().Get("error"); regError != "" {
		data.Error = "Registration failed. Please try again."
	}

	// Render register template
	err := h.renderer.Render(w, r, "auth/register", data)
	if err != nil {
		h.renderer.RenderError(w, r, http.StatusInternalServerError, err)
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	if err := r.ParseForm(); err != nil {
		h.renderer.RenderError(w, r, http.StatusBadRequest, err)
		return
	}

	// Collect registration data
	username := r.Form.Get("username")
	email := r.Form.Get("email")
	password := r.Form.Get("password")

	// Validate input
	if err := validateRegisterRequest(map[string]string{
		"username": username,
		"email":    email,
		"password": password,
	}); err != nil {
		http.Redirect(w, r, "/register?error=registration_failed", http.StatusSeeOther)
		return
	}

	// Create user
	user := &models.User{
		Username: username,
		Email:    email,
		Password: password,
	}

	user, err := h.userRepo.Create(user)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			http.Redirect(w, r, "/register?error=username_or_email_exists", http.StatusSeeOther)
			return
		}
		log.Printf("User creation error: %v", err)
		h.renderer.RenderError(w, r, http.StatusInternalServerError, err)
		return
	}

	// Generate JWT token
	token, err := h.authMiddleware.GenerateToken(user.ID, user.Username)
	if err != nil {
		log.Printf("Token generation error: %v", err)
		h.renderer.RenderError(w, r, http.StatusInternalServerError, err)
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

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Implement logout logic
	// Clear session, remove tokens, etc.

	// Redirect to login page
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *AuthHandler) AuthStatus(w http.ResponseWriter, r *http.Request) {
	// Get the user from the context (assuming it was set by the auth middleware)
	user, ok := r.Context().Value("user").(*models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Prepare the authentication status response
	response := AuthResponse{
		UserID:  user.ID,
		Message: "User is authenticated",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	// Simple home route handler
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Welcome to the Forum API",
		"status":  "healthy",
	})
}
