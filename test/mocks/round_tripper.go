package mocks

import "net/http"

// --- Mock RoundTripper ---

type MockRoundTripper struct {
	ReqFn func(req *http.Request) (*http.Response, error)
}

func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.ReqFn(req)
}