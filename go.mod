module github.com/110y/bootes

go 1.16

require (
	github.com/110y/bootes-api v1.0.0
	github.com/110y/run v1.0.0
	github.com/110y/servergroup v0.0.0-20210626143417-cad3a5ad181e
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace v0.16.0
	github.com/envoyproxy/go-control-plane v0.9.8
	github.com/go-logr/logr v0.4.0
	github.com/go-logr/zapr v0.4.0
	github.com/golang/protobuf v1.5.2
	github.com/google/go-cmp v0.5.6
	github.com/google/uuid v1.2.0
	github.com/kelseyhightower/envconfig v1.4.0
	go.opentelemetry.io/otel v0.16.0
	go.opentelemetry.io/otel/exporters/stdout v0.16.0
	go.opentelemetry.io/otel/exporters/trace/jaeger v0.16.0
	go.opentelemetry.io/otel/sdk v0.16.0
	go.uber.org/zap v1.17.0
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c
	google.golang.org/grpc v1.33.2
	google.golang.org/protobuf v1.26.0
	k8s.io/api v0.21.1
	k8s.io/apimachinery v0.21.1
	k8s.io/client-go v0.21.1
	k8s.io/klog/v2 v2.8.0
	sigs.k8s.io/controller-runtime v0.9.0
)
