package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"forum/middleware"
	"forum/mocks"
	"forum/models"
)

// PostRepositoryInterface defines the interface for post repository
type PostRepositoryInterface interface {
	Create(post *models.Post, categoryIDs []int64) error
	GetByID(postID string) (*models.Post, error)
	ListPosts(page, limit int) ([]models.Post, int, error)
	GetPostsByCategory(categoryID int64) ([]models.Post, error)
	GetUserPosts(userID int64) ([]models.Post, error)
	GetLikedPosts(userID int64) ([]models.Post, error)
	UpdatePost(post *models.Post, categoryIDs []int64, userID int64) error
	DeletePost(postID string, userID int64) error
	FilterPosts(filters struct {
		CategoryID  *int64
		UserID      *int64
		LikedByUser *int64
	}) ([]models.Post, error)
}

// MockPostRepository simulates post repository for testing
type MockPostRepository struct {
	mock.Mock
}

func (m *MockPostRepository) Create(post *models.Post, categoryIDs []int64) error {
	args := m.Called(post, categoryIDs)
	return args.Error(0)
}

func (m *MockPostRepository) GetByID(postID string) (*models.Post, error) {
	args := m.Called(postID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Post), args.Error(1)
}

func (m *MockPostRepository) ListPosts(page, limit int) ([]models.Post, int, error) {
	args := m.Called(page, limit)
	return args.Get(0).([]models.Post), args.Int(1), args.Error(2)
}

func (m *MockPostRepository) GetPostsByCategory(categoryID int64) ([]models.Post, error) {
	args := m.Called(categoryID)
	return args.Get(0).([]models.Post), args.Error(1)
}

func (m *MockPostRepository) GetUserPosts(userID int64) ([]models.Post, error) {
	args := m.Called(userID)
	return args.Get(0).([]models.Post), args.Error(1)
}

func (m *MockPostRepository) GetLikedPosts(userID int64) ([]models.Post, error) {
	args := m.Called(userID)
	return args.Get(0).([]models.Post), args.Error(1)
}

func (m *MockPostRepository) UpdatePost(post *models.Post, categoryIDs []int64, userID int64) error {
	args := m.Called(post, categoryIDs, userID)
	return args.Error(0)
}

func (m *MockPostRepository) DeletePost(postID string, userID int64) error {
	args := m.Called(postID, userID)
	return args.Error(0)
}

func (m *MockPostRepository) FilterPosts(filters struct {
	CategoryID  *int64
	UserID      *int64
	LikedByUser *int64
},
) ([]models.Post, error) {
	args := m.Called(filters)
	return args.Get(0).([]models.Post), args.Error(1)
}

type PostPayload map[PostPayloadKey]interface{}

type PostPayloadKey string

const (
	TitleKey      PostPayloadKey = "title"
	ContentKey    PostPayloadKey = "content"
	CategoriesKey PostPayloadKey = "categories"
)

// Define a custom type for context keys to avoid string key collisions
type contextKey string

// Define a specific context key for user claims
var UserClaimsKey contextKey = "user_claims"

