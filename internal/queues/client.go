package queues

import (
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
)

type Queue interface {
	EnqueueWelcomeEmail(payload WelcomeEmailPayload) error
	EnqueueAccountNumberEmail(payload AccountNumberEmailPayload) error
}

type QueueClient struct {
	client *asynq.Client
}

func NewQueueClient(redisAddr string) *QueueClient {
	return &QueueClient{
		client: asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr}),
	}
}

func (q *QueueClient) EnqueueWelcomeEmail(payload WelcomeEmailPayload) error {
	b, _ := json.Marshal(payload)
	task := asynq.NewTask(TypeWelcomeEmail, b)

	_, err := q.client.Enqueue(task,
		asynq.MaxRetry(20),            // exponential retry
		asynq.Timeout(30),             // 30s execution timeout
		asynq.Retention(24*time.Hour), // keep job result
	)
	return err
}

func (q *QueueClient) EnqueueAccountNumberEmail(payload AccountNumberEmailPayload) error {
	b, _ := json.Marshal(payload)
	task := asynq.NewTask(TypeAccountNumberEmail, b)

	_, err := q.client.Enqueue(task,
		asynq.MaxRetry(20),            // exponential retry
		asynq.Timeout(30),             // 30s execution timeout
		asynq.Retention(24*time.Hour), // keep job result
	)
	return err
}

// func (q *QueueClient) EnqueueExternalTransfer(payload ExternalTransferPayload) error {
//     b, _ := json.Marshal(payload)
//     task := asynq.NewTask(TaskProcessExternalTransfer, b)

//     _, err := q.client.Enqueue(task,
//         asynq.MaxRetry(20),                // exponential retry
//         asynq.Timeout(30),                 // 30s execution timeout
//         asynq.Retention(24*time.Hour),     // keep job result
//     )
//     return err
// }

// func (q *QueueClient) EnqueueInternalTransfer(payload InternalTransferPayload) error { ... }

// func (q *QueueClient) EnqueueVACredit(payload VACreditPayload) error { ... }

// func (q *QueueClient) EnqueueEmail(payload EmailPayload) error { ... }
