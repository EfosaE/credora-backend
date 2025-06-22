package service

import (
	"github.com/EfosaE/credora-backend/domain/logger"
	"github.com/EfosaE/credora-backend/domain/monnify"
)



type MonnifyService struct {
	monnifyRepo monnify.MonnifyRepository
	logger      *logger.Logger
}

func NewMonnifyService(monnifyRepo monnify.MonnifyRepository, logger *logger.Logger) *MonnifyService {
	return &MonnifyService{
		monnifyRepo: monnifyRepo,
		logger:      logger,
	}
}


func (s *MonnifyService) CreateCustomer(customer *monnify.CreateCustomerRequest) (*monnify.CreateCustomerResponse, error) {
	return s.monnifyRepo.CreateReservedAccount(customer)
}
