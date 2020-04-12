package jaeger

import (
	"fmt"

	"github.com/go-logr/logr"
	"go.opentelemetry.io/otel/api/core"
	"go.opentelemetry.io/otel/api/key"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/sdk/trace"
)

const serviceName = "bootes"

func Initialize(c *trace.Config, endpoint string, logger logr.Logger) (func(), error) {
	_, flush, err := jaeger.NewExportPipeline(
		jaeger.WithCollectorEndpoint(endpoint),
		jaeger.WithProcess(jaeger.Process{
			ServiceName: serviceName,
			Tags: []core.KeyValue{
				key.String("exporter", "jaeger"),
			},
		}),
		jaeger.RegisterAsGlobal(),
		jaeger.WithSDK(c),
		jaeger.WithOnError(func(err error) {
			logger.Error(err, "error when uploading spans to Jaeger")
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create jaeger trace exporter: %w", err)
	}

	return func() {
		logger.Info("flushing jaeger traces")
		flush()
	}, nil
}
