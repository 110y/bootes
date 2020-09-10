module github.com/110y/bootes

go 1.15

require (
	github.com/110y/bootes-api v0.0.0-20200715085629-385882e22027
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace v0.1.1-0.20200514210843-966afdc5d38c
	github.com/envoyproxy/go-control-plane v0.9.6
	github.com/go-logr/logr v0.2.1
	github.com/go-logr/zapr v0.1.1
	github.com/golang/protobuf v1.4.2
	github.com/google/go-cmp v0.4.0
	github.com/google/uuid v1.1.1
	github.com/kelseyhightower/envconfig v1.4.0
	go.opentelemetry.io/otel v0.5.0
	go.opentelemetry.io/otel/exporters/trace/jaeger v0.5.0
	go.uber.org/zap v1.14.0
	google.golang.org/grpc v1.31.0
	google.golang.org/protobuf v1.24.0
	k8s.io/api v0.19.1
	k8s.io/apimachinery v0.19.1
	k8s.io/client-go v0.19.1
	sigs.k8s.io/controller-runtime v0.6.2
)
