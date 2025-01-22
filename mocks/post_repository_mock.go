package mocks

import (
	"forum/models"
	"github.com/stretchr/testify/mock"
)

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
