package _redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// SentinelConnection represents a connection to Redis using Sentinel for high availability
type SentinelConnection struct {
	config       SentinelConfig
	masterClient *redis.Client
	slaveClient  *redis.Client
}

// NewSentinelConnection creates a new Redis connection using Sentinel
func NewSentinelConnection(config SentinelConfig) *SentinelConnection {
	return &SentinelConnection{
		config: config,
	}
}

// Connect establishes connection to Redis using Sentinel
func (c *SentinelConnection) Connect(ctx context.Context) error {
	if err := c.config.Validate(); err != nil {
		return fmt.Errorf("invalid sentinel config: %w", err)
	}

	connectCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Create a failover client that uses Sentinel for automatic master discovery and failover
	c.masterClient = redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:       c.config.MasterName,
		SentinelAddrs:    c.config.SentinelAddresses,
		Password:         c.config.Password,
		DB:               c.config.DB,
		SentinelPassword: c.config.SentinelPassword,
	})

	// Test the master connection
	if err := c.masterClient.Ping(connectCtx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis master using Sentinel: %w", err)
	}

	// If slave connection is requested, establish it
	if c.config.UseSlaveConnection {
		// Create a failover client that connects to slaves for read-only operations
		c.slaveClient = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:       c.config.MasterName,
			SentinelAddrs:    c.config.SentinelAddresses,
			Password:         c.config.Password,
			DB:               c.config.DB,
			SentinelPassword: c.config.SentinelPassword,
			ReplicaOnly:      true, // Connect only to replicas
		})

		// Test the slave connection
		if err := c.slaveClient.Ping(connectCtx).Err(); err != nil {
			// Close the master connection since we're failing
			c.masterClient.Close()
			return fmt.Errorf("failed to connect to Redis slave using Sentinel: %w", err)
		}
	}

	return nil
}

// Close closes the Redis connection
func (c *SentinelConnection) Close() error {
	var masterErr, slaveErr error

	if c.masterClient != nil {
		masterErr = c.masterClient.Close()
	}

	if c.slaveClient != nil {
		slaveErr = c.slaveClient.Close()
	}

	// Return the first error encountered
	if masterErr != nil {
		return fmt.Errorf("error closing master connection: %w", masterErr)
	}
	if slaveErr != nil {
		return fmt.Errorf("error closing slave connection: %w", slaveErr)
	}

	return nil
}

// Ping checks if the Redis connection is alive
func (c *SentinelConnection) Ping(ctx context.Context) error {
	if c.masterClient == nil {
		return fmt.Errorf("redis client not connected")
	}
	return c.masterClient.Ping(ctx).Err()
}

// IsHealthy checks if the Redis connection is healthy
func (c *SentinelConnection) IsHealthy(ctx context.Context) bool {
	return c.Ping(ctx) == nil
}

// GetMasterClient returns the underlying Redis failover client (master)
func (c *SentinelConnection) GetMasterClient() *redis.Client {
	return c.masterClient
}

// GetSlaveClient returns a client connected to a replica/slave node (read-only)
func (c *SentinelConnection) GetSlaveClient() *redis.Client {
	return c.slaveClient
}

// HasSlaveConnected returns true if a slave connection is available
func (c *SentinelConnection) HasSlaveConnected() bool {
	return c.slaveClient != nil
}

// Ensure SentinelConnection implements SentinelClient interface
var _ SentinelClient = (*SentinelConnection)(nil)
