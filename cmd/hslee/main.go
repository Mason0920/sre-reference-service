package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/Mason0920/sre-reference-service/pkg/httpmetrics"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

// Version metadata injected at build time via -ldflags.
var (
	version   = "dev"
	commit    = "none"
	buildTime = "unknown"
)

var logger *slog.Logger

type ctxLoggerKey struct{}

func initTracer(ctx context.Context, serviceName string) (func(context.Context) error, error) {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "localhost:4317"
	}

	// Use a bounded timeout so startup doesn't hang when the OTLP endpoint is unreachable.
	dialCtx, dialCancel := context.WithTimeout(ctx, 3*time.Second)
	defer dialCancel()
	exporter, err := otlptracegrpc.New(dialCtx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("creating OTLP exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithProcess(),
		resource.WithHost(),
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(version),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("creating resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(buildSampler()),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp.Shutdown, nil
}

// buildSampler returns a sampler configured via OTEL_TRACES_SAMPLER_ARG (0.0–1.0 ratio).
// Defaults to AlwaysSample when the env var is absent or invalid.
func buildSampler() sdktrace.Sampler {
	arg := os.Getenv("OTEL_TRACES_SAMPLER_ARG")
	if arg == "" {
		return sdktrace.AlwaysSample()
	}
	r, err := strconv.ParseFloat(arg, 64)
	if err != nil || r < 0 || r > 1 {
		return sdktrace.AlwaysSample()
	}
	return sdktrace.ParentBased(sdktrace.TraceIDRatioBased(r))
}

// mapCarrier implements propagation.TextMapCarrier for generic map-based message
// headers (Kafka, NATS, Redis Streams, etc.).
type mapCarrier map[string]string

func (m mapCarrier) Get(key string) string        { return m[key] }
func (m mapCarrier) Set(key string, value string) { m[key] = value }
func (m mapCarrier) Keys() []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// InjectTraceContext injects the current span context into message headers.
func InjectTraceContext(ctx context.Context, headers map[string]string) {
	otel.GetTextMapPropagator().Inject(ctx, mapCarrier(headers))
}

// ExtractTraceContext extracts a remote span context from message headers.
func ExtractTraceContext(ctx context.Context, headers map[string]string) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, mapCarrier(headers))
}

// traceLogMiddleware injects trace_id and span_id into the request-scoped logger.
func traceLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		span := trace.SpanFromContext(r.Context())
		sc := span.SpanContext()
		reqLogger := logger
		if sc.IsValid() {
			reqLogger = logger.With(
				"trace_id", sc.TraceID().String(),
				"span_id", sc.SpanID().String(),
			)
		}
		ctx := context.WithValue(r.Context(), ctxLoggerKey{}, reqLogger)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func logFromCtx(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(ctxLoggerKey{}).(*slog.Logger); ok {
		return l
	}
	return logger
}

func main() {
	var level slog.Level
	switch os.Getenv("LOG_LEVEL") {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	logger = slog.New(handler).With(
		"service", "hslee",
		"service_name", "hslee",
		"version", version,
		"commit", commit,
	)
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var shutdown func(context.Context) error
	if tracerShutdown, err := initTracer(ctx, "hslee"); err != nil {
		logger.Warn("tracing disabled - running without distributed tracing", "error", err)
	} else {
		shutdown = tracerShutdown
		logger.Info("OpenTelemetry tracer initialized",
			"endpoint", os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	}

	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	reg := prometheus.NewRegistry()
	reg.MustRegister(collectors.NewGoCollector(), collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	metrics := httpmetrics.New(reg)

	var ready atomic.Bool
	ready.Store(true)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/ready", readyHandler(&ready))
	mux.HandleFunc("/version", versionHandler)
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))

	// Wrap with OTel HTTP instrumentation; degrades to no-op if tracing not initialized.
	// Health checks excluded to reduce trace noise. The metrics middleware wraps
	// the mux directly so it can read the matched route pattern.
	rootHandler := otelhttp.NewHandler(
		traceLogMiddleware(metrics.Middleware(mux)),
		"hslee-server",
		otelhttp.WithMessageEvents(otelhttp.ReadEvents, otelhttp.WriteEvents),
		otelhttp.WithFilter(func(r *http.Request) bool {
			return r.URL.Path != "/health" && r.URL.Path != "/ready" && r.URL.Path != "/metrics"
		}),
	)

	logger.Info("server starting", "addr", addr, "built", buildTime)

	srv := &http.Server{
		Addr:    addr,
		Handler: rootHandler,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()

	// Fail readiness first and keep serving for a short drain window, so the
	// load balancer stops sending new requests before connections close.
	ready.Store(false)
	drain := drainDelay()
	logger.Info("shutting down server", "drain_delay", drain.String())
	time.Sleep(drain)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown error", "error", err)
	}

	// Flush in-flight spans only after the HTTP server has fully drained.
	if shutdown != nil {
		tracerCtx, tracerCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer tracerCancel()
		if err := shutdown(tracerCtx); err != nil {
			logger.Error("tracer shutdown error", "error", err)
		}
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	l := logFromCtx(r.Context())
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok", "version": version}); err != nil {
		l.Error("health handler encode error", "error", err)
		return
	}
	l.Debug("health check", "remote_addr", r.RemoteAddr)
}

func versionHandler(w http.ResponseWriter, r *http.Request) {
	l := logFromCtx(r.Context())
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"version":   version,
		"commit":    commit,
		"buildTime": buildTime,
	}); err != nil {
		l.Error("version handler encode error", "error", err)
		return
	}
	l.Debug("version requested", "remote_addr", r.RemoteAddr)
}

// readyHandler reports 503 once shutdown has started so that Kubernetes and
// load balancers take the pod out of rotation before the server stops.
func readyHandler(ready *atomic.Bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if !ready.Load() {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"draining"}`))
			return
		}
		_, _ = w.Write([]byte(`{"status":"ready"}`))
	}
}

// drainDelay reads SHUTDOWN_DRAIN_DELAY (Go duration, default 5s).
func drainDelay() time.Duration {
	if v := os.Getenv("SHUTDOWN_DRAIN_DELAY"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d >= 0 {
			return d
		}
	}
	return 5 * time.Second
}
