package infrastructure

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/EfosaE/credora-backend/domain/monnify"
	"github.com/EfosaE/credora-backend/test/mocks"
	"github.com/EfosaE/credora-backend/test/stubs"
	"github.com/stretchr/testify/assert"
)

func TestAuthenticate_Success(t *testing.T) {
	body, _ := json.Marshal(stubs.StubAuthenticateResponse)
	mockTransport := &mocks.MockRoundTripper{
		ReqFn: func(req *http.Request) (*http.Response, error) {
			// respBody := `{
			// 	"responseCode": "0",
			// 	"responseMessage": "success",
			// 	"responseBody": {
			// 		"accessToken": "mocked-token"
			// 	}
			// }`
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(string(body))),
				Header:     make(http.Header),
			}, nil
		},
	}

	httpClient := &http.Client{Transport: mockTransport}
	client := NewMonnifyClient(&monnify.MonnifyConfig{}, httpClient)

	err := client.Authenticate()
	assert.NoError(t, err)
	assert.Equal(t, "mocked-token", client.config.Token)
}

