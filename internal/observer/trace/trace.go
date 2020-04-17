package trace

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"go.opentelemetry.io/otel/api/global"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/110y/bootes/internal/observer/trace/cloudtrace"
	"github.com/110y/bootes/internal/observer/trace/jaeger"
	"github.com/110y/bootes/internal/observer/trace/stdout"
)

const tracerName = "bootes.io"

type Config struct {
	UseStdout              bool
	UseJaeger              bool
	JaegerEndpoint         string
	UseGCPCloudTrace       bool
	GCPCloudTraceProjectID string
	Logger                 logr.Logger
}

func Initialize(config *Config) (func(), error) {
	c := &sdktrace.Config{
		DefaultSampler: sdktrace.AlwaysSample(),
	}

	if config.UseStdout {
		if err := stdout.Initialize(c); err != nil {
			return nil, fmt.Errorf("failed to initialize stdout tracer: %w", err)
		}
	}

	var flushers []func()

	if config.UseJaeger {
		flush, err := jaeger.Initialize(c, config.JaegerEndpoint, config.Logger.WithName("jaeger"))
		if err != nil {
			return nil, fmt.Errorf("failed to initialize jaeger tracer: %w", err)
		}

		flushers = append(flushers, flush)
	}

	if config.UseGCPCloudTrace {
		if err := cloudtrace.Initialize(c, config.GCPCloudTraceProjectID, config.Logger.WithName("cloud-trace")); err != nil {
			return nil, fmt.Errorf("failed to initialize cloud trace tracer: %w", err)
		}
	}

	return func() {
		for _, flush := range flushers {
			flush()
		}
	}, nil
}

func NewSpan(ctx context.Context, name string) (context.Context, Span) {
	ctx, span := global.Tracer(tracerName).Start(ctx, name)
	return ctx, &openTelemetrySpan{span: span}
}
