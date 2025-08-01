// domain/simulator/repository.go

package simulator

import "context"

type SimulatorRepository interface {
	SendMoney(ctx context.Context, req TransferRequest) error
}
