package cloudtrace

import (
	"fmt"

	exporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"github.com/go-logr/logr"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/sdk/trace"
)

func Initialize(c *trace.Config, projectID string, logger logr.Logger) error {
	exp, err := exporter.NewExporter(
		exporter.WithProjectID(projectID),
		exporter.WithOnError(func(err error) {
			logger.Error(err, "error when tracing with Cloud Trace")
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to create cloud trace exporter: %w", err)
	}

	tp, err := trace.NewProvider(trace.WithSyncer(exp))
	if err != nil {
		return fmt.Errorf("failed to create cloud trace provider: %w", err)
	}

	global.SetTraceProvider(tp)

	return nil
}
