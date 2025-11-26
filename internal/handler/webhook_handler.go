package handler

import (
	"bytes"
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
	r.Body = io.NopCloser(bytes.NewBuffer(rawBody))

	// -----------------------------
	// Validate signature
	// -----------------------------
	if !h.monnifySvc.ValidateWebhook(rawBody, signature) {
		http.Error(w, "Invalid webhook signature", http.StatusUnauthorized)
		return
	}

	// -----------------------------
	// Parse wrapper body
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

	// Ignore events other than successful transactions
	if payload.EventType != webhook.EventSuccessfulTransaction {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ignored other this type of event (for now)")
		return
	}

	// -----------------------------
	// Parse EventData (transaction)
	// -----------------------------
	var tx webhook.SuccessfulTransaction
	if err := json.Unmarshal(payload.EventData, &tx); err != nil {
		response.SendError(w, r, response.BadRequest(err, "Invalid EventData format"))
		return
	}

	// Parse paidOn field
	if paidOn, err := monnify.ParseMonnifyTime(tx.PaidOnRaw); err == nil {
		tx.PaidOn = paidOn
	}

	// fmt.Println("✅ Successful Monnify payment:")
	// utils.PrintJSON(tx)

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
		fmt.Println("⚠️ Duplicate webhook ignored:", tx.PaymentReference)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "duplicate")
		return
	}

	// -----------------------------
	// Attempt to enqueue for worker
	// -----------------------------
	inboundReq := &webhook.InboundTransferPayload{
		TransactionDetails: tx,
	}

	fmt.Println("Worker Webhook Payload Sent...")
	if err := h.queue.EnqueueWebhookInboundTransfer(inboundReq); err != nil {

		log.Printf("⚠️ Failed to enqueue inbound transfer. Writing fallback to DB. error=%v", err)

		// ------------------------------------------------------
		// STORE failed webhook in DB for retry later
		// ------------------------------------------------------
		saveErr := h.idemSvc.AddToIdempotencyTable(
			r.Context(),
			idemKey,
			operation.OperationTypeWebhookInboundTransfer,
			rawBody, // store original payload
			transaction.StatusFailed,
		)

		if saveErr != nil {
			log.Printf("❌ FAILED to persist failed webhook event: %v", saveErr)
		}

		// Important: Still return 200 so Monnify does not retry forever
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "received-but-queued-failed")
		return
	}

	fmt.Println("Worker Webhook Payload Enqueued Successfully")

	// -----------------------------
	// Add idempotency and mark as pending, the worker would update the status because enqueue succeeded
	// -----------------------------
	if err := h.idemSvc.AddToIdempotencyTable(
		r.Context(),
		idemKey,
		operation.OperationTypeWebhookInboundTransfer,
		rawBody, // store original payload
		transaction.StatusPending,
	); err != nil {
		log.Printf("⚠️ Failed to store idempotency key after enqueue: %v", err)
		// do NOT fail request, do NOT crash
	}
	// if err := h.idemSvc.MarkProcessed(r.Context(), idemKey); err != nil {
	// 	fmt.Printf("⚠️ Failed to store idempotency key after enqueue: %v\n", err)
	// 	log.Printf("⚠️ Failed to store idempotency key after enqueue: %v", err)
	// 	// do NOT fail request, do NOT crash
	// }

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "received")
}
