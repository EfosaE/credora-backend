// domain/operation/converters.go
package operation

import (
	"github.com/shopspring/decimal"
)

func (dto *InternalTransferDTO) ToDomain(fromAcctNum string) (*InternalTransferReq, error) {

	amount, err := decimal.NewFromString(dto.Amount)
	if err != nil {
		return nil, err
	}

	return &InternalTransferReq{
		FromAcctNum: fromAcctNum,
		ToAcctNum:   dto.ToAccount,
		Amount:      amount,
	}, nil
}
