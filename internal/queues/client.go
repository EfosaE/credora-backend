package queues

import (
	"encoding/json"
	"time"

	"github.com/EfosaE/credora-backend/domain/webhook"
	"github.com/hibiken/asynq"
)

type Queue interface {
	EnqueueWelcomeEmail(payload WelcomeEmailPayload) error
	EnqueueAccountNumberEmail(payload AccountNumberEmailPayload) error
	EnqueueInternalTransfer(payload InternalTransferTaskPayload) error
	EnqueueWebhookInboundTransfer(payload *webhook.InboundTransferPayload) error
}

type QueueClient struct {
	client *asynq.Client
}

func NewQueueClient(redisAddr string) *QueueClient {
	return &QueueClient{
		client: asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr}),
	}
}

type QueueOptions struct {
	Queue     string
	MaxRetry  int
	Timeout   time.Duration
	Deadline  time.Duration
	Retention time.Duration
}

// --- Task Option Profiles ---

var (
	// Default tasks (emails, notifications, analytics, webhooks)
	DefaultTaskOptions = QueueOptions{
		Queue:     "default",
		MaxRetry:  5,
		Timeout:   60 * time.Second,
		Retention: 1 * time.Hour,
	}

	// Critical tasks (account creation, state transitions, payments)
	CriticalTaskOptions = QueueOptions{
		Queue:     "critical",
		MaxRetry:  3,
		Timeout:   90 * time.Second,
		Retention: 1 * time.Hour,
	}
)

// --- Generic enqueue function ---

func (q *QueueClient) enqueue(taskType string, payload any, opts QueueOptions) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	task := asynq.NewTask(taskType, b)

	var taskOpts []asynq.Option

	if opts.Queue != "" {
		taskOpts = append(taskOpts, asynq.Queue(opts.Queue))
	}

	if opts.MaxRetry > 0 {
		taskOpts = append(taskOpts, asynq.MaxRetry(opts.MaxRetry))
	}

	if opts.Timeout > 0 {
		taskOpts = append(taskOpts, asynq.Timeout(opts.Timeout))
	}

	if opts.Deadline > 0 {
		taskOpts = append(taskOpts, asynq.Deadline(time.Now().Add(opts.Deadline)))
	}

	if opts.Retention > 0 {
		taskOpts = append(taskOpts, asynq.Retention(opts.Retention))
	}

	_, err = q.client.Enqueue(task, taskOpts...)
	return err
}

// --- Public API exposable methods ---

func (q *QueueClient) EnqueueWelcomeEmail(payload WelcomeEmailPayload) error {
	return q.enqueue(TypeWelcomeEmail, payload, DefaultTaskOptions)
}

func (q *QueueClient) EnqueueAccountNumberEmail(payload AccountNumberEmailPayload) error {
	return q.enqueue(TypeAccountNumberEmail, payload, CriticalTaskOptions)
}

func (q *QueueClient) EnqueueInternalTransfer(payload InternalTransferTaskPayload) error {
	return q.enqueue(TypeInternalTransfer, payload, CriticalTaskOptions)
}

func (q *QueueClient) EnqueueWebhookInboundTransfer(payload *webhook.InboundTransferPayload) error {
	return q.enqueue(TypeWebhookInboundTransfer, payload, CriticalTaskOptions)
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
