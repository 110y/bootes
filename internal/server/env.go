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
	XDSGRPCPort             int  `envconfig:"XDS_GRPC_PORT" required:"true"`
	XDSGRPCEnableChannelz   bool `envconfig:"XDS_GRPC_ENABLE_CHANNELZ"`
	XDSGRPCEnableReflection bool `envconfig:"XDS_GRPC_ENABLE_REFLECTION"`

	K8SMetricsServerPort int `envconfig:"K8S_METRICS_SERVER_PORT" required:"true"`
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
