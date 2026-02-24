package authsvc

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/EfosaE/credora-backend/domain/account"
	"github.com/EfosaE/credora-backend/domain/auth"
	domainerr "github.com/EfosaE/credora-backend/domain/domianerrors"
	"github.com/EfosaE/credora-backend/domain/user"
	"github.com/EfosaE/credora-backend/infrastructure"
	"github.com/EfosaE/credora-backend/internal/config"
	"github.com/EfosaE/credora-backend/internal/utils"
	"github.com/EfosaE/credora-backend/service"
)

type AuthService struct {
	txManager         infrastructure.TxManager
	userRepo          user.UserRepository
	passwordResetRepo auth.PasswordResetRepository
	acctRepo          account.AccountRepository
	tokenService      auth.TokenService
	mailer            service.EmailService
}

func NewAuthService(txManager infrastructure.TxManager, userRepo user.UserRepository, passwordResetRepo auth.PasswordResetRepository, tokenService auth.TokenService, acctRepo account.AccountRepository, mailer service.EmailService) *AuthService {
	return &AuthService{
		txManager:         txManager,
		userRepo:          userRepo,
		passwordResetRepo: passwordResetRepo,
		tokenService:      tokenService,
		acctRepo:          acctRepo,
		mailer:            mailer,
	}
}

func (s *AuthService) Login(ctx context.Context, accountNumber, password string) (*account.GetUserDetailsWithAccountRow, string, error) {
	u, err := s.acctRepo.GetUserByAccountNumber(ctx, accountNumber)
	if err != nil {
		return nil, "", fmt.Errorf("login: %w", domainerr.ErrUserNotFound)
	}

	if !CheckPasswordHash(password, u.Password) {
		return nil, "", domainerr.ErrInvalidCredentials
	}

	token, err := s.tokenService.GenerateToken(ctx, auth.TokenPayload{
		UserID:        u.UserId,
		AccountNumber: u.AccountNumber,
		Name:          u.FullName,
	})
	if err != nil {
		return nil, "", fmt.Errorf("login: %w", err)
	}

	return u, token, nil
}

func (s *AuthService) RequestPasswordReset(ctx context.Context, email string) error {
	u, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		fmt.Printf("Failed to get user by email: %v\n", err)
		return nil // swallow
	}

	utils.PrintJSON(u) // Print the user for debugging
	raw, hash, expiresAt, err := GenerateResetToken()
	if err != nil {
		fmt.Printf("Failed to generate reset token: %v\n", err)
		return nil
	}

	fmt.Printf("Generated reset token: raw=%s, hash=%s, expiresAt=%v\n", raw, hash, expiresAt) // Print the generated token for debugging
	if _, err := s.passwordResetRepo.Create(ctx, u.ID, hash, expiresAt); err != nil {
		fmt.Printf("Failed to create password reset record: %v\n", err)
		return nil
	}

	resetLink := fmt.Sprintf("%s/reset-password?token=%s&email=%s", config.App.FrontendURL, raw, u.Email)

	fmt.Printf("Generated reset link: %s\n", resetLink) // Print the reset link for debugging
	// This can be ENQUEUED - we don't want to block the response if email sending fails. We can log the error instead.
	err = s.mailer.SendPasswordResetEmail(ctx, u.Email, resetLink)
	if err != nil {
		fmt.Printf("Failed to send password reset email: %v\n", err)
		return err
	}
	return nil
}

func (s *AuthService) ValidatePasswordResetRequest(
	ctx context.Context,
	email string,
	token string,
	newPassword string,
) error {

	// 1. Fetch user (anti-enumeration)
	u, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		log.Printf("ValidatePasswordResetRequest: failed to fetch user. email=%s err=%v", email, err)
		return err
	}

	// 2. Fetch active reset token
	reset, err := s.passwordResetRepo.GetActiveToken(ctx, u.ID)
	if err != nil {
		log.Printf("ValidatePasswordResetRequest: failed to fetch reset token. user_id=%s err=%v", u.ID, err)
		return err
	}

	// 3. Validate token
	if !ValidateResetToken(token, reset.TokenHash) {
		log.Printf("ValidatePasswordResetRequest: invalid token attempt. user_id=%s", u.ID)
		return fmt.Errorf("invalid or expired reset token")
	}

	// 4. Check expiration
	if time.Now().After(reset.ExpiresAt) {
		log.Printf("ValidatePasswordResetRequest: expired token used. user_id=%s expires_at=%v", u.ID, reset.ExpiresAt)
		return fmt.Errorf("invalid or expired reset token")
	}

	// 5. Hash new password
	hashedPassword, err := HashPassword(newPassword)
	if err != nil {
		log.Printf("ValidatePasswordResetRequest: failed to hash password. user_id=%s err=%v", u.ID, err)
		return err
	}

	// 6. Update password + consume token (transaction)
	err = s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.userRepo.UpdatePassword(txCtx, u.ID, hashedPassword); err != nil {
			log.Printf("ValidatePasswordResetRequest: failed to update password. user_id=%s err=%v", u.ID, err)
			return err
		}

		if err := s.passwordResetRepo.MarkUsed(txCtx, reset.ID, time.Now()); err != nil {
			log.Printf("ValidatePasswordResetRequest: failed to mark token used. user_id=%s reset_id=%d err=%v", u.ID, reset.ID, err)
			return err
		}

		return nil
	})

	if err != nil {
		log.Printf("ValidatePasswordResetRequest: transaction failed. user_id=%s err=%v", u.ID, err)
		return err
	}

	log.Printf("ValidatePasswordResetRequest: password reset successful. user_id=%s", u.ID)
	return nil
}
