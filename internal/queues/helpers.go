package queues

import (
	"context"
	"log"
)

func (h *Handlers) logError(ctx context.Context, err error, msg string, meta map[string]any) {
	// Add error automatically to meta
	if meta == nil {
		meta = map[string]any{}
	}
	meta["error"] = err.Error()

	// 1️⃣ Terminal logging (Asynq worker console)
	log.Printf("ERROR: %s | %v | meta=%v\n", msg, err, meta)

	// 2️⃣ App logger (writes to file)
	h.AppLogger.Error(msg, meta)
}

func (h *Handlers) logInfo(msg string, meta map[string]any) {
	log.Printf("INFO: %s | meta=%v\n", msg, meta)
	h.AppLogger.Info(msg, meta)
}
