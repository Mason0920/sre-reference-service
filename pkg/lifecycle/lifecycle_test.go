package lifecycle_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Mason0920/sre-reference-service/pkg/lifecycle"
)

func TestReadinessHandler(t *testing.T) {
	r := lifecycle.NewReadiness()
	h := r.Handler()

	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("ready: got %d, want 200", rec.Code)
	}

	r.StartDraining()
	rec = httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("draining: got %d, want 503", rec.Code)
	}
}

func TestDrainDelay(t *testing.T) {
	cases := map[string]time.Duration{
		"":      lifecycle.DefaultDrainDelay,
		"0s":    0,
		"12s":   12 * time.Second,
		"bogus": lifecycle.DefaultDrainDelay,
		"-1s":   lifecycle.DefaultDrainDelay,
	}
	for in, want := range cases {
		t.Setenv("SHUTDOWN_DRAIN_DELAY", in)
		if got := lifecycle.DrainDelay(); got != want {
			t.Errorf("DrainDelay(%q) = %v, want %v", in, got, want)
		}
	}
}
