package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"forum/middleware"
	"forum/models"
	"forum/repository"
)

// MockInteractionRepository simulates interaction repository for testing
type MockInteractionRepository struct {
	mock.Mock
}

func (m *MockInteractionRepository) AddInteraction(userID, entityID int64, entityType string, interactionType models.InteractionType) error {
	args := m.Called(userID, entityID, entityType, interactionType)
	return args.Error(0)
}

func (m *MockInteractionRepository) GetInteractionCounts(entityID int64, entityType string) (likes, dislikes int, err error) {
	args := m.Called(entityID, entityType)
	return args.Int(0), args.Int(1), args.Error(2)
}

func (m *MockInteractionRepository) RemoveInteraction(userID, entityID int64, entityType string) error {
	args := m.Called(userID, entityID, entityType)
	return args.Error(0)
}

// Ensure MockInteractionRepository implements the interface
var _ repository.InteractionRepositoryInterface = (*MockInteractionRepository)(nil)

func TestLikeEntity(t *testing.T) {
	mockRepo := new(MockInteractionRepository)
	mockAuthMiddleware := &middleware.AuthMiddleware{}
	handler := NewInteractionHandler(mockRepo, mockAuthMiddleware)

	testCases := []struct {
		name           string
		payload        map[string]interface{}
		mockBehavior   func()
		expectedStatus int
	}{
		{
			name: "Successful Like",
			payload: map[string]interface{}{
				"entity_id":   1,
				"entity_type": "post",
			},
			mockBehavior: func() {
				mockRepo.On("AddInteraction", int64(1), int64(1), "post", models.InteractionType(models.Like)).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid Entity Type",
			payload: map[string]interface{}{
				"entity_id":   1,
				"entity_type": "",
			},
			mockBehavior:   func() {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior()

			payload, _ := json.Marshal(tc.payload)
			req := httptest.NewRequest(http.MethodPost, "/like", bytes.NewBuffer(payload))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.LikeEntity(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetInteractionCounts(t *testing.T) {
	mockRepo := new(MockInteractionRepository)
	mockAuthMiddleware := &middleware.AuthMiddleware{}
	handler := NewInteractionHandler(mockRepo, mockAuthMiddleware)

	testCases := []struct {
		name           string
		entityID       int64
		entityType     string
		mockBehavior   func()
		expectedStatus int
	}{
		{
			name:       "Successful Interaction Counts",
			entityID:   1,
			entityType: "post",
			mockBehavior: func() {
				mockRepo.On("GetInteractionCounts", int64(1), "post").Return(5, 2, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid Entity Type",
			entityID:       1,
			entityType:     "",
			mockBehavior:   func() {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior()

			req := httptest.NewRequest(http.MethodGet, "/interactions", nil)
			q := req.URL.Query()
			q.Add("entity_id", "1")
			q.Add("entity_type", tc.entityType)
			req.URL.RawQuery = q.Encode()

			w := httptest.NewRecorder()

			handler.GetInteractionCounts(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}
