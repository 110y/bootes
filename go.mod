module github.com/110y/bootes

go 1.14

require (
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace v0.1.1-0.20200514210843-966afdc5d38c
	github.com/GoogleContainerTools/kpt v0.24.0
	github.com/envoyproxy/go-control-plane v0.9.5
	github.com/go-delve/delve v1.4.0
	github.com/go-logr/logr v0.1.0
	github.com/go-logr/zapr v0.1.1
	github.com/golang/protobuf v1.3.5
	github.com/google/go-cmp v0.4.0
	github.com/google/uuid v1.1.1
	github.com/kelseyhightower/envconfig v1.4.0
	go.opentelemetry.io/otel v0.5.0
	go.opentelemetry.io/otel/exporters/trace/jaeger v0.5.0
	go.uber.org/zap v1.14.0
	google.golang.org/grpc v1.29.1
	k8s.io/api v0.17.3
	k8s.io/apimachinery v0.17.3
	k8s.io/client-go v0.17.3
	sigs.k8s.io/controller-runtime v0.5.1
	sigs.k8s.io/controller-tools v0.2.7
	sigs.k8s.io/kind v0.7.0
)
