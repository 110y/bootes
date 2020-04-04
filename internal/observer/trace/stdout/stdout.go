package stdout

import (
	"fmt"

	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/exporters/trace/stdout"
	"go.opentelemetry.io/otel/sdk/trace"
)

func Initialize(config *trace.Config) error {
	options := stdout.Options{
		PrettyPrint: false,
	}
	exporter, err := stdout.NewExporter(options)
	if err != nil {
		return fmt.Errorf("failed to create stdout trace exporter: %w", err)
	}

	provider, err := trace.NewProvider(
		trace.WithConfig(*config),
		trace.WithSyncer(exporter),
	)
	if err != nil {
		return fmt.Errorf("failed to create stdout trace provider: %w", err)
	}

	global.SetTraceProvider(provider)

	return nil
}
