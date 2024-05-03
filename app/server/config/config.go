package config

import (
	"math/rand"
	"time"
)

const (
	charset    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	replIDSize = 40
)

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

type Config struct {
	port       string
	role       string
	replicaOf  string
	replID     string
	replOffset int
}

type Option func(c *Config)

func NewConfig(options ...Option) *Config {
	config := &Config{
		port:       "6379",
		role:       "master",
		replID:     generateReplicationID(),
		replOffset: 0,
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

func (c *Config) ReplID() string {
	return c.replID
}

func (c *Config) ReplOffset() int {
	return c.replOffset
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

func generateReplicationID() string {
	b := make([]byte, replIDSize)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
