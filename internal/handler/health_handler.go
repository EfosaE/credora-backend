package handler

import (
	"net/http"
	"time"

	"github.com/EfosaE/credora-backend/internal/response"
	"github.com/EfosaE/credora-backend/internal/utils"
	"github.com/hibiken/asynq"
)

type HealthHandler struct {
	inspector     *asynq.Inspector
	queueName     string
	queueCapacity int
	startTime     time.Time
}

func NewHealthHandler(inspector *asynq.Inspector, queueName string, queueCapacity int) *HealthHandler {
	return &HealthHandler{
		inspector:     inspector,
		queueName:     queueName,
		queueCapacity: queueCapacity,
		startTime:     time.Now(),
	}
}

type healthData struct {
	Status        string `json:"status"`
	PendingJobs   int    `json:"pending_jobs"`
	ActiveJobs    int    `json:"active_jobs"`
	QueueCapacity int    `json:"queue_capacity"`
	UptimeSeconds int64  `json:"uptime_seconds"`
}

// Liveness — app is running
func (h *HealthHandler) Liveness(w http.ResponseWriter, r *http.Request) {
	response.SendSuccess(w, r, response.OK(
		response.Obj("status", "ok"),
		nil,
		"Service is alive",
	))
}

// Readiness — app + queue state for load testing visibility
func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	info, err := h.inspector.GetQueueInfo(h.queueName)
	if err != nil {
		data, err := utils.StructToMap(healthData{
			Status:        "degraded",
			QueueCapacity: h.queueCapacity,
			UptimeSeconds: int64(time.Since(h.startTime).Seconds()),
		})
		if err != nil {
			response.SendError(w, r, response.InternalServerError(err, err.Error()))
			return
		}
		response.SendSuccess(w, r, response.OK(data, nil, "Queue unavailable"))
		return
	}

	status := "ok"
	if (info.Pending + info.Active) >= h.queueCapacity {
		status = "degraded"
	}

	data, err := utils.StructToMap(healthData{
		Status:        status,
		PendingJobs:   info.Pending,
		ActiveJobs:    info.Active,
		QueueCapacity: h.queueCapacity,
		UptimeSeconds: int64(time.Since(h.startTime).Seconds()),
	})
	if err != nil {
		response.SendError(w, r, response.InternalServerError(err, err.Error()))
		return
	}

	response.SendSuccess(w, r, response.OK(data, nil, "Service is ready"))
}
