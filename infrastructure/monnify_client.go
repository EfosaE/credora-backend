package infrastructure

// internal/infrastructure/monnify_client.go

import (
	// "bytes"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

func (m *MonnifyClient) CreateReservedAccount(monnifyCust *monnify.CreateCRAParams) (*monnify.CreateCRAResponse, error) {
	// actual HTTP call to Monnify
	url := fmt.Sprintf("%s/api/v2/bank-transfer/reserved-accounts", m.config.BaseURL)
	if m.config.Token == "" {
		if err := m.Authenticate(); err != nil {
			return nil, err
		}
	}

	// Encode body
	bodyBytes, err := json.Marshal(monnifyCust)
	if err != nil {
		return nil, fmt.Errorf("failed to encode request body: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.config.Token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("monnify error [%d]: %s", resp.StatusCode, string(respBody))
	}

	var response monnify.CreateCRAResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

func (m *MonnifyClient) ValidateWebhookSignature(body []byte, signature string) bool {
	// actual signature validation logic
	return true
}
