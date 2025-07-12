package authsvc

import (
	"context"

	"github.com/EfosaE/credora-backend/domain/account"
	"github.com/EfosaE/credora-backend/domain/auth"
	domainerr "github.com/EfosaE/credora-backend/domain/domianerrors"
	// "github.com/EfosaE/credora-backend/domain/user"
)

type AuthService struct {
	// userRepo     user.UserRepository
	acctRepo     account.AccountRepository
	tokenService auth.TokenService
}

func NewAuthService(tokenService auth.TokenService, acctRepo account.AccountRepository) *AuthService {
	return &AuthService{
		// userRepo:     userRepo,
		tokenService: tokenService,
		acctRepo:     acctRepo,
	}
}

func (s *AuthService) Login(ctx context.Context, accountNumber, password string) (string, error) {
	user, err := s.acctRepo.GetUserByAccountNumber(ctx, accountNumber)
	// fmt.Println("Error fetching  by account:", err)
	if err != nil {
		return "", domainerr.ErrUserNotFound
	}

	// 2. Verify password
	if !CheckPasswordHash(password, user.Password) {
		return "", domainerr.ErrInvalidCredentials
	}

	// 3. Generate tokens
	return s.tokenService.GenerateToken(ctx, auth.TokenPayload{
		UserID:        user.ID,
		AccountNumber: user.AccountNumber,
		Name:          user.FullName,
	})
}
