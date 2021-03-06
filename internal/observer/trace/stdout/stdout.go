package stdout

import (
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout"
	"go.opentelemetry.io/otel/sdk/trace"
)

func Initialize(config *trace.Config) error {
	exporter, err := stdout.NewExporter()
	if err != nil {
		return fmt.Errorf("failed to create stdout trace exporter: %w", err)
	}

	provider := trace.NewTracerProvider(
		trace.WithConfig(*config),
		trace.WithSyncer(exporter),
	)

	otel.SetTracerProvider(provider)

	return nil
}