func TestCreatePost(t *testing.T) {
	mockPostRepo := new(MockPostRepository)
	mockAuthMiddleware := mocks.NewMockAuthMiddleware(t)
	mockTemplateRenderer := NewTemplateRenderer(mockAuthMiddleware)
	handler := NewPostHandler(mockPostRepo, mockAuthMiddleware, mockTemplateRenderer)

	testCases := []struct {
		name           string
		payload        PostPayload
		mockBehavior   func()
		expectedStatus int
	}{
		{
			name: "Successful Post Creation",
			payload: PostPayload{
				TitleKey:      "Test Post",
				ContentKey:    "This is a test post content",
				CategoriesKey: []int64{1, 2},
			},
			mockBehavior: func() {
				mockPostRepo.On("Create", mock.AnythingOfType("*models.Post"), mock.AnythingOfType("[]int64")).Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Empty Title",
			payload: PostPayload{
				TitleKey:      "",
				ContentKey:    "This is a test post content",
				CategoriesKey: []int64{1, 2},
			},
			mockBehavior:   func() {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior()

			payload, _ := json.Marshal(tc.payload)
			req := httptest.NewRequest(http.MethodPost, "/posts", bytes.NewBuffer(payload))
			req.Header.Set("Content-Type", "application/json")

			// Add user context for authentication
			ctx := context.WithValue(req.Context(), UserClaimsKey, &middleware.Claims{
				UserID: 1,
			})
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()

			handler.CreatePost(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			mockPostRepo.AssertExpectations(t)
		})
	}
}

func TestGetPosts(t *testing.T) {
	mockPostRepo := new(MockPostRepository)
	mockAuthMiddleware := mocks.NewMockAuthMiddleware(t)
	mockTemplateRenderer := NewTemplateRenderer(mockAuthMiddleware)
	handler := NewPostHandler(mockPostRepo, mockAuthMiddleware, mockTemplateRenderer)

	testCases := []struct {
		name           string
		page           int
		limit          int
		mockBehavior   func()
		expectedStatus int
	}{
		{
			name:  "Get All Posts",
			page:  1,
			limit: 10,
			mockBehavior: func() {
				mockPostRepo.On("ListPosts", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return([]models.Post{
					{ID: 1, Title: "Post 1"},
					{ID: 2, Title: "Post 2"},
				}, 2, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "Get Posts by Category",
			page:  1,
			limit: 10,
			mockBehavior: func() {
				mockPostRepo.On("GetPostsByCategory", int64(1)).Return([]models.Post{
					{ID: 1, Title: "Category Post 1"},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid Category ID",
			page:           1,
			limit:          10,
			mockBehavior:   func() {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior()

			req := httptest.NewRequest(http.MethodGet, "/posts", nil)
			q := req.URL.Query()
			q.Add("page", fmt.Sprintf("%d", tc.page))
			q.Add("limit", fmt.Sprintf("%d", tc.limit))
			req.URL.RawQuery = q.Encode()

			// Add user context for authentication
			ctx := context.WithValue(req.Context(), UserClaimsKey, &middleware.Claims{
				UserID: 1,
			})
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()

			handler.ListPosts(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			if tc.expectedStatus == http.StatusOK {
				var posts []models.Post
				err := json.NewDecoder(w.Body).Decode(&posts)
				assert.NoError(t, err)
			}

			mockPostRepo.AssertExpectations(t)
		})
	}
}

func TestGetPost(t *testing.T) {
	mockPostRepo := new(MockPostRepository)
	mockAuthMiddleware := mocks.NewMockAuthMiddleware(t)
	mockTemplateRenderer := NewTemplateRenderer(mockAuthMiddleware)
	handler := NewPostHandler(mockPostRepo, mockAuthMiddleware, mockTemplateRenderer)

	testCases := []struct {
		name           string
		postID         string
		mockBehavior   func()
		expectedStatus int
	}{
		{
			name:   "Get Post by ID",
			postID: "1",
			mockBehavior: func() {
				mockPostRepo.On("GetByID", "1").Return(&models.Post{
					ID:      1,
					Title:   "Test Post",
					Content: "Test Content",
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior()

			// Create a response recorder and request
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/posts/1", nil)

			// Add chi context for URL parameters
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.postID)
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			// Add user context for authentication
			ctx := context.WithValue(r.Context(), UserClaimsKey, &middleware.Claims{
				UserID: 1,
			})
			r = r.WithContext(ctx)

			handler.GetPost(w, r)

			// Check response
			resp := w.Result()
			body, _ := ioutil.ReadAll(resp.Body)

			assert.Equal(t, tc.expectedStatus, resp.StatusCode)

			var post models.Post
			err := json.Unmarshal(body, &post)
			assert.NoError(t, err)
			assert.Equal(t, "Test Post", post.Title)

			mockPostRepo.AssertExpectations(t)
		})
	}
}
