package simulator

type TransferRequest struct {
	Amount            float64 `json:"amount" validate:"required,gt=0"`               // Amount sent must be greater than 0
	Narration         string  `json:"narration" validate:"required"`                 // Required narration
	SenderName        string  `json:"sender_name" validate:"required"`               // Required sender name
	SenderAccount     string  `json:"sender_account" validate:"required,len=10"`     // NUBAN is typically 10 digits
	SenderBankCode    string  `json:"sender_bank_code" validate:"required,len=3"`    // Bank code usually 3-digit (customize if needed)
	RecipientAccount  string  `json:"recipient_account" validate:"required,len=10"`  // Recipient account number
	RecipientBankName string  `json:"recipient_bank_name" validate:"omitempty"`      // Optional
	ForceFail         bool    `json:"force_fail"`                                   // Optional for simulating failure
}
