package _rabbitmq

import (
	"errors"
	"fmt"
	"time"
)

// Config holds the configuration for RabbitMQ connection
type Config struct {
	Host         string        `json:"host" yaml:"host"`
	Port         int           `json:"port" yaml:"port"`
	Username     string        `json:"username" yaml:"username"`
	Password     string        `json:"password" yaml:"password"`
	VHost        string        `json:"vhost" yaml:"vhost"`
	ConnTimeout  time.Duration `json:"conn_timeout" yaml:"conn_timeout"`
	RetryTimeout time.Duration `json:"retry_timeout" yaml:"retry_timeout"`
	MaxRetries   int           `json:"max_retries" yaml:"max_retries"`
}

// DefaultConfig returns a default configuration for RabbitMQ
func DefaultConfig() *Config {
	return &Config{
		Host:         "localhost",
		Port:         5672,
		Username:     "guest",
		Password:     "guest",
		VHost:        "/",
		ConnTimeout:  5 * time.Second,
		RetryTimeout: 2 * time.Second,
		MaxRetries:   3,
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Host == "" {
		return errors.New("host is required")
	}
	if c.Port <= 0 {
		return errors.New("port must be greater than 0")
	}
	if c.Username == "" {
		return errors.New("username is required")
	}
	if c.Password == "" {
		return errors.New("password is required")
	}
	return nil
}

// GetURI returns the RabbitMQ connection URI
func (c *Config) GetURI() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
		c.Username, c.Password, c.Host, c.Port, c.VHost)
}

// ExchangeConfig holds the configuration for a RabbitMQ exchange
type ExchangeConfig struct {
	Name       string `json:"name" yaml:"name"`
	Type       string `json:"type" yaml:"type"`
	Durable    bool   `json:"durable" yaml:"durable"`
	AutoDelete bool   `json:"auto_delete" yaml:"auto_delete"`
	Internal   bool   `json:"internal" yaml:"internal"`
	NoWait     bool   `json:"no_wait" yaml:"no_wait"`
}

// QueueConfig holds the configuration for a RabbitMQ queue
type QueueConfig struct {
	Name       string `json:"name" yaml:"name"`
	Durable    bool   `json:"durable" yaml:"durable"`
	AutoDelete bool   `json:"auto_delete" yaml:"auto_delete"`
	Exclusive  bool   `json:"exclusive" yaml:"exclusive"`
	NoWait     bool   `json:"no_wait" yaml:"no_wait"`
}

// BindingConfig holds the configuration for binding a queue to an exchange
type BindingConfig struct {
	Exchange   string `json:"exchange" yaml:"exchange"`
	Queue      string `json:"queue" yaml:"queue"`
	RoutingKey string `json:"routing_key" yaml:"routing_key"`
	NoWait     bool   `json:"no_wait" yaml:"no_wait"`
}
