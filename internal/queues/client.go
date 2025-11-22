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

func (q *QueueClient) enqueueDefault(taskType string, payload any) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	task := asynq.NewTask(taskType, b)

	_, err = q.client.Enqueue(task,
		asynq.MaxRetry(5),
		asynq.Timeout(60*time.Second),
		asynq.Queue("default"),
		asynq.Deadline(time.Now().Add(10*time.Minute)),
		asynq.Retention(1*time.Hour),
	)

	return err
}

func (q *QueueClient) enqueueCritical(taskType string, payload any) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	task := asynq.NewTask(taskType, b)

	_, err = q.client.Enqueue(task,
		asynq.MaxRetry(10),
		asynq.Timeout(60*time.Second),
		asynq.Queue("critical"),
		asynq.Deadline(time.Now().Add(2*time.Minute)),
		asynq.Retention(1*time.Hour),
	)

	return err
}

func (q *QueueClient) EnqueueWelcomeEmail(payload WelcomeEmailPayload) error {
	return q.enqueueDefault(TypeWelcomeEmail, payload)
}

func (q *QueueClient) EnqueueAccountNumberEmail(payload AccountNumberEmailPayload) error {
	return q.enqueueCritical(TypeAccountNumberEmail, payload)
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
