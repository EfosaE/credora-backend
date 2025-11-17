// domain/operation/converters.go
package operation

import (
	"github.com/shopspring/decimal"
)

func (dto *InternalTransferDTO) ToDomain() (*InternalTransferReq, error) {

	amount, err := decimal.NewFromString(dto.Amount)
	if err != nil {
		return nil, err
	}

	return &InternalTransferReq{
		FromAcctNum: dto.FromAccount,
		ToAcctNum:   dto.ToAccount,
		Amount:      amount,
	}, nil
}
