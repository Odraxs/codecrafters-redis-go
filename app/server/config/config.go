package config

type Config struct {
	port      string
	role      string
	replicaOf string
}

type Option func(c *Config)

func NewConfig(options ...Option) *Config {
	config := &Config{
		port: "6379",
		role: "master",
	}

	for _, opt := range options {
		opt(config)
	}

	return config
}

func (c *Config) Port() string {
	return c.port
}

func (c *Config) Role() string {
	return c.role
}

func WithPort(port string) Option {
	return func(c *Config) {
		c.port = port
	}
}

func WithSlave(master string) Option {
	return func(c *Config) {
		c.role = "slave"
		c.replicaOf = master
	}
}
