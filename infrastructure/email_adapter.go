package infrastructure

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/EfosaE/credora-backend/domain/email"
	"github.com/EfosaE/credora-backend/internal/config"
	"github.com/wneessen/go-mail"
	"golang.org/x/time/rate"
)

type client struct {
	Host     string
	Port     int
	Username string
	Password string
}
type EmailAdapter struct {
	client
	limiter *rate.Limiter
}

func NewEmailAdapter() *EmailAdapter {
	port, err := strconv.Atoi(config.App.MailtrapPort)
	if err != nil {
		log.Fatal("Invalid port:", err)
	}
	return &EmailAdapter{
		client: client{
			Host:     config.App.MailtrapHost,
			Port:     port,
			Username: config.App.MailtrapUser,
			Password: config.App.MailtrapPass,
		},
		limiter: rate.NewLimiter(
			rate.Limit(config.App.Email.Rate), // refill email token per second
			config.App.Email.Burst,            // max burst : store this amount at a time.
		),
	}
}

// func (r *ResendAdapter) SendEmail(ctx context.Context, req email.SendEmailRequest) (*resend.SendEmailResponse, error) {
// 	sent, err := r.client.Emails.Send(&resend.SendEmailRequest{
// 		From:    req.From,
// 		To:      []string{req.To},
// 		Subject: req.Subject,
// 		Html:    req.Html, // Already rendered in EmailService
// 		ReplyTo: req.ReplyTo,
// 	})
// 	fmt.Println(err)
// 	utils.PrintJSON(sent)

// 	if err != nil {
// 		return nil, err
// 	}
// 	return sent, nil
// }

func (s *EmailAdapter) SendEmail(ctx context.Context, req email.SendEmailRequest) error {
	if err := s.limiter.Wait(ctx); err != nil {
		return err
	}
	msg := mail.NewMsg()
	if err := msg.From(req.From); err != nil {
		return fmt.Errorf("failed to set from address: %w", err)
	}
	if err := msg.To(req.To); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}
	msg.Subject(req.Subject)
	msg.SetBodyString(mail.TypeTextHTML, req.Html)

	client, err := mail.NewClient(
		s.Host,
		mail.WithPort(s.Port),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(s.Username),
		mail.WithPassword(s.Password),
	)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}

	if err := client.DialAndSend(msg); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}

// To ensure this repo satisfies the interface defined in the domain.
var _ email.EmailSender = (*EmailAdapter)(nil)
