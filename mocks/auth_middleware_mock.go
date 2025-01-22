package mocks

import (
	"context"
	"net/http"

	"forum/middleware"

	"github.com/stretchr/testify/mock"
)

type MockAuthMiddleware struct {
	mock.Mock
}

func NewMockAuthMiddleware(t interface{}) *MockAuthMiddleware {
	return &MockAuthMiddleware{
		Mock: mock.Mock{},
	}
}

func (m *MockAuthMiddleware) GenerateToken(userID int, username string) (string, error) {
	args := m.Called(userID, username)
	return args.String(0), args.Error(1)
}

func (m *MockAuthMiddleware) ValidateToken(tokenString string) (*middleware.Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*middleware.Claims), args.Error(1)
}

func (m *MockAuthMiddleware) ProtectRoute(next http.Handler) http.Handler {
	args := m.Called(next)
	return args.Get(0).(http.Handler)
}

func (m *MockAuthMiddleware) GetUserIDFromContext(ctx context.Context) (int, bool) {
	args := m.Called(ctx)
	return args.Int(0), args.Bool(1)
}

func (m *MockAuthMiddleware) IsAuthenticated(r *http.Request) bool {
	args := m.Called(r)
	return args.Bool(0)
}
