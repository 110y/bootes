package jaeger

import (
	"fmt"

	"github.com/go-logr/logr"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/sdk/trace"
)

const serviceName = "bootes"

func Initialize(c *trace.Config, endpoint string, logger logr.Logger) (func(), error) {
	tp, flush, err := jaeger.NewExportPipeline(
		jaeger.WithCollectorEndpoint(endpoint),
		jaeger.WithProcess(jaeger.Process{
			ServiceName: serviceName,
			Tags: []label.KeyValue{
				label.String("exporter", "jaeger"),
			},
		}),
		jaeger.WithSDK(c),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create jaeger trace exporter: %w", err)
	}

	global.SetTracerProvider(tp)

	return func() {
		logger.Info("flushing jaeger traces")
		flush()
	}, nil
}
