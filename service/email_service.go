package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/EfosaE/credora-backend/domain/email"
	"github.com/EfosaE/credora-backend/domain/event"
	"github.com/EfosaE/credora-backend/domain/user"
	"github.com/EfosaE/credora-backend/internal/utils"
)

type EmailService interface {
	SendAccountNumberEmail(ctx context.Context, to, bank, accountNumber string) error
	SendPasswordResetEmail(ctx context.Context, to, resetLink string) error
	SendWelcomeEmail(ctx context.Context, user user.User) error
}

type EmailServiceImpl struct {
	sender   email.EmailSender
	eventBus event.EventBus
}

func NewEmailService(sender email.EmailSender, eventBus event.EventBus) *EmailServiceImpl {
	return &EmailServiceImpl{sender: sender, eventBus: eventBus}
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
		"UserName":    user.Name,
		"CompanyName": "Credora",
		"AccountID":   user.ID.String(),
		"LoginURL":    "https://vaultix.osamwonyiefosa02.workers.dev",
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
	html, err := email.RenderTemplate("password_email", map[string]string{
		"ResetLink":   resetLink,
		"ExpiresIn":   "10 minutes",
		"CompanyName": "Credora",
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

func (s *EmailServiceImpl) SubscribeToUserCreatedEvents(ctx context.Context) error {
	consumer := utils.WorkerID("email")
	return s.eventBus.Subscribe(
		ctx,
		event.StreamUserEvents,
		"email-service-group",
		consumer,
		func(ctx context.Context, msg event.EventMessage) error {

			// Ignore other event types if stream is shared
			if msg.EventType != event.EventUserCreated {
				return nil
			}

			var evt event.UserCreatedEvent
			if err := json.Unmarshal([]byte(msg.Data), &evt); err != nil {
				return fmt.Errorf("failed to decode user.created event: %w", err)
			}

			// Send email REMEMBER TO ENQUEUE THIS
			if err := s.SendAccountNumberEmail(
				ctx,
				evt.Email,
				evt.BankName,
				evt.AccountNumber,
			); err != nil {
				return fmt.Errorf("failed to send account email: %w", err)
			}

			return nil
		},
	)
}
