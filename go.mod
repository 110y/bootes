module github.com/110y/bootes

go 1.15

require (
	github.com/110y/bootes-api v0.0.0-20200715085629-385882e22027
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace v0.12.1-0.20201022231920-402639db27fd
	github.com/envoyproxy/go-control-plane v0.9.7
	github.com/go-logr/logr v0.2.1
	github.com/go-logr/zapr v0.2.0
	github.com/golang/protobuf v1.4.2
	github.com/google/go-cmp v0.5.2
	github.com/google/uuid v1.1.2
	github.com/kelseyhightower/envconfig v1.4.0
	go.opentelemetry.io/otel v0.13.0
	go.opentelemetry.io/otel/exporters/stdout v0.13.0
	go.opentelemetry.io/otel/exporters/trace/jaeger v0.13.0
	go.opentelemetry.io/otel/sdk v0.13.0
	go.uber.org/zap v1.16.0
	google.golang.org/grpc v1.33.1
	google.golang.org/protobuf v1.25.0
	k8s.io/api v0.19.3
	k8s.io/apimachinery v0.19.3
	k8s.io/client-go v0.19.3
	k8s.io/klog/v2 v2.3.0
	sigs.k8s.io/controller-runtime v0.6.3
)
