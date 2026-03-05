package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	HttpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "credora",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HttpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "credora",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "HTTP request latency",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)

func Register(reg prometheus.Registerer) {
	reg.MustRegister(HttpRequestsTotal)
	reg.MustRegister(HttpRequestDuration)
}
