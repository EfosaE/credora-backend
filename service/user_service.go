package service

import (
	"context"

	"github.com/EfosaE/credora-backend/domain/logger"
	"github.com/EfosaE/credora-backend/domain/user"
)

type UserService struct {
	repo   user.UserRepository
	logger *logger.Logger
}

func NewUserService(repo user.UserRepository, logger *logger.Logger) *UserService {
	return &UserService{
		repo:   repo,
		logger: logger,
	}
}

func (s *UserService) CreateUser(ctx context.Context, user *user.CreateUserRequest) (*user.User, error) {
	result, err := s.repo.Create(ctx, user)
	if err != nil {
		s.logger.Error("repo.Create failed", map[string]any{"error": err.Error()})
		return nil, err
	}
	s.logger.Info("User successfully created", map[string]any{"userID": result.ID})
	return result, nil

}
