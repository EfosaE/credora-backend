package service

import (
	"github.com/EfosaE/credora-backend/domain/monnify"
	"github.com/rs/zerolog"
)

type MonnifyService struct {
	MonnifyRepo monnify.MonnifyRepository
	logger      zerolog.Logger
}

func NewMonnifyService(
	monnifyRepo monnify.MonnifyRepository,
	logger zerolog.Logger,
) *MonnifyService {

	serviceLogger := logger.With().
		Str("service", "monnify-service").
		Logger()

	return &MonnifyService{
		MonnifyRepo: monnifyRepo,
		logger:      serviceLogger,
	}
}

func (s *MonnifyService) CreateCustomer(customer *monnify.CreateCRAParams) (*monnify.CreateCRAResponse, error) {

	log := s.logger.With().
		Str("account_reference", customer.AccountReference).
		Str("customer_name", customer.AccountName).
		Logger()

	log.Info().Msg("creating monnify reserved account")

	resp, err := s.MonnifyRepo.CreateReservedAccount(customer)
	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to create monnify reserved account")
		return nil, err
	}

	log.Info().
		Str("reservation_reference", resp.ResponseBody.ReservationReference).
		Msg("monnify reserved account created successfully")

	return resp, nil
}

func (s *MonnifyService) DeleteCustomer(acctRef string) (*monnify.CreateCRAResponse, error) {

	log := s.logger.With().
		Str("account_reference", acctRef).
		Logger()

	log.Info().Msg("deleting monnify reserved account")

	resp, err := s.MonnifyRepo.DeleteReservedAccount(acctRef)
	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to delete monnify reserved account")
		return nil, err
	}

	log.Info().Msg("monnify reserved account deleted successfully")
	return resp, nil
}

func (s *MonnifyService) ValidateWebhook(body []byte, signature string) bool {

	log := s.logger.With().
		Int("payload_size_bytes", len(body)).
		Logger()

	log.Info().Msg("validating monnify webhook signature")

	valid := s.MonnifyRepo.ValidateWebhookSignature(body, signature)
	if !valid {
		log.Warn().Msg("invalid monnify webhook signature")
		return false
	}

	log.Info().Msg("monnify webhook signature validated successfully")
	return true
}
