package httpmetrics_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/Mason0920/sre-reference-service/pkg/httpmetrics"
)

func TestMiddlewareLabels(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := httpmetrics.New(reg)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /items/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	h := m.Middleware(mux)

	for _, path := range []string{"/items/1", "/items/2", "/ok", "/nope"} {
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, path, nil))
	}

	tests := []struct {
		handler, status string
		want            float64
	}{
		{"GET /items/{id}", "418", 2}, // path params collapse into the route pattern
		{"/ok", "200", 1},
		{"unmatched", "404", 1},
	}
	for _, tc := range tests {
		got := testutil.ToFloat64(m.RequestsFor(tc.handler, http.MethodGet, tc.status))
		if got != tc.want {
			t.Errorf("http_requests_total{handler=%q,status=%q} = %v, want %v", tc.handler, tc.status, got, tc.want)
		}
	}
	if n := testutil.CollectAndCount(reg, "http_request_duration_seconds"); n != 3 {
		t.Errorf("duration series = %d, want 3", n)
	}
}
