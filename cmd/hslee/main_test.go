package main

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestReadyHandler(t *testing.T) {
	var ready atomic.Bool
	ready.Store(true)
	h := readyHandler(&ready)

	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("ready: got %d, want 200", rec.Code)
	}

	ready.Store(false)
	rec = httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("draining: got %d, want 503", rec.Code)
	}
}

func TestDrainDelay(t *testing.T) {
	cases := map[string]time.Duration{"": 5 * time.Second, "0s": 0, "12s": 12 * time.Second, "bogus": 5 * time.Second, "-1s": 5 * time.Second}
	for in, want := range cases {
		t.Setenv("SHUTDOWN_DRAIN_DELAY", in)
		if got := drainDelay(); got != want {
			t.Errorf("drainDelay(%q) = %v, want %v", in, got, want)
		}
	}
}
