package handler

import (
	// "bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/EfosaE/credora-backend/domain/monnify"
	"github.com/EfosaE/credora-backend/domain/operation"
	"github.com/EfosaE/credora-backend/domain/transaction"
	"github.com/EfosaE/credora-backend/domain/webhook"
	"github.com/EfosaE/credora-backend/internal/queues"
	"github.com/EfosaE/credora-backend/internal/response"
	"github.com/EfosaE/credora-backend/internal/utils"
	"github.com/EfosaE/credora-backend/service"
	accountsvc "github.com/EfosaE/credora-backend/service/account"
	idempotencysvc "github.com/EfosaE/credora-backend/service/idempotency"
	transactionsvc "github.com/EfosaE/credora-backend/service/transaction"
)

type WebHookHandler struct {
	acctSvc    *accountsvc.AccountService
	trxSvc     *transactionsvc.TransactionService
	monnifySvc *service.MonnifyService
	idemSvc    *idempotencysvc.IdempotencyService
	queue      queues.Queue
}

func NewWebHookHandler(acctSvc *accountsvc.AccountService, trxSvc *transactionsvc.TransactionService, monnify *service.MonnifyService, idemSvc *idempotencysvc.IdempotencyService, queue queues.Queue) *WebHookHandler {
	return &WebHookHandler{acctSvc: acctSvc, trxSvc: trxSvc, monnifySvc: monnify, idemSvc: idemSvc, queue: queue}
}

func (h *WebHookHandler) HandleMonnifyWebhook(w http.ResponseWriter, r *http.Request) {
	signature := r.Header.Get("monnify-signature")

	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		response.SendError(w, r, response.BadRequest(err, "Could not read webhook body"))
		return
	}

	// -----------------------------
	// Validate signature
	// -----------------------------
	if !h.monnifySvc.ValidateWebhook(rawBody, signature) {
		http.Error(w, "Invalid webhook signature", http.StatusUnauthorized)
		return
	}

	// -----------------------------
	// Parse wrapper payload
	// -----------------------------
	var payload monnify.MonnifyWebhook
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		response.SendError(w, r, response.BadRequest(err, "Invalid JSON payload"))
		return
	}

	utils.PrintJSON(map[string]any{
		"msg":     "Webhook received",
		"payload": payload,
	})

	// -----------------------------
	// Route event type
	// -----------------------------
	switch payload.EventType {

	case webhook.EventSuccessfulTransaction:
		h.handleSuccessfulTransaction(w, r, rawBody, payload)

	default:
		log.Printf("Unhandled webhook event type: %s", payload.EventType)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ignored event type")
	}
}


func (h *WebHookHandler) handleSuccessfulTransaction(
	w http.ResponseWriter,
	r *http.Request,
	rawBody []byte,
	payload monnify.MonnifyWebhook,
) {

	// -----------------------------
	// Parse EventData
	// -----------------------------
	var tx webhook.SuccessfulTransaction
	if err := json.Unmarshal(payload.EventData, &tx); err != nil {
		response.SendError(w, r, response.BadRequest(err, "Invalid EventData format"))
		return
	}

	// Parse paidOn
	if paidOn, err := monnify.ParseMonnifyTime(tx.PaidOnRaw); err == nil {
		tx.PaidOn = paidOn
	}

	// -----------------------------
	// Idempotency check
	// -----------------------------
	idemKey := "monnify:" + tx.PaymentReference

	exists, err := h.idemSvc.Exists(r.Context(), idemKey)
	if err != nil {
		response.SendError(w, r, response.InternalServerError(err, "Failed to check idempotency"))
		return
	}

	if exists {
		log.Println("Duplicate webhook ignored:", tx.PaymentReference)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "duplicate")
		return
	}

	// -----------------------------
	// Queue worker payload
	// -----------------------------
	inboundReq := &webhook.InboundTransferPayload{
		TransactionDetails: tx,
	}

	if err := h.queue.EnqueueWebhookInboundTransfer(inboundReq); err != nil {

		log.Printf("Failed to enqueue inbound transfer. Writing fallback to DB. error=%v", err)

		saveErr := h.idemSvc.AddToIdempotencyTable(
			r.Context(),
			idemKey,
			operation.OperationTypeWebhookInboundTransfer,
			rawBody,
			transaction.StatusFailed,
		)

		if saveErr != nil {
			log.Printf("FAILED to persist failed webhook event: %v", saveErr)
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "received-but-queue-failed")
		return
	}

	// -----------------------------
	// Save idempotency as pending
	// -----------------------------
	if err := h.idemSvc.AddToIdempotencyTable(
		r.Context(),
		idemKey,
		operation.OperationTypeWebhookInboundTransfer,
		rawBody,
		transaction.StatusPending,
	); err != nil {
		log.Printf("Failed to store idempotency key: %v", err)
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "received")
}