package middleware

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

type MockAuthMiddleware struct {
	mock.Mock
}

func (m *MockAuthMiddleware) GenerateToken(userID int, username string) (string, error) {
	args := m.Called(userID, username)
	return args.String(0), args.Error(1)
}

func (m *MockAuthMiddleware) ValidateToken(tokenString string) (*Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Claims), args.Error(1)
}

func (m *MockAuthMiddleware) ProtectRoute(next http.Handler) http.Handler {
	args := m.Called(next)
	return args.Get(0).(http.Handler)
}
