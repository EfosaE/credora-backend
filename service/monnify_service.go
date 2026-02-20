package service

import (
	"github.com/EfosaE/credora-backend/domain/monnify"
	"github.com/rs/zerolog"
)

type MonnifyService struct {
	MonnifyRepo monnify.MonnifyRepository
	logger      zerolog.Logger
}

func NewMonnifyService(monnifyRepo monnify.MonnifyRepository, logger zerolog.Logger) *MonnifyService {
	return &MonnifyService{
		MonnifyRepo: monnifyRepo,
		logger:      logger,
	}
}

func (s *MonnifyService) CreateCustomer(customer *monnify.CreateCRAParams) (*monnify.CreateCRAResponse, error) {
	return s.MonnifyRepo.CreateReservedAccount(customer)
}

func (s *MonnifyService) DeleteCustomer(acctRef string) (*monnify.CreateCRAResponse, error) {
	return s.MonnifyRepo.DeleteReservedAccount(acctRef)
}

func (s *MonnifyService) ValidateWebhook(body []byte, signature string) bool {
	return s.MonnifyRepo.ValidateWebhookSignature(body, signature)
}
