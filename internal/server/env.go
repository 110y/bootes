package server

import (
	"sync"

	"github.com/kelseyhightower/envconfig"
)

var (
	environmentsMu = sync.Mutex{}
	envs           *environments
)

type environments struct {
	HealthProbeServerPort int `envconfig:"HEALTH_PROBE_SERVER_PORT" required:"true"`

	XDSGRPCPort             int  `envconfig:"XDS_GRPC_PORT" required:"true"`
	XDSGRPCEnableChannelz   bool `envconfig:"XDS_GRPC_ENABLE_CHANNELZ"`
	XDSGRPCEnableReflection bool `envconfig:"XDS_GRPC_ENABLE_REFLECTION"`

	K8SWebhookServerPort int `envconfig:"K8S_WEBHOOK_SERVER_PORT" required:"true"`
	K8SMetricsServerPort int `envconfig:"K8S_METRICS_SERVER_PORT" required:"true"`

	TraceUseStdout              bool   `envconfig:"TRACE_USE_STDOUT"`
	TraceUseJaeger              bool   `envconfig:"TRACE_USE_JAEGER"`
	TraceJaegerEndpoint         string `envconfig:"TRACE_JAEGER_ENDPOINT"`
	TraceUseGCPCloudTrace       bool   `envconfig:"TRACE_USE_GCP_CLOUD_TRACE"`
	TraceGCPCloudTraceProjectID string `envconfig:"TRACE_GCP_CLOUD_TRACE_PROJECT_ID"`
}

func getEnvironments() (*environments, error) {
	environmentsMu.Lock()
	defer environmentsMu.Unlock()

	if envs == nil {
		var e environments
		if err := envconfig.Process("", &e); err != nil {
			return nil, err
		}

		envs = &e
	}

	return envs, nil
}
