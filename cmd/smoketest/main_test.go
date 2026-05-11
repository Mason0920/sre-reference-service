package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestRunner(srv *httptest.Server) *smokeRunner {
	return &smokeRunner{
		client:  srv.Client(),
		baseURL: srv.URL,
	}
}

func TestCheckHealth_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
			t.Errorf("encode: %v", err)
		}
	}))
	defer srv.Close()

	if err := newTestRunner(srv).checkHealth(context.Background()); err != nil {
		t.Errorf("checkHealth() = %v, want nil", err)
	}
}

func TestCheckHealth_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	if err := newTestRunner(srv).checkHealth(context.Background()); err == nil {
		t.Error("checkHealth() expected error for 500 status")
	}
}

func TestCheckHealth_BadStatusField(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{"status": "degraded"}); err != nil {
			t.Errorf("encode: %v", err)
		}
	}))
	defer srv.Close()

	if err := newTestRunner(srv).checkHealth(context.Background()); err == nil {
		t.Error("checkHealth() expected error for non-ok status field")
	}
}

func TestCheckHealth_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if _, err := w.Write([]byte("{invalid")); err != nil {
			t.Errorf("write: %v", err)
		}
	}))
	defer srv.Close()

	if err := newTestRunner(srv).checkHealth(context.Background()); err == nil {
		t.Error("checkHealth() expected error for invalid JSON")
	}
}

func TestCheckVersion_NoExpected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{"version": "v1.0.0", "commit": "abc", "buildTime": "2024"}); err != nil {
			t.Errorf("encode: %v", err)
		}
	}))
	defer srv.Close()

	if err := newTestRunner(srv).checkVersion(context.Background(), ""); err != nil {
		t.Errorf("checkVersion() = %v, want nil", err)
	}
}

func TestCheckVersion_MatchExpected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{"version": "v1.2.3", "commit": "abc", "buildTime": "2024"}); err != nil {
			t.Errorf("encode: %v", err)
		}
	}))
	defer srv.Close()

	if err := newTestRunner(srv).checkVersion(context.Background(), "v1.2.3"); err != nil {
		t.Errorf("checkVersion() = %v, want nil", err)
	}
}

func TestCheckVersion_Mismatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{"version": "v1.0.0", "commit": "abc", "buildTime": "2024"}); err != nil {
			t.Errorf("encode: %v", err)
		}
	}))
	defer srv.Close()

	if err := newTestRunner(srv).checkVersion(context.Background(), "v2.0.0"); err == nil {
		t.Error("checkVersion() expected error for version mismatch")
	}
}

func TestCheckVersion_EmptyVersionField(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{"version": "", "commit": "abc", "buildTime": "2024"}); err != nil {
			t.Errorf("encode: %v", err)
		}
	}))
	defer srv.Close()

	if err := newTestRunner(srv).checkVersion(context.Background(), ""); err == nil {
		t.Error("checkVersion() expected error for empty version field")
	}
}

func TestCheckVersion_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	if err := newTestRunner(srv).checkVersion(context.Background(), ""); err == nil {
		t.Error("checkVersion() expected error for 404 status")
	}
}

func TestCheckVersion_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if _, err := w.Write([]byte("{invalid")); err != nil {
			t.Errorf("write: %v", err)
		}
	}))
	defer srv.Close()

	if err := newTestRunner(srv).checkVersion(context.Background(), ""); err == nil {
		t.Error("checkVersion() expected error for invalid JSON")
	}
}

func TestCheckHealth_InvalidURL(t *testing.T) {
	runner := &smokeRunner{client: &http.Client{}, baseURL: "://invalid"}
	if err := runner.checkHealth(context.Background()); err == nil {
		t.Error("checkHealth() expected error for invalid URL")
	}
}

func TestCheckHealth_ServerClosed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	url := srv.URL
	srv.Close()
	runner := &smokeRunner{client: &http.Client{}, baseURL: url}
	if err := runner.checkHealth(context.Background()); err == nil {
		t.Error("checkHealth() expected error for closed server")
	}
}

func TestCheckVersion_InvalidURL(t *testing.T) {
	runner := &smokeRunner{client: &http.Client{}, baseURL: "://invalid"}
	if err := runner.checkVersion(context.Background(), ""); err == nil {
		t.Error("checkVersion() expected error for invalid URL")
	}
}

func TestCheckVersion_ServerClosed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	url := srv.URL
	srv.Close()
	runner := &smokeRunner{client: &http.Client{}, baseURL: url}
	if err := runner.checkVersion(context.Background(), ""); err == nil {
		t.Error("checkVersion() expected error for closed server")
	}
}
