package tracing

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	ServiceName    = "user-api"
	ServiceVersion = "1.0.0"
)

// TracingConfig holds tracing configuration
type TracingConfig struct {
	Enabled      bool
	ExporterType string // "console", "otlp"
	OTLPEndpoint string
	SamplingRate float64
	Environment  string
}

// InitTracing initializes OpenTelemetry tracing
func InitTracing(config TracingConfig) (func(context.Context) error, error) {
	if !config.Enabled {
		log.Println("Tracing is disabled")
		return func(context.Context) error { return nil }, nil
	}

	// Create resource
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(ServiceName),
			semconv.ServiceVersion(ServiceVersion),
			semconv.DeploymentEnvironment(config.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create exporter based on configuration
	var exporter sdktrace.SpanExporter
	switch config.ExporterType {
	case "console":
		exporter, err = stdouttrace.New(
			stdouttrace.WithPrettyPrint(),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create console exporter: %w", err)
		}
		log.Println("Using console trace exporter")

	case "otlp":
		opts := []otlptracehttp.Option{
			otlptracehttp.WithInsecure(),
		}
		if config.OTLPEndpoint != "" {
			opts = append(opts, otlptracehttp.WithEndpoint(config.OTLPEndpoint))
		}

		exporter, err = otlptracehttp.New(context.Background(), opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
		}
		log.Printf("Using OTLP trace exporter with endpoint: %s", config.OTLPEndpoint)

	default:
		return nil, fmt.Errorf("unsupported exporter type: %s", config.ExporterType)
	}

	// Create sampler
	var sampler sdktrace.Sampler
	if config.SamplingRate >= 1.0 {
		sampler = sdktrace.AlwaysSample()
	} else if config.SamplingRate <= 0.0 {
		sampler = sdktrace.NeverSample()
	} else {
		sampler = sdktrace.TraceIDRatioBased(config.SamplingRate)
	}

	// Create trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	// Set global trace provider
	otel.SetTracerProvider(tp)

	// Set global propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	log.Printf("Tracing initialized successfully with sampling rate: %.2f", config.SamplingRate)

	// Return shutdown function
	return tp.Shutdown, nil
}

// GetTracer returns a tracer for the given name
func GetTracer(name string) trace.Tracer {
	return otel.Tracer(name)
}

// StartSpan starts a new span with the given name and options
func StartSpan(ctx context.Context, tracer trace.Tracer, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return tracer.Start(ctx, spanName, opts...)
}

// AddSpanAttributes adds attributes to a span
func AddSpanAttributes(span trace.Span, attrs ...attribute.KeyValue) {
	span.SetAttributes(attrs...)
}

// AddSpanEvent adds an event to a span
func AddSpanEvent(span trace.Span, name string, attrs ...attribute.KeyValue) {
	span.AddEvent(name, trace.WithAttributes(attrs...))
}

// RecordError records an error on a span
func RecordError(span trace.Span, err error) {
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

// GetTraceID extracts trace ID from context
func GetTraceID(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasTraceID() {
		return spanCtx.TraceID().String()
	}
	return ""
}

// GetSpanID extracts span ID from context
func GetSpanID(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasSpanID() {
		return spanCtx.SpanID().String()
	}
	return ""
}

// LoadTracingConfigFromEnv loads tracing configuration from environment variables
func LoadTracingConfigFromEnv(environment string) TracingConfig {
	config := TracingConfig{
		Environment: environment,
	}

	// Parse enabled flag
	if enabled := os.Getenv("TRACING_ENABLED"); enabled != "" {
		config.Enabled, _ = strconv.ParseBool(enabled)
	} else {
		// Default to enabled in development, disabled in production
		config.Enabled = environment == "development"
	}

	// Parse exporter type
	config.ExporterType = os.Getenv("TRACING_EXPORTER")
	if config.ExporterType == "" {
		if environment == "development" {
			config.ExporterType = "console"
		} else {
			config.ExporterType = "otlp"
		}
	}

	// Parse OTLP endpoint
	config.OTLPEndpoint = os.Getenv("TRACING_OTLP_ENDPOINT")
	if config.OTLPEndpoint == "" {
		config.OTLPEndpoint = "http://localhost:4318/v1/traces"
	}

	// Parse sampling rate
	if samplingStr := os.Getenv("TRACING_SAMPLING_RATE"); samplingStr != "" {
		if rate, err := strconv.ParseFloat(samplingStr, 64); err == nil {
			config.SamplingRate = rate
		} else {
			config.SamplingRate = 1.0 // Default to 100% sampling
		}
	} else {
		if environment == "development" {
			config.SamplingRate = 1.0 // 100% sampling in development
		} else {
			config.SamplingRate = 0.1 // 10% sampling in production
		}
	}

	return config
}

// Common span attribute keys
var (
	AttrHTTPMethod     = attribute.Key("http.method")
	AttrHTTPURL        = attribute.Key("http.url")
	AttrHTTPStatusCode = attribute.Key("http.status_code")
	AttrHTTPUserAgent  = attribute.Key("http.user_agent")
	AttrHTTPClientIP   = attribute.Key("http.client_ip")
	AttrUserID         = attribute.Key("user.id")
	AttrUserEmail      = attribute.Key("user.email")
	AttrRequestSize    = attribute.Key("http.request.size")
	AttrResponseSize   = attribute.Key("http.response.size")
	AttrErrorType      = attribute.Key("error.type")
	AttrErrorMessage   = attribute.Key("error.message")
	AttrDBOperation    = attribute.Key("db.operation")
	AttrDBTable        = attribute.Key("db.table")
)
