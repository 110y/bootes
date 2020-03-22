package server

import (
	"sync"

	"github.com/kelseyhightower/envconfig"
)

var (
	environmentsMutext = sync.Mutex{}
	envs               *environments
)

type environments struct {
	XDSGRPCPort             int  `envconfig:"XDS_GRPC_PORT" required:"true"`
	XDSGRPCEnableChannelz   bool `envconfig:"XDS_GRPC_ENABLE_CHANNELZ" required:"true"`
	XDSGRPCEnableReflection bool `envconfig:"XDS_GRPC_ENABLE_REFLECTION required:"true"`
}

func getEnvironments() (*environments, error) {
	environmentsMutext.Lock()
	defer environmentsMutext.Unlock()

	if envs == nil {
		var e environments
		if err := envconfig.Process("", &e); err != nil {
			return nil, err
		}

		envs = &e
	}

	return envs, nil
}
