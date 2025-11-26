package accountsvc_test

import (
	"context"
	"testing"

	"github.com/EfosaE/credora-backend/domain/account"
	accountsvc "github.com/EfosaE/credora-backend/service/account"
	"github.com/EfosaE/credora-backend/test/mocks"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestCreditUserBalance(t *testing.T) {
	mockRepo := &mocks.MockAcctRepo{}

	amount := decimal.NewFromFloat(1500.0)
	acctNum := "1234567890"
	acctID := uuid.New() // generate a valid UUID

	expectedResp := &account.CreditAcctResp{
		AcctId:  acctID,
		Balance: amount,
	}

	// Mock the CreditFunc behavior
	mockRepo.CreditFunc = func(ctx context.Context, amt decimal.Decimal, accNum string) (*account.CreditAcctResp, error) {
		require.Equal(t, amount, amt)
		require.Equal(t, acctNum, accNum)
		return expectedResp, nil
	}

	service := &accountsvc.AccountService{AcctRepo: mockRepo}

	resp, err := service.CreditUserBalance(context.Background(), amount, acctNum)
	require.NoError(t, err)
	require.Equal(t, expectedResp, resp)
}
