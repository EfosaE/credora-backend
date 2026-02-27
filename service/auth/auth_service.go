package authsvc

import (
	"context"
	"fmt"
	"time"

	"github.com/EfosaE/credora-backend/domain/account"
	"github.com/EfosaE/credora-backend/domain/auth"
	domainerr "github.com/EfosaE/credora-backend/domain/domianerrors"
	"github.com/EfosaE/credora-backend/domain/txmanager"
	"github.com/EfosaE/credora-backend/domain/user"
	"github.com/EfosaE/credora-backend/internal/config"
	"github.com/EfosaE/credora-backend/service"
	"github.com/rs/zerolog"
)

type AuthService struct {
	txManager         txmanager.TxManager
	userRepo          user.UserRepository
	passwordResetRepo auth.PasswordResetRepository
	acctRepo          account.AccountRepository
	tokenService      auth.TokenService
	mailer            service.EmailService
	logger            zerolog.Logger
}

func NewAuthService(
	txManager txmanager.TxManager,
	userRepo user.UserRepository,
	passwordResetRepo auth.PasswordResetRepository,
	tokenService auth.TokenService,
	acctRepo account.AccountRepository,
	mailer service.EmailService,
	logger zerolog.Logger,
) *AuthService {

	serviceLogger := logger.With().
		Str("service", "auth-service").
		Logger()

	return &AuthService{
		txManager:         txManager,
		userRepo:          userRepo,
		passwordResetRepo: passwordResetRepo,
		tokenService:      tokenService,
		acctRepo:          acctRepo,
		mailer:            mailer,
		logger:            serviceLogger,
	}
}

func (s *AuthService) Login(ctx context.Context, accountNumber, password string) (*account.GetUserDetailsWithAccountRow, string, error) {

	logCtx := s.logger.With().
		Str("account_number", accountNumber).
		Logger()

	logCtx.Info().Msg("login attempt")

	u, err := s.acctRepo.GetUserByAccountNumber(ctx, accountNumber)
	if err != nil {
		logCtx.Warn().Err(err).Msg("user not found for login")
		return nil, "", fmt.Errorf("login: %w", domainerr.ErrUserNotFound)
	}

	if !CheckPasswordHash(password, u.Password) {
		logCtx.Warn().Msg("invalid password attempt")
		return nil, "", domainerr.ErrInvalidCredentials
	}

	token, err := s.tokenService.GenerateToken(ctx, auth.TokenPayload{
		UserID:        u.UserId,
		AccountNumber: u.AccountNumber,
		Name:          u.FullName,
	})
	if err != nil {
		logCtx.Error().Err(err).Msg("failed to generate auth token")
		return nil, "", fmt.Errorf("login: %w", err)
	}

	logCtx.Info().Str("user_id", u.UserId.String()).Msg("login successful")
	return u, token, nil
}

func (s *AuthService) RequestPasswordReset(ctx context.Context, email string) error {

	logCtx := s.logger.With().
		Str("email", email).
		Logger()

	logCtx.Info().Msg("requesting password reset")

	u, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		logCtx.Warn().Err(err).Msg("user not found for password reset")
		return nil // swallow to avoid enumeration
	}

	raw, hash, expiresAt, err := GenerateResetToken()
	if err != nil {
		logCtx.Error().Err(err).Msg("failed to generate reset token")
		return nil
	}

	if _, err := s.passwordResetRepo.Create(ctx, u.ID, hash, expiresAt); err != nil {
		logCtx.Error().Err(err).Msg("failed to create password reset record")
		return nil
	}

	resetLink := fmt.Sprintf("%s/reset-password?token=%s&email=%s", config.App.FrontendURL, raw, u.Email)

	logCtx.Info().Str("reset_link", resetLink).Msg("generated password reset link")

	if err := s.mailer.SendPasswordResetEmail(ctx, u.Email, resetLink); err != nil {
		logCtx.Error().Err(err).Msg("failed to send password reset email")
		return err
	}

	logCtx.Info().Msg("password reset email sent successfully")
	return nil
}

func (s *AuthService) ValidatePasswordResetRequest(ctx context.Context, email, token, newPassword string) error {

	logCtx := s.logger.With().Str("email", email).Logger()

	u, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		logCtx.Error().Err(err).Msg("failed to fetch user for password reset validation")
		return err
	}

	reset, err := s.passwordResetRepo.GetActiveToken(ctx, u.ID)
	if err != nil {
		logCtx.Error().Err(err).Msg("failed to fetch active reset token")
		return err
	}

	if !ValidateResetToken(token, reset.TokenHash) {
		logCtx.Warn().Msg("invalid password reset token attempt")
		return fmt.Errorf("invalid or expired reset token")
	}

	if time.Now().After(reset.ExpiresAt) {
		logCtx.Warn().Time("expires_at", reset.ExpiresAt).Msg("expired password reset token used")
		return fmt.Errorf("invalid or expired reset token")
	}

	hashedPassword, err := HashPassword(newPassword)
	if err != nil {
		logCtx.Error().Err(err).Msg("failed to hash new password")
		return err
	}

	err = s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.userRepo.UpdatePassword(txCtx, u.ID, hashedPassword); err != nil {
			logCtx.Error().Err(err).Msg("failed to update user password")
			return err
		}

		if err := s.passwordResetRepo.MarkUsed(txCtx, reset.ID, time.Now()); err != nil {
			logCtx.Error().Err(err).Msg("failed to mark reset token as used")
			return err
		}

		return nil
	})

	if err != nil {
		logCtx.Error().Err(err).Msg("transaction failed while resetting password")
		return err
	}

	logCtx.Info().Msg("password reset successfully completed")
	return nil
}
