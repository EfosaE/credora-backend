package usersvc

import (
	"context"

	"github.com/EfosaE/credora-backend/domain/monnify"
	"github.com/EfosaE/credora-backend/domain/user"
	"github.com/EfosaE/credora-backend/internal/config"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (s *UserService) CreateVirtualAccount(ctx context.Context, req *user.CreateUserRequest, acctRef string) (*monnify.CreateCRAResponse, error) {
	monnifyCustResp, err := s.monnifySvc.CreateCustomer(&monnify.CreateCRAParams{
		AccountReference:     acctRef,
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
	s.logger.Info("Monnify virtual account created", map[string]any{"accountRef": acctRef})
	return monnifyCustResp, nil
}

func (s *UserService) SendPostSignupEmails(user user.User, acct *monnify.CreateCRAResponse) {
	go func() {
		ctx := context.Background()

		// if err := s.emailSvc.SendWelcomeEmail(ctx, user); err != nil {
		// 	s.logger.Error("failed to send welcome email", map[string]any{
		// 		"error": err.Error(),
		// 		"to":    user.Email,
		// 	})
		// } else {
		// 	s.logger.Info("Welcome email sent", map[string]any{"to": user.Email})
		// }

		firstAcct := acct.ResponseBody.Accounts[0]
		if err := s.emailSvc.SendAccountNumberEmail(ctx, user.Email, firstAcct.BankName, firstAcct.AccountNumber); err != nil {
			s.logger.Error("failed to send account number email", map[string]any{
				"error": err.Error(),
				"to":    user.Email,
			})
		} else {
			s.logger.Info("Account number email sent", map[string]any{
				"to": user.Email,
				"id": acct.ResponseBody.AccountReference,
			})
		}
	}()
}
