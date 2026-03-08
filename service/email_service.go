package service

import (
	"context"

	"github.com/EfosaE/credora-backend/domain/email"
	"github.com/EfosaE/credora-backend/domain/event"
	"github.com/EfosaE/credora-backend/domain/monnify"
	"github.com/EfosaE/credora-backend/domain/user"
	"github.com/rs/zerolog"
)

type EmailService interface {
	SendAccountNumberEmail(ctx context.Context, to string, accounts []monnify.ReservedAccount) error
	SendPasswordResetEmail(ctx context.Context, to, resetLink string) error
	SendWelcomeEmail(ctx context.Context, user user.User) error
}

type EmailServiceImpl struct {
	sender   email.EmailSender
	eventBus event.EventBus
	logger   zerolog.Logger
}

func NewEmailService(
	sender email.EmailSender,
	eventBus event.EventBus,
	logger zerolog.Logger,
) *EmailServiceImpl {

	serviceLogger := logger.With().
		Str("service", "email-service").
		Logger()

	return &EmailServiceImpl{
		sender:   sender,
		eventBus: eventBus,
		logger:   serviceLogger,
	}
}

func (s *EmailServiceImpl) SendAccountNumberEmail(ctx context.Context, to string, accounts []monnify.ReservedAccount) error {
	log := s.logger.With().
		Str("email", to).
		Int("accounts", len(accounts)).
		Logger()

	log.Info().Msg("sending account number email")

	html, err := email.RenderTemplate("account_email", map[string]any{
		"Accounts": accounts,
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to render account email template")
		return err
	}

	if err := s.sender.SendEmail(ctx, email.SendEmailRequest{
		From:    "accounts@credora.com",
		To:      to,
		Subject: "Your Virtual Account Details",
		Html:    html,
	}); err != nil {
		log.Error().Err(err).Msg("failed to send account number email")
		return err
	}

	log.Info().Msg("account number email sent successfully")
	return nil
}

func (s *EmailServiceImpl) SendWelcomeEmail(ctx context.Context, user user.User) error {

	log := s.logger.With().
		Str("email", user.Email).
		Str("user_id", user.ID.String()).
		Logger()

	log.Info().Msg("sending welcome email")

	html, err := email.RenderTemplate("welcome_email", map[string]string{
		"UserName":    user.FullName,
		"CompanyName": "Credora",
		"AccountID":   user.ID.String(),
		"LoginURL":    "https://vaultix.osamwonyiefosa02.workers.dev",
	})

	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to render welcome email template")
		return err
	}

	if err := s.sender.SendEmail(ctx, email.SendEmailRequest{
		From:    "support@credora.com",
		To:      user.Email,
		Subject: "Welcome to Credora",
		Html:    html,
	}); err != nil {
		log.Error().
			Err(err).
			Msg("failed to send welcome email")
		return err
	}

	log.Info().Msg("welcome email sent successfully")
	return nil
}

func (s *EmailServiceImpl) SendPasswordResetEmail(ctx context.Context, to, resetLink string) error {

	log := s.logger.With().
		Str("email", to).
		Logger()

	log.Info().Msg("sending password reset email")

	html, err := email.RenderTemplate("password_email", map[string]string{
		"ResetLink":   resetLink,
		"ExpiresIn":   "10 minutes",
		"CompanyName": "Credora",
	})

	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to render password reset email template")
		return err
	}

	if err := s.sender.SendEmail(ctx, email.SendEmailRequest{
		From:    "support@credora.com",
		To:      to,
		Subject: "Reset Your Password",
		Html:    html,
	}); err != nil {
		log.Error().
			Err(err).
			Msg("failed to send password reset email")
		return err
	}

	log.Info().Msg("password reset email sent successfully")
	return nil
}

// func (s *EmailServiceImpl) SubscribeToUserCreatedEvents(ctx context.Context) error {

// 	log := s.logger.With().
// 		Str("stream", event.StreamUserEvents).
// 		Str("consumer_group", "email-service-group").
// 		Logger()

// 	log.Info().Msg("subscribing to user created events")

// 	consumer := utils.WorkerID("email")

// 	return s.eventBus.Subscribe(
// 		ctx,
// 		event.StreamUserEvents,
// 		"email-service-group",
// 		consumer,
// 		func(ctx context.Context, msg event.EventMessage) error {

// 			eventLog := log.With().
// 				Str("event_type", msg.EventType).
// 				Logger()

// 			// Ignore other event types
// 			if msg.EventType != event.EventUserCreated {
// 				eventLog.Debug().Msg("ignoring unrelated event type")
// 				return nil
// 			}

// 			eventLog.Info().Msg("processing user created event")

// 			var evt event.UserCreatedEvent
// 			if err := json.Unmarshal([]byte(msg.Data), &evt); err != nil {
// 				eventLog.Error().
// 					Err(err).
// 					Msg("failed to decode user created event")
// 				return fmt.Errorf("failed to decode user.created event: %w", err)
// 			}

// 			eventLog = eventLog.With().
// 				Str("email", evt.Email).
// 				Str("user_id", evt.UserID.String()).
// 				Logger()

// 			if err := s.SendAccountNumberEmail(
// 				ctx,
// 				evt.Email,
// 				evt.BankName,
// 				evt.AccountNumber,
// 			); err != nil {
// 				eventLog.Error().
// 					Err(err).
// 					Msg("failed to send account number email from event")
// 				return fmt.Errorf("failed to send account email: %w", err)
// 			}

// 			eventLog.Info().Msg("user created event processed successfully")
// 			return nil
// 		},
// 	)
// }
