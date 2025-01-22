package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"forum/models"
)

// MockCategoryRepository simulates category repository for testing
type MockCategoryRepository struct {
	mock.Mock
}

func (m *MockCategoryRepository) GetAll() ([]models.Category, error) {
	args := m.Called()
	return args.Get(0).([]models.Category), args.Error(1)
}

func (m *MockCategoryRepository) Create(category *models.Category) error {
	args := m.Called(category)
	return args.Error(0)
}

func (m *MockCategoryRepository) GetByID(categoryID int64) (*models.Category, error) {
	args := m.Called(categoryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Category), args.Error(1)
}

func (m *MockCategoryRepository) ListCategories() ([]models.Category, error) {
	args := m.Called()
	return args.Get(0).([]models.Category), args.Error(1)
}

func (m *MockCategoryRepository) Update(category *models.Category) error {
	args := m.Called(category)
	return args.Error(0)
}

func (m *MockCategoryRepository) Delete(categoryID int64) error {
	args := m.Called(categoryID)
	return args.Error(0)
}

func TestListCategories(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	handler := NewCategoryHandler(mockRepo)

	testCases := []struct {
		name           string
		mockBehavior   func()
		expectedStatus int
		expectedCount  int
	}{
		{
			name: "Successful Categories Retrieval",
			mockBehavior: func() {
				mockRepo.On("GetAll").Return([]models.Category{
					{ID: 1, Name: "Technology"},
					{ID: 2, Name: "Sports"},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name: "Database Error",
			mockBehavior: func() {
				mockRepo.On("GetAll").Return([]models.Category{}, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedCount:  0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior()

			req := httptest.NewRequest(http.MethodGet, "/categories", nil)
			w := httptest.NewRecorder()

			handler.ListCategories(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			if tc.expectedStatus == http.StatusOK {
				var categories []models.Category
				err := json.NewDecoder(w.Body).Decode(&categories)
				assert.NoError(t, err)
				assert.Len(t, categories, tc.expectedCount)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
