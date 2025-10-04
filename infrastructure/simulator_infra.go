// infrastructure/simulator/inmemory_repo.go
//This is a webhook simulator because Monnify doesnt send webhooks in sandbox environment

package infrastructure

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/EfosaE/credora-backend/domain/simulator"
	"github.com/EfosaE/credora-backend/test/stubs"
)

type InMemoryRepo struct {
	WebhookURL string
}

func NewInMemoryRepo(webhookURL string) *InMemoryRepo {
	return &InMemoryRepo{WebhookURL: webhookURL}
}

func (r *InMemoryRepo) SendMoney(ctx context.Context, req simulator.TransferRequest) error {
	event := stubs.BuildSimulatedSuccessEventWbHk(req)

	// Marshal the simulated webhook payload
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// fmt.Println(r.WebhookURL)

	// Send simulated webhook
	resp, err := http.Post(r.WebhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	log.Printf("Simulated transfer of ₦%.2f to %s in %s → webhook status: %s\n", req.Amount, req.RecipientAccount, req.RecipientBankName, resp.Status)
	return nil
}


