// Package lifecycle holds the readiness state used for graceful shutdown.
//
// On SIGTERM the server first reports "not ready" and keeps serving for a
// short drain window. Kubernetes endpoints and load balancer target groups
// take the pod out of rotation during that window, so in-flight and late
// requests are not cut off when the listener closes.
package lifecycle

import (
	"net/http"
	"os"
	"sync/atomic"
	"time"
)

// DefaultDrainDelay is used when SHUTDOWN_DRAIN_DELAY is unset or invalid.
const DefaultDrainDelay = 5 * time.Second

// Readiness reports whether the process should receive new traffic.
type Readiness struct {
	ready atomic.Bool
}

// NewReadiness returns a Readiness that starts in the ready state.
func NewReadiness() *Readiness {
	r := &Readiness{}
	r.ready.Store(true)
	return r
}

// StartDraining marks the process as not ready.
func (r *Readiness) StartDraining() { r.ready.Store(false) }

// Ready reports the current state.
func (r *Readiness) Ready() bool { return r.ready.Load() }

// Handler serves 200 while ready and 503 while draining.
func (r *Readiness) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if !r.Ready() {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"draining"}`))
			return
		}
		_, _ = w.Write([]byte(`{"status":"ready"}`))
	}
}

// DrainDelay reads SHUTDOWN_DRAIN_DELAY as a Go duration (for example "10s").
func DrainDelay() time.Duration {
	if v := os.Getenv("SHUTDOWN_DRAIN_DELAY"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d >= 0 {
			return d
		}
	}
	return DefaultDrainDelay
}
