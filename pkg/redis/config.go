package _redis

import (
	"errors"
	"strconv"
)

// Config represents configuration for a single Redis node
type Config struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// ClusterConfig represents configuration for a Redis cluster
type ClusterConfig struct {
	// Addresses is a list of Redis node addresses in the format "host:port"
	Addresses []string
	// Password for the Redis cluster
	Password string
	// RouteByLatency allows routing read-only commands to the closest master or replica node
	RouteByLatency bool
	// RouteRandomly allows routing read-only commands randomly across master and replica nodes
	RouteRandomly bool
	// MaxRedirects limits the number of redirects before giving up
	MaxRedirects int
	// UseSlaveConnection indicates whether to create a separate client for slave/read-only operations
	UseSlaveConnection bool
	// SlaveReadOnly forces slave client to read from replicas
	SlaveReadOnly bool
}

// SentinelConfig represents configuration for Redis Sentinel
type SentinelConfig struct {
	// MasterName is the name of the master in Sentinel configuration
	MasterName string
	// SentinelAddresses is a list of Sentinel addresses in the format "host:port"
	SentinelAddresses []string
	// Password for the Redis master
	Password string
	// DB is the database to select
	DB int
	// SentinelPassword is the password for Sentinel if different from Redis password
	SentinelPassword string
	// UseSlaveConnection indicates whether to establish a connection to a slave/replica
	UseSlaveConnection bool
	// SlaveReadOnly forces slave connection to be read-only (recommended)
	SlaveReadOnly bool
}

func DefaultConfig() Config {
	return Config{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		DB:       0,
	}
}

func DefaultClusterConfig() ClusterConfig {
	return ClusterConfig{
		Addresses:          []string{"localhost:6379"},
		Password:           "",
		RouteByLatency:     false,
		RouteRandomly:      false,
		MaxRedirects:       3,
		UseSlaveConnection: false,
		SlaveReadOnly:      true,
	}
}

func DefaultSentinelConfig() SentinelConfig {
	return SentinelConfig{
		MasterName:         "mymaster",
		SentinelAddresses:  []string{"localhost:26379"},
		Password:           "",
		DB:                 0,
		SentinelPassword:   "",
		UseSlaveConnection: false,
		SlaveReadOnly:      true,
	}
}

func (c *Config) Validate() error {
	if c.Host == "" {
		return errors.New("host is required")
	}
	if c.Port == 0 {
		return errors.New("port is required")
	}
	if c.DB < 0 {
		return errors.New("db must be greater than or equal to 0")
	}
	return nil
}

func (c *Config) GetDNS() string {
	return c.Host + ":" + strconv.Itoa(c.Port)
}

func (c *ClusterConfig) Validate() error {
	if len(c.Addresses) == 0 {
		return errors.New("at least one address is required")
	}
	if c.MaxRedirects < 0 {
		return errors.New("max redirects must be greater than or equal to 0")
	}
	return nil
}

func (c *SentinelConfig) Validate() error {
	if c.MasterName == "" {
		return errors.New("master name is required")
	}
	if len(c.SentinelAddresses) == 0 {
		return errors.New("at least one sentinel address is required")
	}
	if c.DB < 0 {
		return errors.New("db must be greater than or equal to 0")
	}
	return nil
}
