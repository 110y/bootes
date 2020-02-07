package xds

type Config struct {
	Port             uint16
	EnableReflection bool
}

func (c *Config) validate() error {
	return nil
}
