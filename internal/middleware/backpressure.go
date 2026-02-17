package custmiddleware

import (
	"net/http"

	domainerr "github.com/EfosaE/credora-backend/domain/domianerrors"
	"github.com/EfosaE/credora-backend/internal/response"
	"github.com/hibiken/asynq"
)

type BackpressureMiddleware struct {
	inspector    *asynq.Inspector
	maxQueueSize int
	queueName    string
}

func NewBackpressure(inspector *asynq.Inspector, maxSize int, queueName string) *BackpressureMiddleware {
	return &BackpressureMiddleware{
		inspector:    inspector,
		maxQueueSize: maxSize,
		queueName:    queueName,
	}
}

func (bp *BackpressureMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check pending tasks in queue
		info, err := bp.inspector.GetQueueInfo(bp.queueName)
		if err != nil {
			// Log error but allow request through to avoid blocking on inspection failures
			next.ServeHTTP(w, r)
			return
		}

		pending := info.Pending + info.Active
		if pending >= bp.maxQueueSize {
			response.SendError(w, r,
				response.BuildServerError(
					http.StatusServiceUnavailable,
					domainerr.ErrTooManyPendingJobs,
					"System is currently overloaded.",
				).WithHeaders(map[string]string{
					"Retry-After":       "30",
					"X-RateLimit-Reset": "1234567890",
				}))
			return
		}

		next.ServeHTTP(w, r)
	})
}
