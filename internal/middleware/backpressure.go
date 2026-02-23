package custmiddleware

import (
	"fmt"
	"net/http"
	"time"

	domainerr "github.com/EfosaE/credora-backend/domain/domianerrors"
	"github.com/EfosaE/credora-backend/internal/response"
	"github.com/hibiken/asynq"
)

type BackpressureMiddleware struct {
	inspector         *asynq.Inspector
	maxQueueSize      int
	queueName         string
	sustainablePerSec float64 // from config.App.Job, passed in at construction
}

func NewBackpressure(inspector *asynq.Inspector, maxSize int, queueName string, sustainablePerSec float64) *BackpressureMiddleware {
	return &BackpressureMiddleware{
		inspector:         inspector,
		sustainablePerSec: sustainablePerSec,
		maxQueueSize:      maxSize,
		queueName:         queueName,
	}
}

func (bp *BackpressureMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		info, err := bp.inspector.GetQueueInfo(bp.queueName)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		pending := info.Pending + info.Active
		if pending >= bp.maxQueueSize {
			// Retry-After: estimated drain time = queue depth / sustainable throughput.
			// Tells the client how long to wait before retrying.
			retryAfterSecs := bp.estimateRetryAfter(pending)

			// X-RateLimit-Reset: Unix timestamp of when the queue should be clear.
			resetAt := time.Now().Add(time.Duration(retryAfterSecs) * time.Second).Unix()

			response.SendError(w, r,
				response.BuildServerError(
					http.StatusServiceUnavailable,
					domainerr.ErrTooManyPendingJobs,
					"System is currently overloaded.",
				).WithHeaders(map[string]string{
					"Retry-After":       fmt.Sprintf("%d", retryAfterSecs),
					"X-RateLimit-Reset": fmt.Sprintf("%d", resetAt),
				}))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// estimateRetryAfter returns how many seconds until the queue drains enough
// to accept new work, based on current depth and sustainable throughput.
func (bp *BackpressureMiddleware) estimateRetryAfter(currentDepth int) int {
	if bp.sustainablePerSec <= 0 {
		return 30 // safe fallback if misconfigured
	}
	// How long to drain from currentDepth down to half of maxQueueSize,
	// giving the client a realistic window rather than telling them to
	// retry exactly at the limit.
	targetDepth := bp.maxQueueSize / 2
	jobsToDrain := currentDepth - targetDepth
	if jobsToDrain <= 0 {
		return 5 // queue is barely over — retry soon
	}
	secs := int(float64(jobsToDrain) / bp.sustainablePerSec)
	if secs < 1 {
		return 1
	}
	return secs
}
