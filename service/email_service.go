package service

import (
	"context"

	"github.com/EfosaE/credora-backend/domain/email"
	"github.com/EfosaE/credora-backend/domain/user"
)

type EmailService interface {
	SendAccountNumberEmail(ctx context.Context, to, bank, accountNumber string) error
	SendPasswordResetEmail(ctx context.Context, to, resetLink string) error
	SendWelcomeEmail(ctx context.Context, user user.User) error
}

type EmailServiceImpl struct {
	sender email.EmailSender
}

func NewEmailService(sender email.EmailSender) *EmailServiceImpl {
	return &EmailServiceImpl{sender: sender}
}

func (s *EmailServiceImpl) SendAccountNumberEmail(ctx context.Context, to, bank, accountNumber string) error {
	html, err := email.RenderTemplate("account_email", map[string]string{
		"Bank":          bank,
		"AccountNumber": accountNumber,
	})

	if err != nil {
		return err
	}

	return s.sender.SendEmail(ctx, email.SendEmailRequest{
		From:    "accounts@credora.com",
		To:      to,
		Subject: "Your Virtual Account Details",
		Html:    html,
	})
}

func (s *EmailServiceImpl) SendWelcomeEmail(ctx context.Context, user user.User) error {
	html, err := email.RenderTemplate("welcome_email", map[string]string{
		"UserName": user.Name,
		"CompanyName": "Credora",
		"AccountID":   user.ID.String(),
		"LoginURL": "https://vaultix.osamwonyiefosa02.workers.dev",
	})

	if err != nil {
		return err
	}

	return s.sender.SendEmail(ctx, email.SendEmailRequest{
		From:    "support@credora.com",
		To:      user.Email,
		Subject: "Welcome to Credora",
		Html:    html,
	})
}

func (s *EmailServiceImpl) SendPasswordResetEmail(ctx context.Context, to, resetLink string) error {
	html, err := email.RenderTemplate("password_reset", map[string]string{
		"ResetLink": resetLink,
	})
	if err != nil {
		return err
	}

	return s.sender.SendEmail(ctx, email.SendEmailRequest{
		From:    "support@credora.com",
		To:      to,
		Subject: "Reset Your Password",
		Html:    html,
	})
}
