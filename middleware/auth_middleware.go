package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"forum/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthMiddlewareInterface interface {
	GenerateToken(userID int, username string) (string, error)
	ValidateToken(tokenString string) (*Claims, error)
	ProtectRoute(next http.Handler) http.Handler
	GetUserIDFromContext(ctx context.Context) (int, bool)
	IsAuthenticated(r *http.Request) bool
}

type AuthMiddleware struct {
	sessionRepo repository.SessionRepositoryInterface
	secretKey   []byte
	expiration  time.Duration
}

type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type contextKey string

var UserClaimsKey contextKey = "user_claims"

func NewAuthMiddleware(sessionRepo repository.SessionRepositoryInterface, secretKey string, expirationHours int) *AuthMiddleware {
	return &AuthMiddleware{
		sessionRepo: sessionRepo,
		secretKey:   []byte(secretKey),
		expiration:  time.Duration(expirationHours) * time.Hour,
	}
}

func (m *AuthMiddleware) GenerateToken(userID int, username string) (string, error) {
	return generateToken(m.secretKey, m.expiration, userID, username)
}

func (m *AuthMiddleware) ValidateToken(tokenString string) (*Claims, error) {
	return validateToken(m.secretKey, tokenString)
}

func (m *AuthMiddleware) ProtectRoute(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header or cookie
		tokenString, err := m.extractToken(r)
		if err != nil {
			http.Error(w, "Failed to extract token", http.StatusInternalServerError)
			return
		}
		if tokenString == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		claims, err := m.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Add user info to request context
		ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *AuthMiddleware) extractToken(r *http.Request) (string, error) {
	// Check Authorization header
	bearerToken := r.Header.Get("Authorization")
	if len(strings.Split(bearerToken, " ")) == 2 {
		return strings.Split(bearerToken, " ")[1], nil
	}

	// Check cookie
	cookie, err := r.Cookie("token")
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// HashPassword securely hashes a password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPasswordHash compares a password with its hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GetUserIDFromContext retrieves the user ID from the context
func GetUserIDFromContext(ctx context.Context) (int, bool) {
	claims, ok := ctx.Value(UserClaimsKey).(*Claims)
	if !ok {
		return 0, false
	}
	return claims.UserID, true
}

func (m *AuthMiddleware) GetUserIDFromContext(ctx context.Context) (int, bool) {
	return GetUserIDFromContext(ctx)
}

func generateToken(secretKey []byte, expiration time.Duration, userID int, username string) (string, error) {
	expirationTime := time.Now().Add(expiration)
	claims := &Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "forum",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

func validateToken(secretKey []byte, tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func (m *AuthMiddleware) IsAuthenticated(r *http.Request) bool {
	// Extract token from the request
	tokenString, err := m.extractToken(r)
	if err != nil {
		return false
	}

	// Validate the token
	_, err = m.ValidateToken(tokenString)
	return err == nil
}
