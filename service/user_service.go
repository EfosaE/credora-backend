package service

import (
	"context"

	"github.com/EfosaE/credora-backend/domain/logger"
	"github.com/EfosaE/credora-backend/domain/user"
	"github.com/EfosaE/credora-backend/internal/utils"
)

type UserService struct {
	userRepo   user.UserRepository
	logger     *logger.Logger
	monnifySvc *MonnifyService
}

func NewUserService(userRepo user.UserRepository, logger *logger.Logger, monnifySvc *MonnifyService) *UserService {
	return &UserService{
		userRepo: userRepo,
		logger:   logger,
		monnifySvc: monnifySvc,
	}
}

func (s *UserService) CreateUser(ctx context.Context, user *user.CreateUserRequest) (*user.User, error) {
	utils.PrintJSON(user) // Print the user request for debugging
	hashedPassword, _ := HashPassword(user.Password)

	user.Password = hashedPassword

	result, err := s.userRepo.Create(ctx, user)
	if err != nil {
		s.logger.Error("failed to create user", map[string]any{"error": err.Error()})
		return nil, err
	}
	s.logger.Info("User successfully created", map[string]any{"userID": result.ID})
	return result, nil

}
