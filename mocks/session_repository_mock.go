package mocks

import (
	"github.com/stretchr/testify/mock"
	"forum/models"
	"forum/repository"
)

type MockSessionRepository struct {
	mock.Mock
}

func NewMockSessionRepository() *MockSessionRepository {
	return &MockSessionRepository{}
}

func (m *MockSessionRepository) CreateSession(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockSessionRepository) ValidateSession(sessionToken string) (*models.User, error) {
	args := m.Called(sessionToken)
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

var _ repository.SessionRepositoryInterface = &MockSessionRepository{}
