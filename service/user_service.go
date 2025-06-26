package service

import (
	"context"
	"github.com/EfosaE/credora-backend/domain/logger"
	// "github.com/EfosaE/credora-backend/domain/monnify"
	"github.com/EfosaE/credora-backend/domain/user"
	// "github.com/EfosaE/credora-backend/internal/config"
	"github.com/EfosaE/credora-backend/internal/utils"
)

type UserService struct {
	userRepo   user.UserRepository
	logger     *logger.Logger
	monnifySvc *MonnifyService
	emailSvc   EmailService
}

func NewUserService(userRepo user.UserRepository, logger *logger.Logger, monnifySvc *MonnifyService, emailSvc EmailService) *UserService {
	return &UserService{
		userRepo:   userRepo,
		logger:     logger,
		monnifySvc: monnifySvc,
		emailSvc:   emailSvc,
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

	// send a welcome email in a go routine to the user
	go func(user user.User) {
		eCtx := context.Background()
		err := s.emailSvc.SendWelcomeEmail(eCtx, user)
		if err != nil {
			s.logger.Error("failed to send welcome email", map[string]any{
				"error": err.Error(),
				"to":    user.Email,
			})
			return
		}
		s.logger.Info("Welcome email sent", map[string]any{
			"to": user.Email,
		})
	}(*result)

	// // create a bank account for the user
	// monnifyCustResp, err := s.monnifySvc.CreateCustomer(&monnify.CreateCRAParams{
	// 	AccountReference:     result.ID.String(),
	// 	AccountName:          req.Name,
	// 	CurrencyCode:         "NGN",
	// 	ContractCode:         config.App.MonnifyContractCode,
	// 	CustomerEmail:        req.Email,
	// 	Nin:                  req.Nin,
	// 	CustomerName:         req.Name,
	// 	GetAllAvailableBanks: true,
	// })

	// if err != nil {
	// 	s.logger.Error("failed to create monnify customer", map[string]any{"error": err.Error()})
	// 	return nil, err
	// }

	// // send an email in a go routine to the user
	// go func(email, bankName, acctNumber string) {
	// 	eCtx := context.Background()
	// 	sent, err := s.emailSvc.SendAccountNumberEmail(eCtx, email, bankName, acctNumber)
	// 	if err != nil {
	// 		s.logger.Error("failed to send welcome email", map[string]any{
	// 			"error": err.Error(),
	// 			"to":    email,
	// 		})
	// 		return
	// 	}
	// 	s.logger.Info("Welcome email sent", map[string]any{
	// 		"id": sent.Id,
	// 	})
	// }(req.Email, monnifyCustResp.ResponseBody.Accounts[0].BankName, monnifyCustResp.ResponseBody.Accounts[0].AccountNumber)

	// s.logger.Info("User successfully created", map[string]any{"userID": result.ID, "user_account_ref": monnifyCustResp.ResponseBody.AccountReference})

	userResp := &user.CreateUserResponse{
		ID:    result.ID,
		Name:  req.Name,
		Email: req.Email,
		// AccountReference:     monnifyCustResp.ResponseBody.AccountReference,
		// AccountName:          monnifyCustResp.ResponseBody.AccountName,
		// Accounts:             monnifyCustResp.ResponseBody.Accounts,
		// ReservationReference: monnifyCustResp.ResponseBody.ReservationReference,
		// BankName:             monnifyCustResp.ResponseBody.Accounts[0].BankName,
		// AccountNumber:        monnifyCustResp.ResponseBody.Accounts[0].AccountNumber,
		// Status:               monnifyCustResp.ResponseBody.Status,
		CreatedAt: result.CreatedAt,
	}
	return userResp, nil

}
