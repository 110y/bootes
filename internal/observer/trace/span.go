package trace

import (
	"go.opentelemetry.io/otel/api/trace"
)

var _ Span = (*openTelemetrySpan)(nil)

type Span interface {
	End()
}

type openTelemetrySpan struct {
	span trace.Span
}

func (s *openTelemetrySpan) End() {
	s.span.End()
}
