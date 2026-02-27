//go:build integration
// +build integration

package handler

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EfosaE/credora-backend/infrastructure"
	"github.com/EfosaE/credora-backend/test/fakes"

	// "github.com/EfosaE/credora-backend/internal/db/sqlc"
	"github.com/EfosaE/credora-backend/service"
	accountsvc "github.com/EfosaE/credora-backend/service/account"
	idempotencysvc "github.com/EfosaE/credora-backend/service/idempotency"
	transactionsvc "github.com/EfosaE/credora-backend/service/transaction"
	"github.com/EfosaE/credora-backend/test"
	"github.com/stretchr/testify/require"
)

func TestMonnifyWebhook_IdempotencyRecorded(t *testing.T) {
	ctx := context.Background()

	// Spin up Postgres test DB
	pool, cleanup := test.SetupTestDB(t)
	defer cleanup()

	// queries := sqlc.New(pool)

	mockQueue := &fakes.MockQueueClient{}
	mockEvtBus := &fakes.MockEventBus{}
	testLogger := test.SetupTestLogger()

	// ----- Setup dependencies -----
	idemRepo := infrastructure.NewSqlcIdempotencyRepository(pool)
	acctRepo := infrastructure.NewSqlcAccountRepository(pool)
	trxRepo := infrastructure.NewSqlcTransactionRepository(pool)
	// mockMonnifyRepo := fakes.MockMonnifyRepo{}

	idemSvc := idempotencysvc.NewIdempotencyService(idemRepo)
	acctSvc := accountsvc.NewAccountService(acctRepo, testLogger, mockEvtBus)
	trxSvc := transactionsvc.NewTransactionService(trxRepo, testLogger, mockEvtBus)
	mockMonnifyRepo := &fakes.MockMonnifyRepo{}
	monnifySvc := &service.MonnifyService{
		MonnifyRepo: mockMonnifyRepo,
	}

	handler := NewWebHookHandler(acctSvc, trxSvc, monnifySvc, idemSvc, mockQueue)

	// ----- Build test server -----
	srv := httptest.NewServer(http.HandlerFunc(handler.HandleMonnifyWebhook))
	defer srv.Close()

	// ----- Fake payload -----
	payload := `{
        "eventType":"SUCCESSFUL_TRANSACTION",
        "eventData": {
            "paymentReference":"REF12345",
            "amountPaid":5000,
            "paidOn":"2025-11-26T10:00:00"
        }
    }`

	// ----- Send webhook -----
	req, _ := http.NewRequest("POST", srv.URL, bytes.NewBufferString(payload))
	req.Header.Set("monnify-signature", "anything")
	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, 200, res.StatusCode)

	// ----- Assert: idempotency was recorded -----
	exists, err := idemSvc.Exists(ctx, "monnify:REF12345")
	require.NoError(t, err)
	require.True(t, exists, "idempotency key should exist after successful enqueue")

	// ----- Send webhook again -----
	req2, _ := http.NewRequest("POST", srv.URL, bytes.NewBufferString(payload))
	req2.Header.Set("monnify-signature", "anything")
	res2, err := http.DefaultClient.Do(req2)
	require.NoError(t, err)
	require.Equal(t, 200, res2.StatusCode)

	body, _ := io.ReadAll(res2.Body)
	require.Contains(t, string(body), "duplicate")
}
