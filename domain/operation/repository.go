package operation

import (
	"context"
)

type OperationRepository interface {
	InternalTransfer(ctx context.Context, req *InternalTransferReq) error
}


