package xds

type Config struct {
	Port                 uint32
	EnableGRPCChannelz   bool
	EnableGRPCReflection bool
}

func (c *Config) validate() error {
	return nil
}
