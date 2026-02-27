package custmiddleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

func RequestLogger(baseLogger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			// Get request ID (if using chi's middleware)
			requestID := middleware.GetReqID(r.Context())
			fmt.Println("RequestID =", requestID)
			// Create request-scoped logger
			reqLogger := baseLogger.With().
				Str("request_id", requestID).
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("remote_ip", r.RemoteAddr).
				Logger()

				// In your request middleware, store request ID in context directly
			// rather than a full logger
			ctx := context.WithValue(r.Context(), middleware.RequestIDKey, requestID)

			next.ServeHTTP(ww, r.WithContext(ctx))

			duration := time.Since(start)

			reqLogger.Info().
				Int("status", ww.Status()).
				Int("bytes", ww.BytesWritten()).
				Dur("duration", duration).
				Msg("http request completed")
		})
	}
}
