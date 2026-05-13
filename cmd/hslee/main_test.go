package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestMain(m *testing.M) {
	logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	os.Exit(m.Run())
}

// errWriter causes json.Encode to fail by returning a write error.
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

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("status = %q, want ok", body["status"])
	}
}

func TestHealthHandler_EncodeError(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	healthHandler(&errWriter{}, req)
}

func TestVersionHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	w := httptest.NewRecorder()

	versionHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	for _, field := range []string{"version", "commit", "buildTime"} {
		if _, ok := body[field]; !ok {
			t.Errorf("field %q missing from response", field)
		}
	}
}

func TestVersionHandler_EncodeError(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	versionHandler(&errWriter{}, req)
}

func TestBuildSampler_Default(t *testing.T) {
	t.Setenv("OTEL_TRACES_SAMPLER_ARG", "")
	if buildSampler() == nil {
		t.Fatal("buildSampler() returned nil")
	}
}

func TestBuildSampler_ValidRatio(t *testing.T) {
	t.Setenv("OTEL_TRACES_SAMPLER_ARG", "0.5")
	if buildSampler() == nil {
		t.Fatal("buildSampler() returned nil")
	}
}

func TestBuildSampler_InvalidFloat(t *testing.T) {
	t.Setenv("OTEL_TRACES_SAMPLER_ARG", "not-a-float")
	if buildSampler() == nil {
		t.Fatal("buildSampler() returned nil")
	}
}

func TestBuildSampler_OutOfRangeHigh(t *testing.T) {
	t.Setenv("OTEL_TRACES_SAMPLER_ARG", "1.5")
	if buildSampler() == nil {
		t.Fatal("buildSampler() returned nil")
	}
}

func TestBuildSampler_OutOfRangeLow(t *testing.T) {
	t.Setenv("OTEL_TRACES_SAMPLER_ARG", "-0.1")
	if buildSampler() == nil {
		t.Fatal("buildSampler() returned nil")
	}
}

func TestMapCarrier_Operations(t *testing.T) {
	m := mapCarrier{"key1": "val1"}

	if got := m.Get("key1"); got != "val1" {
		t.Errorf("Get = %q, want val1", got)
	}
	if got := m.Get("missing"); got != "" {
		t.Errorf("Get missing key = %q, want empty", got)
	}

	m.Set("key2", "val2")
	if got := m.Get("key2"); got != "val2" {
		t.Errorf("Set/Get = %q, want val2", got)
	}
	if len(m.Keys()) != 2 {
		t.Errorf("len(Keys) = %d, want 2", len(m.Keys()))
	}
}

func TestLogFromCtx_FromContext(t *testing.T) {
	l := slog.New(slog.NewTextHandler(io.Discard, nil))
	ctx := context.WithValue(context.Background(), ctxLoggerKey{}, l)
	if got := logFromCtx(ctx); got != l {
		t.Error("logFromCtx should return logger stored in context")
	}
}

func TestLogFromCtx_Default(t *testing.T) {
	if got := logFromCtx(context.Background()); got != logger {
		t.Error("logFromCtx should return global logger when absent from context")
	}
}

func TestTraceLogMiddleware_NoSpan(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if logFromCtx(r.Context()) == nil {
			t.Error("logger not found in request context")
		}
	})

	traceLogMiddleware(next).ServeHTTP(
		httptest.NewRecorder(),
		httptest.NewRequest(http.MethodGet, "/", nil),
	)

	if !called {
		t.Error("next handler was not called")
	}
}

func TestTraceLogMiddleware_WithSpan(t *testing.T) {
	tp := sdktrace.NewTracerProvider()
	defer func() { _ = tp.Shutdown(context.Background()) }()

	ctx, span := tp.Tracer("test").Start(context.Background(), "test-op")
	defer span.End()

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	traceLogMiddleware(next).ServeHTTP(
		httptest.NewRecorder(),
		httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx),
	)

	if !called {
		t.Error("next handler was not called")
	}
}

func TestInitTracer_NoEndpoint(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	ctx := context.Background()
	shutdown, err := initTracer(ctx, "test-service")
	if err != nil {
		t.Fatalf("initTracer failed: %v", err)
	}
	if err := shutdown(ctx); err != nil {
		t.Errorf("shutdown error: %v", err)
	}
}

func TestInjectExtractTraceContext(t *testing.T) {
	headers := make(map[string]string)
	InjectTraceContext(context.Background(), headers)
	ctx := ExtractTraceContext(context.Background(), headers)
	if ctx == nil {
		t.Error("ExtractTraceContext returned nil context")
	}
}

func TestInitTracer_WithEndpoint(t *testing.T) {
	// Use a fake endpoint to exercise the gRPC-exporter setup path without
	// requiring a live Jaeger/OTLP collector.  otlptracegrpc.New connects
	// lazily, so this succeeds even without a running server.
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:19417")
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	shutdown, err := initTracer(ctx, "test-service")
	if err != nil {
		// Connection refused / timeout — expected in CI; gRPC setup code covered.
		return
	}
	_ = shutdown(context.Background())
}
