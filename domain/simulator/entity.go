package simulator

type TransferRequest struct {
	Amount            float64 `json:"amount" validate:"required,gt=0"`             // Amount sent must be greater than 0
	Narration         string  `json:"narration" validate:"required"`               // Required narration
	SenderName        string  `json:"senderName" validate:"required"`              // Required sender name
	SenderAccount     string  `json:"senderAccount" validate:"required,len=10"`    // NUBAN is typically 10 digits
	SenderBankCode    string  `json:"senderBankCode" validate:"required,len=3"`    // Bank code usually 3-digit (customize if needed)
	RecipientAccount  string  `json:"recipientAccount" validate:"required,len=10"` // Recipient account number
	RecipientBankName string  `json:"recipientBankName" validate:"omitempty"`      // Optional
	ForceFail         bool    `json:"forceFail"`                                   // Optional for simulating failure
}
