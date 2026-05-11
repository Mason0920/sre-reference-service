package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// errWriter triggers encode errors by failing on Write.
type errWriter struct{ header http.Header }

func (e *errWriter) Header() http.Header {
	if e.header == nil {
		e.header = make(http.Header)
	}
	return e.header
}
func (e *errWriter) Write(_ []byte) (int, error) { return 0, errors.New("write error") }
func (e *errWriter) WriteHeader(_ int)           {}

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	healthHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.StatusCode, http.StatusOK)
	}
	if ct := res.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var body map[string]string
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("body status = %q, want ok", body["status"])
	}
}

func TestVersionHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	w := httptest.NewRecorder()

	versionHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.StatusCode, http.StatusOK)
	}
	if ct := res.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var body map[string]string
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	for _, field := range []string{"version", "commit", "buildTime"} {
		if _, ok := body[field]; !ok {
			t.Errorf("field %q missing from version response", field)
		}
	}
}

func TestHealthHandler_EncodeError(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	healthHandler(&errWriter{}, req)
}

func TestVersionHandler_EncodeError(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	versionHandler(&errWriter{}, req)
}

func TestNewServeMux(t *testing.T) {
	srv := httptest.NewServer(newServeMux())
	defer srv.Close()

	for _, path := range []string{"/health", "/version"} {
		resp, err := http.Get(srv.URL + path)
		if err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("%s status = %d, want 200", path, resp.StatusCode)
		}
	}
}

func TestVersionHandler_DefaultValues(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	w := httptest.NewRecorder()

	versionHandler(w, req)

	var body map[string]string
	if err := json.NewDecoder(w.Result().Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["version"] != "dev" {
		t.Errorf("version = %q, want dev", body["version"])
	}
	if body["commit"] != "none" {
		t.Errorf("commit = %q, want none", body["commit"])
	}
	if body["buildTime"] != "unknown" {
		t.Errorf("buildTime = %q, want unknown", body["buildTime"])
	}
}
