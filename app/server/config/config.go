package config

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	charset        = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	replIDSize     = 40
	defaultPort    = "6379"
	defaultDir     = "/tmp/redis-files"
	defaultRDBFile = "dump.rdb"
)

const (
	RoleMaster = "master"
	RoleSlave  = "slave"
)

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

type Config struct {
	port        string
	role        string
	replicaOf   string
	replID      string
	replOffset  int
	slaves      []*Slave
	dir         string
	rdbFileName string
}

type Option func(c *Config)

func NewConfig(options ...Option) *Config {
	config := &Config{
		port:        defaultPort,
		role:        RoleMaster,
		replID:      generateReplicationID(),
		replOffset:  0,
		slaves:      []*Slave{},
		dir:         defaultDir,
		rdbFileName: defaultRDBFile,
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

func (c *Config) ReplicaOf() string {
	return c.replicaOf
}

func (c *Config) Dir() string {
	return c.dir
}

func (c *Config) RDBFileName() string {
	return c.rdbFileName
}

func (c *Config) Slaves() []*Slave {
	return c.slaves
}

func (c *Config) AddSlave(slave *Slave) {
	c.slaves = append(c.slaves, slave)
}

func (c *Config) UpdateOffset(bytes int) {
	c.replOffset += bytes
}

func (c *Config) RDBFilePath() string {
	return fmt.Sprintf("%s/%s", c.dir, c.rdbFileName)
}

func WithPort(port string) Option {
	return func(c *Config) {
		c.port = port
	}
}

func WithSlave(master string) Option {
	return func(c *Config) {
		c.role = RoleSlave
		c.replicaOf = master
	}
}

func WithDir(dir string) Option {
	return func(c *Config) {
		c.dir = dir
	}
}

func WithRDBFileName(fileName string) Option {
	return func(c *Config) {
		c.rdbFileName = fileName
	}
}

func generateReplicationID() string {
	b := make([]byte, replIDSize)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
