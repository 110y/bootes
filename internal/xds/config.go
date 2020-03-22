package xds

type Config struct {
	Port                 int
	EnableGRPCChannelz   bool
	EnableGRPCReflection bool
}

func (c *Config) validate() error {
	return nil
}
