package queues

import (
	"encoding/json"
	"time"

	"github.com/EfosaE/credora-backend/domain/operation"
	"github.com/hibiken/asynq"
)

type Queue interface {
	EnqueueWelcomeEmail(payload WelcomeEmailPayload) error
	EnqueueAccountNumberEmail(payload AccountNumberEmailPayload) error
	EnqueueInternalTransfer(payload *operation.InternalTransferReq) error
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
	// Email tasks (non-critical)
	DefaultTaskOptions = QueueOptions{
		Queue:     "default",
		MaxRetry:  5,
		Timeout:   60 * time.Second,
		Deadline:  10 * time.Minute,
		Retention: 1 * time.Hour,
	}

	// Critical tasks (account creation, etc.)
	CriticalTaskOptions = QueueOptions{
		Queue:     "critical",
		MaxRetry:  10,
		Timeout:   60 * time.Second,
		Deadline:  10 * time.Minute,
		Retention: 1 * time.Hour,
	}

	// High-risk, long-running tasks: internal transfers
	InternalTransferOptions = QueueOptions{
		Queue:     "critical",
		MaxRetry:  10,
		Timeout:   5 * time.Minute,  // give handler more time
		Deadline:  7 * time.Minute,  // must be > timeout
		Retention: 24 * time.Hour,   // auditing and replay
	}
)

// --- Generic enqueue function ---

func (q *QueueClient) enqueue(taskType string, payload any, opts QueueOptions) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	task := asynq.NewTask(taskType, b)

	_, err = q.client.Enqueue(
		task,
		asynq.Queue(opts.Queue),
		asynq.MaxRetry(opts.MaxRetry),
		asynq.Timeout(opts.Timeout),
		asynq.Deadline(time.Now().Add(opts.Deadline)),
		asynq.Retention(opts.Retention),
	)

	return err
}

// --- Public API exposable methods ---

func (q *QueueClient) EnqueueWelcomeEmail(payload WelcomeEmailPayload) error {
	return q.enqueue(TypeWelcomeEmail, payload, DefaultTaskOptions)
}

func (q *QueueClient) EnqueueAccountNumberEmail(payload AccountNumberEmailPayload) error {
	return q.enqueue(TypeAccountNumberEmail, payload, CriticalTaskOptions)
}

func (q *QueueClient) EnqueueInternalTransfer(payload *operation.InternalTransferReq) error {
	return q.enqueue(TypeInternalTransfer, payload, InternalTransferOptions)
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
