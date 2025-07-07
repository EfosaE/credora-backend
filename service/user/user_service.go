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
	"github.com/EfosaE/credora-backend/internal/utils"
	"github.com/EfosaE/credora-backend/service"
)

type UserService struct {
	userRepo   user.UserRepository
	logger     *logger.Logger
	eventBus   eventbus.EventBus
	monnifySvc *service.MonnifyService
	// emailSvc   service.EmailService
	// acctSvc    *accountsvc.AccountService
}

func NewUserService(
	userRepo user.UserRepository,
	logger *logger.Logger,
	eventBus eventbus.EventBus,
	monnifySvc *service.MonnifyService,
	// emailSvc service.EmailService,
	// acctSvc *accountsvc.AccountService,
) *UserService {
	return &UserService{
		userRepo:   userRepo,
		logger:     logger,
		eventBus:   eventBus,
		monnifySvc: monnifySvc,
		// emailSvc:   emailSvc,
		// acctSvc:    acctSvc,
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
