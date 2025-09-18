package _redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Connection struct {
	config Config
	client *redis.Client
}

// NewConnection creates a new Redis connection
func NewConnection(config Config) *Connection {
	return &Connection{
		config: config,
	}
}

// Connect establishes connection to Redis
func (c *Connection) Connect(ctx context.Context) error {
	if err := c.config.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	connectCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	c.client = redis.NewClient(&redis.Options{
		Addr:     c.config.GetDNS(),
		Password: c.config.Password,
		DB:       c.config.DB,
	})

	// Test the connection
	if err := c.client.Ping(connectCtx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return nil
}

// Close closes the Redis connection
func (c *Connection) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// Ping checks if the Redis connection is alive
func (c *Connection) Ping(ctx context.Context) error {
	if c.client == nil {
		return fmt.Errorf("Redis client not connected")
	}
	return c.client.Ping(ctx).Err()
}

// IsHealthy checks if the Redis connection is healthy
func (c *Connection) IsHealthy(ctx context.Context) bool {
	return c.Ping(ctx) == nil
}

// GetClient returns the underlying Redis client
func (c *Connection) GetClient() *redis.Client {
	return c.client
}
