package mocks

import (
	"context"

	"github.com/EfosaE/credora-backend/domain/user"
	"github.com/google/uuid"
)

// MockUserRepository is a mock implementation of the UserRepository interface for testing purposes.
// It allows you to define custom behavior for the Create method, which can be useful in unit tests.
type MockUserRepo struct {
	CreateFunc func(ctx context.Context, req *user.CreateUserRequest) (*user.User, error)
	Users      map[int]*user.User
}

func (m *MockUserRepo) Create(ctx context.Context, req *user.CreateUserRequest) (*user.User, error) {
	return m.CreateFunc(ctx, req)
}

func (m *MockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	// Implement this method if needed for your tests
	return nil, nil
}