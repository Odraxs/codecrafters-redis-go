package config

type config struct {
	port string
}

type Option func(c *config)

func NewConfig(options ...Option) *config {
	config := &config{
		port: "6379",
	}

	for _, opt := range options {
		opt(config)
	}

	return config
}

func (c *config) Port() string {
	return c.port
}

func WithPort(port string) Option {
	return func(c *config) {
		c.port = port
	}
}
