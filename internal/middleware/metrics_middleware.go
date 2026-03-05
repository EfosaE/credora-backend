package custmiddleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/EfosaE/credora-backend/internal/metrics"
	"github.com/go-chi/chi/v5/middleware"
)

func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(rw, r)

		duration := time.Since(start).Seconds()

		metrics.HttpRequestsTotal.
			WithLabelValues(r.Method, r.URL.Path, strconv.Itoa(rw.Status())).
			Inc()

		metrics.HttpRequestDuration.
			WithLabelValues(r.Method, r.URL.Path).
			Observe(duration)
	})
}
