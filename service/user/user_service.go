package usersvc

import (
	"context"
	"encoding/json"

	"github.com/EfosaE/credora-backend/domain/event"
	"github.com/EfosaE/credora-backend/domain/logger"
	authsvc "github.com/EfosaE/credora-backend/service/auth"

	// accountsvc "github.com/EfosaE/credora-backend/service/account"

	"github.com/EfosaE/credora-backend/domain/user"

	"github.com/EfosaE/credora-backend/internal/eventbus"
	"github.com/EfosaE/credora-backend/internal/queues"
	"github.com/EfosaE/credora-backend/internal/utils"
	"github.com/EfosaE/credora-backend/service"
)

type UserService struct {
	userRepo   user.UserRepository
	logger     *logger.Logger
	eventBus   eventbus.EventBus
	monnifySvc *service.MonnifyService
	queue      queues.Queue // It should depend on the Queue interface for unit testing
}

func NewUserService(
	userRepo user.UserRepository,
	logger *logger.Logger,
	eventBus eventbus.EventBus,
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
	s.logger.Info("User creation initiated", map[string]any{"userName": req.Name, "email": req.Email})

	utils.PrintJSON(req) // Print the user request for debugging
	hashedPassword, _ := authsvc.HashPassword(req.Password)

	req.Password = hashedPassword

	result, err := s.userRepo.Create(ctx, req)
	if err != nil {
		s.logger.Error("failed to create user", map[string]any{"error": err.Error()})
		return nil, err
	}

	// create a bank account for the user
	monnifyCustResp, err := s.CreateVirtualAccount(ctx, req, result.ID.String())

	if err != nil {
		s.logger.Error("failed to create monnify customer", map[string]any{"error": err.Error()})
		return nil, err
	}

	// Enqueue account number email
	if err := s.SendAccountNumberEmailAsync(result.Email, monnifyCustResp.ResponseBody.Accounts[0].BankName, monnifyCustResp.ResponseBody.Accounts[0].AccountNumber); err != nil {
		s.logger.Error("Failed to enqueue account number email:", map[string]any{"error": err.Error()})
	}
	event := event.UserCreatedEvent{
		UserID:        result.ID,
		AccountNumber: monnifyCustResp.ResponseBody.Accounts[0].AccountNumber,
		Name:          result.Name,
		BankName:      monnifyCustResp.ResponseBody.Accounts[0].BankName,
		Email:         result.Email,
	}

	data, _ := json.Marshal(event)
	s.eventBus.Publish(ctx, "user.created", map[string]any{
		"data": string(data),
	})

	s.logger.Info("User successfully created", map[string]any{"userID": result.ID, "user_account_ref": monnifyCustResp.ResponseBody.AccountReference})

	userResp := &user.CreateUserResponse{

		ID:                   result.ID,
		Name:                 req.Name,
		Email:                req.Email,
		AccountReference:     monnifyCustResp.ResponseBody.AccountReference,
		AccountName:          monnifyCustResp.ResponseBody.AccountName,
		Accounts:             monnifyCustResp.ResponseBody.Accounts,
		ReservationReference: monnifyCustResp.ResponseBody.ReservationReference,
		BankName:             monnifyCustResp.ResponseBody.Accounts[0].BankName,
		AccountNumber:        monnifyCustResp.ResponseBody.Accounts[0].AccountNumber,
		Status:               monnifyCustResp.ResponseBody.Status,
		CreatedAt:            result.CreatedAt,
	}
	return userResp, nil
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*user.User, error) {
	s.logger.Info("Fetching user by email", map[string]any{"email": email})
	usr, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		s.logger.Error("failed to fetch user by email", map[string]any{"error": err.Error()})
		return nil, err
	}
	return usr, nil
}

// func (s *UserService) GetUserByID(ctx context.Context, id string) (*user.User, error) {
// 	s.logger.Info("Fetching user by ID", map[string]any{"id": id})
// 	usr, err := s.userRepo.GetByID(ctx, id)
// 	if err != nil {
// 		s.logger.Error("failed to fetch user by ID", map[string]any{"error": err.Error()})
// 		return nil, err
// 	}
// 	return usr, nil
// }
