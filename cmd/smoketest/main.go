// Package main is a standalone smoke test runner for post-deployment validation.
// It checks key endpoints and verifies the deployed version, exiting non-zero on failure.
//
// Usage:
//
//	smoketest --url http://my-app.example.com [--version v1.2.3] [--timeout 30s]
//	SMOKE_TEST_URL=http://... EXPECTED_VERSION=v1.2.3 smoketest
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	var (
		baseURL         string
		expectedVersion string
		timeout         time.Duration
	)

	flag.StringVar(&baseURL, "url", os.Getenv("SMOKE_TEST_URL"), "Base URL of the deployed application (required)")
	flag.StringVar(&expectedVersion, "version", os.Getenv("EXPECTED_VERSION"), "Expected deployed version string (optional; skips version comparison if empty)")
	flag.DurationVar(&timeout, "timeout", 30*time.Second, "Overall timeout for all smoke test checks")
	flag.Parse()

	if baseURL == "" {
		fmt.Fprintln(os.Stderr, "error: --url flag or SMOKE_TEST_URL env var is required")
		os.Exit(2)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	runner := &smokeRunner{
		client:  &http.Client{},
		baseURL: baseURL,
	}

	failed := false

	// Check 1: /health — verifies the app is up and reports status "ok".
	if err := runner.checkHealth(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL  /health  — %v\n", err)
		failed = true
	} else {
		fmt.Println("PASS  /health  — status ok")
	}

	// Check 2: /version — verifies the correct build is deployed.
	if err := runner.checkVersion(ctx, expectedVersion); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL  /version — %v\n", err)
		failed = true
	} else {
		fmt.Println("PASS  /version — version verified")
	}

	cancel()
	fmt.Println()
	if failed {
		fmt.Fprintln(os.Stderr, "Smoke tests FAILED")
		os.Exit(1)
	}
	fmt.Println("Smoke tests PASSED")
}

type smokeRunner struct {
	client  *http.Client
	baseURL string
}

func (s *smokeRunner) checkHealth(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.baseURL+"/health", nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected HTTP status %d (want 200)", resp.StatusCode)
	}

	var body map[string]string
	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		return fmt.Errorf("decoding response body: %w", err)
	}

	if body["status"] != "ok" {
		return fmt.Errorf("status field is %q (want %q)", body["status"], "ok")
	}

	return nil
}

func (s *smokeRunner) checkVersion(ctx context.Context, expected string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.baseURL+"/version", nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected HTTP status %d (want 200)", resp.StatusCode)
	}

	var body map[string]string
	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		return fmt.Errorf("decoding response body: %w", err)
	}

	got := body["version"]
	if got == "" {
		return fmt.Errorf("version field is empty in response")
	}

	if expected != "" && got != expected {
		return fmt.Errorf("version mismatch: deployed %q, expected %q", got, expected)
	}

	return nil
}
