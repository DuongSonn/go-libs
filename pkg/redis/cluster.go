package _redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ClusterConnection represents a connection to a Redis cluster
type ClusterConnection struct {
	config       ClusterConfig
	masterClient *redis.ClusterClient
	slaveClient  *redis.ClusterClient
}

// NewClusterConnection creates a new Redis cluster connection
func NewClusterConnection(config ClusterConfig) *ClusterConnection {
	return &ClusterConnection{
		config: config,
	}
}

// Connect establishes connection to Redis cluster
func (c *ClusterConnection) Connect(ctx context.Context) error {
	if err := c.config.Validate(); err != nil {
		return fmt.Errorf("invalid cluster config: %w", err)
	}

	connectCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Create the master client for write operations
	c.masterClient = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:          c.config.Addresses,
		Password:       c.config.Password,
		RouteByLatency: c.config.RouteByLatency,
		RouteRandomly:  c.config.RouteRandomly,
		MaxRedirects:   c.config.MaxRedirects,
	})

	// Test the master connection
	if err := c.masterClient.Ping(connectCtx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis cluster: %w", err)
	}

	// If slave connection is requested, create it
	if c.config.UseSlaveConnection {
		// Create a separate client for read-only operations
		c.slaveClient = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:          c.config.Addresses,
			Password:       c.config.Password,
			RouteByLatency: c.config.RouteByLatency || c.config.SlaveReadOnly, // Prefer replicas for read operations
			RouteRandomly:  c.config.RouteRandomly,
			MaxRedirects:   c.config.MaxRedirects,
			ReadOnly:       true, // This is the key setting for read-only mode
		})

		// Test the slave connection
		if err := c.slaveClient.Ping(connectCtx).Err(); err != nil {
			// Close the master connection since we're failing
			c.masterClient.Close()
			return fmt.Errorf("failed to connect to Redis cluster slave client: %w", err)
		}
	}

	return nil
}

// Close closes the Redis cluster connection
func (c *ClusterConnection) Close() error {
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

// Ping checks if the Redis cluster connection is alive
func (c *ClusterConnection) Ping(ctx context.Context) error {
	if c.masterClient == nil {
		return fmt.Errorf("redis cluster client not connected")
	}
	return c.masterClient.Ping(ctx).Err()
}

// IsHealthy checks if the Redis cluster connection is healthy
func (c *ClusterConnection) IsHealthy(ctx context.Context) bool {
	return c.Ping(ctx) == nil
}

// GetMasterClient returns the underlying Redis cluster client for write operations
func (c *ClusterConnection) GetMasterClient() *redis.ClusterClient {
	return c.masterClient
}

// GetSlaveClient returns a cluster client configured for read-only operations
func (c *ClusterConnection) GetSlaveClient() *redis.ClusterClient {
	return c.slaveClient
}

// HasSlaveConnected returns true if a slave client is available
func (c *ClusterConnection) HasSlaveConnected() bool {
	return c.slaveClient != nil
}

// Ensure ClusterConnection implements ClusterClient interface
var _ ClusterClient = (*ClusterConnection)(nil)
