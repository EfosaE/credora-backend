package operationsvc_test

import (
	"context"
	"testing"

	"github.com/EfosaE/credora-backend/domain/account"
	"github.com/EfosaE/credora-backend/domain/event"
	"github.com/EfosaE/credora-backend/domain/operation"
	"github.com/EfosaE/credora-backend/domain/transaction"
	operationsvc "github.com/EfosaE/credora-backend/service/operation"
	"github.com/EfosaE/credora-backend/test"
	"github.com/EfosaE/credora-backend/test/fakes"
	"github.com/EfosaE/credora-backend/test/mocks"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestInternalTransfer_Success_PublishesEvent(t *testing.T) {
	ctx := context.Background()

	mockAcct := new(mocks.MockAccountRepository)
	mockIdem := new(mocks.MockIdempotencyRepository)
	mockBus := new(mocks.MockEventBus)
	testLogger := test.SetupTestLogger()
	txManager := fakes.NewNoOpTxManager()

	svc := operationsvc.NewOperationService(
		txManager,
		mockAcct,
		nil,
		mockIdem,
		testLogger,
		mockBus,
	)

	amount := decimal.NewFromInt(2000)

	// Expect idempotency PROCESSING
	mockIdem.
		On("Upsert",
			mock.Anything,
			"idem-key",
			operation.OperationTypeInternalTransfer,
			mock.Anything,
			transaction.StatusProcessing,
		).
		Return(nil).
		Once()

	// Expect money transfer call
	mockAcct.
		On("InternalMoneyTransfer",
			mock.Anything,
			amount,
			"111",
			"222",
		).
		Return(&account.InternalTransferResp{
			FromAccountId: uuid.New(),
			ToAccountId:   uuid.New(),
			FromUserId:    uuid.New(),
			ToUserId:      uuid.New(),
		}, nil).
		Once()

	//  Expect idempotency SUCCESS
	mockIdem.
		On("Upsert",
			mock.Anything,
			"idem-key",
			operation.OperationTypeInternalTransfer,
			mock.Anything,
			transaction.StatusSuccess,
		).
		Return(nil).
		Once()

	// 4️⃣ Expect event publish
	mockBus.
		On("Publish",
			mock.Anything,
			event.StreamTransferEvents,
			event.EventInternalTransferSuccess,
			mock.Anything,
		).
		Return(nil).
		Once()

	req := &operation.InternalTransferReq{
		FromAcctNum:    "111",
		ToAcctNum:      "222",
		Amount:         amount,
		IdempotencyKey: "idem-key",
	}

	err := svc.InternalTransfer(ctx, req)
	require.NoError(t, err)

	mockAcct.AssertExpectations(t)
	mockIdem.AssertExpectations(t)
	mockBus.AssertExpectations(t)
}
