package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/redis/go-redis/v9"
)

type HealthHandler struct {
	redis *redis.Client
}

func NewHealthHandler(redis *redis.Client) *HealthHandler {
	return &HealthHandler{redis: redis}
}

// Liveness — app is running
func (h *HealthHandler) Liveness(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

// Readiness — app + Redis are ready
func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	resp := make(map[string]string)

	// Check Redis
	if _, err := h.redis.Ping(ctx).Result(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		resp["redis"] = "down"
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp["redis"] = "ok"
	json.NewEncoder(w).Encode(resp)
}
