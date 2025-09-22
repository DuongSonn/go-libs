package _redis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// RedisClient defines the common interface for all Redis client types
type RedisClient interface {
	// Connect establishes connection to Redis
	Connect(ctx context.Context) error

	// Close closes the Redis connection
	Close() error

	// Ping checks if the Redis connection is alive
	Ping(ctx context.Context) error

	// IsHealthy checks if the Redis connection is healthy
	IsHealthy(ctx context.Context) bool
}

// SingleNodeClient defines the interface for a single Redis node client
type SingleNodeClient interface {
	RedisClient

	// GetClient returns the underlying Redis client
	GetClient() *redis.Client
}

// ClusterClient defines the interface for a Redis cluster client
type ClusterClient interface {
	RedisClient

	// GetMasterClient returns the underlying Redis cluster client for write operations
	GetMasterClient() *redis.ClusterClient

	// GetSlaveClient returns a cluster client configured for read-only operations
	GetSlaveClient() *redis.ClusterClient

	// HasSlaveConnected returns true if a slave client is available
	HasSlaveConnected() bool
}

// SentinelClient defines the interface for a Redis sentinel client
type SentinelClient interface {
	RedisClient

	// GetMasterClient returns the underlying Redis failover client (master)
	GetMasterClient() *redis.Client

	// GetSlaveClient returns a client connected to a replica/slave node (read-only)
	GetSlaveClient() *redis.Client

	// HasSlaveConnected returns true if a slave connection is available
	HasSlaveConnected() bool
}
