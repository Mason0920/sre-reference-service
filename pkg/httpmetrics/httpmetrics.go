// Package httpmetrics provides RED (rate, errors, duration) metrics for HTTP
// handlers. Metric names match the SLO recording rules in
// observability/prometheus/rules/slo.yml.
package httpmetrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Metrics holds the collectors registered for HTTP traffic.
type Metrics struct {
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
}

// New registers the HTTP collectors on reg.
func New(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "HTTP requests by route pattern, method and status code.",
		}, []string{"handler", "method", "status"}),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency by route pattern and method.",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
		}, []string{"handler", "method"}),
	}
	reg.MustRegister(m.requests, m.duration)
	return m
}

// Middleware records one sample per request. It must wrap the ServeMux
// directly so that r.Pattern (set by the mux during routing) is available
// after the handler returns. Unmatched paths are folded into one label value
// to keep cardinality bounded.
func (m *Metrics) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		handler := r.Pattern
		if handler == "" {
			handler = "unmatched"
		}
		m.requests.WithLabelValues(handler, r.Method, strconv.Itoa(rec.status)).Inc()
		m.duration.WithLabelValues(handler, r.Method).Observe(time.Since(start).Seconds())
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (s *statusRecorder) WriteHeader(code int) {
	if !s.wroteHeader {
		s.status = code
		s.wroteHeader = true
	}
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusRecorder) Write(b []byte) (int, error) {
	s.wroteHeader = true
	return s.ResponseWriter.Write(b)
}

// Unwrap lets http.ResponseController reach the underlying writer.
func (s *statusRecorder) Unwrap() http.ResponseWriter { return s.ResponseWriter }

// RequestsFor returns the request counter for one label set (used in tests).
func (m *Metrics) RequestsFor(handler, method, status string) prometheus.Counter {
	return m.requests.WithLabelValues(handler, method, status)
}
