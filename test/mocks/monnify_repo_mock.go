package mocks

import (
	"github.com/EfosaE/credora-backend/domain/monnify"
)

type MockMonnifyRepo struct {
	CreateReservedAccountFunc func(req *monnify.CreateCRAParams) (*monnify.CreateCRAResponse, error)
}

func (m *MockMonnifyRepo) CreateReservedAccount(req *monnify.CreateCRAParams) (*monnify.CreateCRAResponse, error) {
	return m.CreateReservedAccountFunc(req)
}


func (m *MockMonnifyRepo) Authenticate() error {
	// Mock implementation for Authenticate
	return nil
}

func (m *MockMonnifyRepo) ValidateWebhookSignature(body []byte, signature string) bool {
	// Mock implementation for ValidateWebhookSignature
	return true
}