package service

import (
	"context"

	"github.com/EfosaE/credora-backend/domain/user"
)

type UserService struct {
	repo user.UserRepository
}

func NewUserService(repo user.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (s *UserService) CreateUser(ctx context.Context, user *user.CreateUserRequest) (*user.User, error) {
	return s.repo.Create(ctx, user)
}