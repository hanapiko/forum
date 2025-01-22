package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"forum/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSessionRepository simulates session repository for testing
type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) CreateSession(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockSessionRepository) ValidateSession(token string) (*models.User, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockSessionRepository) InvalidateSession(userID int) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockSessionRepository) Validate(token string) (int64, error) {
	args := m.Called(token)
	return args.Get(0).(int64), args.Error(1)
}

func TestExtractToken(t *testing.T) {
	mockSessionRepo := new(MockSessionRepository)
	authMiddleware := NewAuthMiddleware(mockSessionRepo, "test-secret-key", 24)

	testCases := []struct {
		name          string
		headerValue   string
		expectedToken string
		expectedError bool
	}{
		{
			name:          "Valid Bearer Token",
			headerValue:   "Bearer test-token-123",
			expectedToken: "test-token-123",
			expectedError: false,
		},
		{
			name:          "Invalid Token Format",
			headerValue:   "Invalid test-token-123",
			expectedToken: "",
			expectedError: true,
		},
		{
			name:          "Empty Authorization Header",
			headerValue:   "",
			expectedToken: "",
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Authorization", tc.headerValue)

			token, err := authMiddleware.extractToken(req)

			if err != nil {
				if !tc.expectedError {
					t.Errorf("extractToken() unexpected error = %v", err)
				}
			} else {
				if tc.expectedError {
					t.Errorf("extractToken() expected error, got none")
				}
				if token != tc.expectedToken {
					t.Errorf("extractToken() got = %v, want %v", token, tc.expectedToken)
				}
			}
		})
	}
}

func TestProtectRoute(t *testing.T) {
	mockSessionRepo := new(MockSessionRepository)
	middleware := NewAuthMiddleware(mockSessionRepo, "test-secret-key", 24)

	// Test handler to wrap
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	testCases := []struct {
		name           string
		token          string
		mockBehavior   func()
		expectedStatus int
	}{
		{
			name:  "Valid Authentication",
			token: "valid-token",
			mockBehavior: func() {
				mockSessionRepo.On("Validate", "valid-token").Return(int64(1), nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "Invalid Token",
			token: "invalid-token",
			mockBehavior: func() {
				mockSessionRepo.On("Validate", "invalid-token").Return(int64(0), assert.AnError)
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior()

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Authorization", "Bearer "+tc.token)
			w := httptest.NewRecorder()

			protectedHandler := middleware.ProtectRoute(testHandler)
			protectedHandler.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			mockSessionRepo.AssertExpectations(t)
		})
	}
}
