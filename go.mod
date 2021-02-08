module github.com/110y/bootes

go 1.15

require (
	github.com/110y/bootes-api v0.0.0-20210204020725-8ba7924f8cd2
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace v0.13.1-0.20201124165336-540edb3af173
	github.com/envoyproxy/go-control-plane v0.9.8
	github.com/go-logr/logr v0.4.0
	github.com/go-logr/zapr v0.4.0
	github.com/golang/protobuf v1.4.3
	github.com/google/go-cmp v0.5.4
	github.com/google/uuid v1.2.0
	github.com/kelseyhightower/envconfig v1.4.0
	go.opentelemetry.io/otel v0.14.0
	go.opentelemetry.io/otel/exporters/stdout v0.14.0
	go.opentelemetry.io/otel/exporters/trace/jaeger v0.14.0
	go.opentelemetry.io/otel/sdk v0.14.0
	go.uber.org/zap v1.16.0
	google.golang.org/grpc v1.33.1
	google.golang.org/protobuf v1.25.0
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	k8s.io/klog/v2 v2.5.0
	sigs.k8s.io/controller-runtime v0.8.1
)
