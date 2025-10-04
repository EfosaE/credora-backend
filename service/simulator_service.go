package service

import (
	"context"
	"errors"

	"github.com/EfosaE/credora-backend/domain/simulator"
	// accountsvc "github.com/EfosaE/credora-backend/service/account"
)

type SimulatorService struct {
	repo simulator.SimulatorRepository
	
}

func NewSimulatorService(repo simulator.SimulatorRepository) *SimulatorService {
	return &SimulatorService{repo: repo}
}

func (s *SimulatorService) SendMoney(ctx context.Context, req *simulator.TransferRequest) error {
	// Optional: validate input
	if req.RecipientAccount == "" || req.Amount <= 100 {
		return errors.New("invalid transfer request")
	}

	// Delegate to repository
	return s.repo.SendMoney(ctx, *req)
}
