package service

import (
	"context"

	"github.com/EfosaE/credora-backend/domain/logger"
	"github.com/EfosaE/credora-backend/domain/monnify"
	"github.com/EfosaE/credora-backend/domain/user"
	"github.com/EfosaE/credora-backend/internal/config"
	"github.com/EfosaE/credora-backend/internal/utils"
)

type UserService struct {
	userRepo   user.UserRepository
	logger     *logger.Logger
	monnifySvc *MonnifyService
}

func NewUserService(userRepo user.UserRepository, logger *logger.Logger, monnifySvc *MonnifyService) *UserService {
	return &UserService{
		userRepo:   userRepo,
		logger:     logger,
		monnifySvc: monnifySvc,
	}
}

func (s *UserService) CreateUser(ctx context.Context, req *user.CreateUserRequest) (*user.CreateUserResponse, error) {
	s.logger.Info("User creation initiated", map[string]any{"userName": req.Name, "email": req.Email})

	utils.PrintJSON(req) // Print the user request for debugging
	hashedPassword, _ := HashPassword(req.Password)

	req.Password = hashedPassword

	result, err := s.userRepo.Create(ctx, req)
	if err != nil {
		s.logger.Error("failed to create user", map[string]any{"error": err.Error()})
		return nil, err
	}

	// create a bank account for the user
	monnifyCustResp, err := s.monnifySvc.CreateCustomer(&monnify.CreateCRAParams{
		AccountReference:     result.ID.String(),
		AccountName:          req.Name,
		CurrencyCode:         "NGN",
		ContractCode:         config.App.MonnifyContractCode,
		CustomerEmail:        req.Email,
		Nin:                  req.Nin,
		CustomerName:         req.Name,
		GetAllAvailableBanks: true,
	})

	if err != nil {
		s.logger.Error("failed to create monnify customer", map[string]any{"error": err.Error()})
		return nil, err
	}

	// send an email in a go routine to the user

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
