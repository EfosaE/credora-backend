package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/EfosaE/credora-backend/domain/monnify"
	"github.com/EfosaE/credora-backend/domain/webhook"
	"github.com/EfosaE/credora-backend/internal/pgerrors"
	"github.com/EfosaE/credora-backend/internal/response"
	"github.com/EfosaE/credora-backend/internal/utils"
	accountsvc "github.com/EfosaE/credora-backend/service/account"
	"github.com/shopspring/decimal"
)

type WebHookHandler struct {
	acctSvc *accountsvc.AccountService
}

func NewWebHookHandler(acctSvc *accountsvc.AccountService) *WebHookHandler {
	return &WebHookHandler{acctSvc: acctSvc}
}

func (h WebHookHandler) HandleMonnifyWebhook(w http.ResponseWriter, r *http.Request) {
	var payload monnify.MonnifyWebhook
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if payload.EventType == webhook.EventSuccessfulTransaction {
		var tx webhook.SuccessfulTransaction
		if err := json.Unmarshal(payload.EventData, &tx); err != nil {
			http.Error(w, "Invalid EventData", http.StatusBadRequest)
			return
		}

		fmt.Printf("✅ Payment received:\n")
		utils.PrintJSON(tx)

		numericStr, _ := decimal.NewFromString(tx.SettlementAmount)

		// 💰 Lookup user by paymentReference or account reference
		result, err := h.acctSvc.CreditUserBalance(r.Context(), numericStr, tx.DestinationAccountInfo.AccountNumber)
		fmt.Println(err)
		if err != nil {
			if pgerrors.HandlePGError(w, r, err) {
				return // already responded
			}
			response.SendError(w, r, response.InternalServerError(err, "Failed to credit wallet"))
			return
		}

		// w.WriteHeader(http.StatusOK)
		// fmt.Fprint(w, "received")
		response.SendSuccess(w, r, response.OK(result, "Account Credited successfully"))
		return
	}

	// Ignore other events for now
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "ignored")
}
