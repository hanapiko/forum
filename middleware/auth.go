package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"forum/models"
	"forum/repository"

	"github.com/golang-jwt/jwt/v5"
)

type AuthMiddleware struct {
	sessionRepo *repository.SessionRepository
	secretKey   []byte
}

func NewAuthMiddleware(sessionRepo *repository.SessionRepository, secretKey string) *AuthMiddleware {
	return &AuthMiddleware{
		sessionRepo: sessionRepo,
		secretKey:   []byte(secretKey),
	}
}

// GenerateToken creates a JWT token for the user
func (m *AuthMiddleware) GenerateToken(user *models.User) (string, error) {
	// Create a new session for the user
	session, err := m.sessionRepo.CreateSession(user.ID)
	if err != nil {
		return "", fmt.Errorf("failed to create session: %v", err)
	}

	// Create JWT claims
	claims := jwt.MapClaims{
		"user_id":    user.ID,
		"session_id": session.UUID,
		"exp":        time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString(m.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %v", err)
	}

	return tokenString, nil
}

// Authenticate checks if a user is logged in and adds user context
func (m *AuthMiddleware) Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// No token, continue as anonymous
			ctx := context.WithValue(r.Context(), "user_id", nil)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Extract token from header
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return m.secretKey, nil
		})
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		// Extract user ID and session ID
		userID, ok := claims["user_id"].(float64)
		if !ok {
			http.Error(w, "Invalid user ID in token", http.StatusUnauthorized)
			return
		}

		sessionID, ok := claims["session_id"].(string)
		if !ok {
			http.Error(w, "Invalid session ID in token", http.StatusUnauthorized)
			return
		}

		// Validate session
		_, err = m.sessionRepo.ValidateSession(sessionID)
		if err != nil {
			http.Error(w, "Invalid session", http.StatusUnauthorized)
			return
		}

		// Add user ID to context
		ctx := context.WithValue(r.Context(), "user_id", int64(userID))
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// ProtectRoute ensures only authenticated users can access the route
func (m *AuthMiddleware) ProtectRoute(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing authorization token", http.StatusUnauthorized)
			return
		}

		// Extract token from header
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return m.secretKey, nil
		})
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		// Extract user ID and session ID
		userID, ok := claims["user_id"].(float64)
		if !ok {
			http.Error(w, "Invalid user ID in token", http.StatusUnauthorized)
			return
		}

		sessionID, ok := claims["session_id"].(string)
		if !ok {
			http.Error(w, "Invalid session ID in token", http.StatusUnauthorized)
			return
		}

		// Validate session
		_, err = m.sessionRepo.ValidateSession(sessionID)
		if err != nil {
			http.Error(w, "Invalid session", http.StatusUnauthorized)
			return
		}

		// Add user ID to context
		ctx := context.WithValue(r.Context(), "user_id", int64(userID))
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// IsLoggedIn checks if the user is logged in
func IsLoggedIn(ctx context.Context) bool {
	userID := ctx.Value("user_id")
	return userID != nil
}

// GetUserIDFromContext retrieves the user ID from the request context
func GetUserIDFromContext(ctx context.Context) (int64, bool) {
	userID := ctx.Value("user_id")
	if userID == nil {
		return 0, false
	}
	return userID.(int64), true
}

// SecretKey returns the secret key used for token signing
func (m *AuthMiddleware) SecretKey() string {
	return string(m.secretKey)
}

// SessionRepo returns the session repository
func (m *AuthMiddleware) SessionRepo() *repository.SessionRepository {
	return m.sessionRepo
}
