package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/EfosaE/credora-backend/domain/monnify"
	"github.com/EfosaE/credora-backend/domain/webhook"
	"github.com/EfosaE/credora-backend/internal/utils"
)

type WebHookHandler struct {}

func NewWebHookHandler() *WebHookHandler {
	return &WebHookHandler{}
}



func (h WebHookHandler) HandleMonnifyWebhook(w http.ResponseWriter, r *http.Request) {
	var payload *monnify.MonnifyWebhook
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if payload.EventType == webhook.EventSuccessfulTransaction {
		// Log or save to DB
		fmt.Printf("✅ Payment received:\n")
		utils.PrintJSON(payload)

		// 💰 Lookup user by `paymentReference` and credit their wallet
		// e.g. repo.CreditWallet(payload.EventData.PaymentReference, payload.EventData.AmountPaid)

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "received")
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "ignored")
}
