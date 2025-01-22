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
	"forum/repository"
)

// MockCommentRepository simulates comment repository for testing
type MockCommentRepository struct {
	mock.Mock
}

func (m *MockCommentRepository) Create(comment *repository.Comment) error {
	args := m.Called(comment)
	return args.Error(0)
}

func (m *MockCommentRepository) GetByPostID(postID int64) ([]repository.Comment, error) {
	args := m.Called(postID)
	repoComments := args.Get(0).([]repository.Comment)

	return repoComments, args.Error(1)
}

// Ensure MockCommentRepository implements the interface
var _ repository.CommentRepositoryInterface = &MockCommentRepository{}

func TestCreateComment(t *testing.T) {
	mockRepo := new(MockCommentRepository)
	mockAuthMiddleware := &middleware.AuthMiddleware{}
	handler := NewCommentHandler(mockRepo, mockAuthMiddleware)

	testCases := []struct {
		name           string
		payload        map[string]interface{}
		mockBehavior   func()
		expectedStatus int
	}{
		{
			name: "Successful Comment Creation",
			payload: map[string]interface{}{
				"post_id": 1,
				"content": "Test comment",
			},
			mockBehavior: func() {
				mockRepo.On("Create", mock.AnythingOfType("*repository.Comment")).Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Empty Comment Content",
			payload: map[string]interface{}{
				"post_id": 1,
				"content": "",
			},
			mockBehavior:   func() {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior()

			payload, _ := json.Marshal(tc.payload)
			req := httptest.NewRequest(http.MethodPost, "/comments", bytes.NewBuffer(payload))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.CreateComment(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetComments(t *testing.T) {
	mockRepo := new(MockCommentRepository)
	mockAuthMiddleware := &middleware.AuthMiddleware{}
	handler := NewCommentHandler(mockRepo, mockAuthMiddleware)

	testCases := []struct {
		name           string
		postID         string
		mockBehavior   func()
		expectedStatus int
		expectedCount  int
	}{
		{
			name:   "Successful Comments Retrieval",
			postID: "1",
			mockBehavior: func() {
				mockRepo.On("GetByPostID", int64(1)).Return([]repository.Comment{
					{ID: 1, PostID: 1, Content: "Comment 1"},
					{ID: 2, PostID: 1, Content: "Comment 2"},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:   "Invalid Post ID",
			postID: "invalid",
			mockBehavior: func() {
				// No mock needed
			},
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior()

			req := httptest.NewRequest(http.MethodGet, "/comments", nil)
			q := req.URL.Query()
			q.Add("post_id", tc.postID)
			req.URL.RawQuery = q.Encode()

			w := httptest.NewRecorder()

			handler.GetComments(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			if tc.expectedStatus == http.StatusOK {
				var comments []repository.Comment
				err := json.NewDecoder(w.Body).Decode(&comments)
				assert.NoError(t, err)
				assert.Len(t, comments, tc.expectedCount)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
