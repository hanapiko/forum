package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"forum/db"
	"forum/repository"

	"github.com/golang-jwt/jwt"
)

type Claims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	jwt.StandardClaims
}

type AuthMiddleware struct {
	userRepo  *repository.UserRepository
	secretKey []byte
}

func NewAuthMiddleware(userRepo *repository.UserRepository, secretKey string) *AuthMiddleware {
	return &AuthMiddleware{
		userRepo:  userRepo,
		secretKey: []byte(secretKey),
	}
}

// GenerateToken creates a new JWT token for a user
func (am *AuthMiddleware) GenerateToken(user *db.User) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(am.secretKey)
}

// ValidateToken checks if a token is valid
func (am *AuthMiddleware) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return am.secretKey, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ProtectRoute is a middleware to protect routes that require authentication
func (am *AuthMiddleware) ProtectRoute(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized: No token provided", http.StatusUnauthorized)
			return
		}

		// Expected format: "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Unauthorized: Invalid token format", http.StatusUnauthorized)
			return
		}

		// Validate token
		_, err := am.ValidateToken(parts[1])
		if err != nil {
			http.Error(w, fmt.Sprintf("Unauthorized: %v", err), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}
}

// ExtractUserIDFromToken retrieves the user ID from a valid token
func (am *AuthMiddleware) ExtractUserIDFromToken(tokenString string) (int64, error) {
	claims, err := am.ValidateToken(tokenString)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}
