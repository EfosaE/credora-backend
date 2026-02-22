package usersvc

import (
	"context"
	"fmt"

	"github.com/EfosaE/credora-backend/domain/event"
	authsvc "github.com/EfosaE/credora-backend/service/auth"
	"github.com/rs/zerolog"

	// accountsvc "github.com/EfosaE/credora-backend/service/account"

	"github.com/EfosaE/credora-backend/domain/user"
	"github.com/EfosaE/credora-backend/internal/queues"
	"github.com/EfosaE/credora-backend/internal/utils"
	"github.com/EfosaE/credora-backend/service"
)

type UserService struct {
	userRepo   user.UserRepository
	logger     zerolog.Logger
	eventBus   event.EventBus
	monnifySvc *service.MonnifyService
	queue      queues.Queue // It should depend on the Queue interface for unit testing
}

func NewUserService(
	userRepo user.UserRepository,
	logger zerolog.Logger,
	eventBus event.EventBus,
	monnifySvc *service.MonnifyService,
	queue queues.Queue,
) *UserService {
	return &UserService{
		userRepo:   userRepo,
		logger:     logger,
		eventBus:   eventBus,
		monnifySvc: monnifySvc,
		queue:      queue,
	}
}

func (s *UserService) CreateUser(ctx context.Context, req *user.CreateUserRequest) (*user.CreateUserResponse, error) {
	s.logger.Info().
		Str("userName", req.Name).
		Str("email", req.Email).
		Msg("User creation initiated")

	// utils.PrintJSON(req) // Print the user request for debugging
	hashedPassword, _ := authsvc.HashPassword(req.Password)

	req.Password = hashedPassword

	result, err := s.userRepo.Create(ctx, req)
	if err != nil {
		s.logger.Error().
			Str("error", err.Error()).
			Msg("failed to create user")
		return nil, err
	}

	// create a bank account for the user
	monnifyCustResp, err := s.CreateVirtualAccount(ctx, req, result.ID.String())

	if err != nil {
		s.logger.Error().
			Str("error", err.Error()).
			Msg("failed to create monnify customer")
		return nil, err
	}

	// Enqueue account number email
	if err := s.SendAccountNumberEmailAsync(result.Email, monnifyCustResp.ResponseBody.Accounts[0].BankName, monnifyCustResp.ResponseBody.Accounts[0].AccountNumber); err != nil {
		s.logger.Error().
			Str("error", err.Error()).
			Msg("Failed to enqueue account number email")
	}
	evt := event.UserCreatedEvent{
		UserID:        result.ID,
		AccountNumber: monnifyCustResp.ResponseBody.Accounts[0].AccountNumber,
		Name:          result.Name,
		BankName:      monnifyCustResp.ResponseBody.Accounts[0].BankName,
		Email:         result.Email,
	}

	payload, err := utils.StructToMap(evt)
	if err != nil {
		return nil, fmt.Errorf("failed to convert typed struct to map: %w", err)
	}
	s.eventBus.Publish(ctx, event.StreamUserEvents, event.EventUserCreated, payload)

	s.logger.Info().
		Str("userID", result.ID.String()).
		Str("userName", result.Name).
		Str("user_account_ref", monnifyCustResp.ResponseBody.AccountReference).
		Msg("User successfully created")

	userResp := &user.CreateUserResponse{
		ID:               result.ID,
		Name:             req.Name,
		Email:            req.Email,
		AccountReference: monnifyCustResp.ResponseBody.AccountReference,
		// AccountName:          monnifyCustResp.ResponseBody.AccountName,
		// Accounts:             monnifyCustResp.ResponseBody.Accounts,
		ReservationReference: monnifyCustResp.ResponseBody.ReservationReference,
		// BankName:             monnifyCustResp.ResponseBody.Accounts[0].BankName,
		// AccountNumber:        monnifyCustResp.ResponseBody.Accounts[0].AccountNumber,
		Status:    monnifyCustResp.ResponseBody.Status,
		CreatedAt: result.CreatedAt,
	}
	return userResp, nil
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*user.User, error) {
	s.logger.Info().
		Str("email", email).
		Msg("Fetching user by email")
	usr, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		s.logger.Error().
			Err(err).
			Msg("failed to fetch user by email")
		return nil, err
	}
	return usr, nil
}
