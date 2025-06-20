package mocks

import (
	"context"

	"github.com/EfosaE/credora-backend/domain/user"
)

type MockUserService struct {
	CreateUserFunc func(ctx context.Context, req *user.CreateUserRequest) (*user.User, error)
}

func (m *MockUserService) CreateUser(ctx context.Context, req *user.CreateUserRequest) (*user.User, error) {
	return m.CreateUserFunc(ctx, req)
}
