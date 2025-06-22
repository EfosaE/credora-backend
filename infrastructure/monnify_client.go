package infrastructure

// internal/infrastructure/monnify_client.go

import (
	// "bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"


	"github.com/EfosaE/credora-backend/domain/monnify"
)

type MonnifyClient struct {
	config *monnify.MonnifyConfig
	client *http.Client
}

func NewMonnifyClient(config *monnify.MonnifyConfig, client *http.Client) *MonnifyClient {

	return &MonnifyClient{
		config: config,
		client: client,
	}
}

// Implement MonnifyRepository
func (m *MonnifyClient) Authenticate() error {
	url := fmt.Sprintf("%s/api/v1/auth/login", m.config.BaseURL)

	// Encode credentials
	authStr := fmt.Sprintf("%s:%s", m.config.ApiKey, m.config.SecretKey)
	authHeader := base64.StdEncoding.EncodeToString([]byte(authStr))

	// Build the request
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", authHeader))
	req.Header.Set("Content-Type", "application/json")

	// Use injected http.Client
	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("authentication request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var body map[string]any
		json.NewDecoder(resp.Body).Decode(&body)
		return fmt.Errorf("auth failed with status %d: %+v", resp.StatusCode, body)
	}

	var authResp monnify.MonnifyAuthResponse

	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return fmt.Errorf("failed to decode auth response: %w", err)
	}

	if authResp.ResponseCode != "0" {
		return errors.New("monnify authentication failed: " + authResp.ResponseMessage)
	}

	// Save token
	m.config.Token = authResp.ResponseBody.AccessToken
	return nil
}

func (m *MonnifyClient) CreateReservedAccount(req *monnify.CreateCustomerRequest) (*monnify.CreateCustomerResponse, error) {
	// actual HTTP call to Monnify
	return &monnify.CreateCustomerResponse{}, nil
}

func (m *MonnifyClient) ValidateWebhookSignature(body []byte, signature string) bool {
	// actual signature validation logic
	return true
}
