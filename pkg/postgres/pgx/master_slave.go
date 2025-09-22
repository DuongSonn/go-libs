package _pgx_postgres

import (
	"context"
	"fmt"
	"sync"
	"time"

	_postgres "go-libs/pkg/postgres"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ _postgres.PgxMasterSlaveClient = (*MasterSlaveConnection)(nil)

// MasterSlaveConnection implements the PgxMasterSlaveClient interface
type MasterSlaveConnection struct {
	config       *_postgres.MasterSlaveConfig
	masterConn   *Connection
	slaveConn    *Connection
	role         string // "master" or "slave"
	mu           sync.RWMutex
	healthTicker *time.Ticker
	stopChan     chan struct{}
}

// NewMasterSlaveConnection creates a new master-slave connection
func NewMasterSlaveConnection(cfg *_postgres.MasterSlaveConfig) *MasterSlaveConnection {
	return &MasterSlaveConnection{
		config:   cfg,
		role:     "master", // Default role is master
		stopChan: make(chan struct{}),
	}
}

// Connect establishes connections to both master and slave (if configured)
func (c *MasterSlaveConnection) Connect(ctx context.Context) error {
	if err := c.config.Validate(); err != nil {
		return fmt.Errorf("invalid master-slave config: %w", err)
	}

	// Connect to master
	c.masterConn = NewConnection(c.config.Master)
	if err := c.masterConn.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to master: %w", err)
	}

	// Connect to slave if enabled
	if c.config.UseSlaveConnection {
		c.slaveConn = NewConnection(c.config.Slave)
		if err := c.slaveConn.Connect(ctx); err != nil {
			// Close master connection
			c.masterConn.Close()
			c.masterConn = nil
			return fmt.Errorf("failed to connect to slave: %w", err)
		}
	}

	// Start health check if enabled
	if c.config.HealthCheckEnabled {
		c.startHealthCheck()
	}

	return nil
}

// Close closes all connections
func (c *MasterSlaveConnection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Stop health check
	if c.healthTicker != nil {
		c.healthTicker.Stop()
		close(c.stopChan)
	}

	var masterErr, slaveErr error

	// Close master connection
	if c.masterConn != nil {
		masterErr = c.masterConn.Close()
		c.masterConn = nil
	}

	// Close slave connection
	if c.slaveConn != nil {
		slaveErr = c.slaveConn.Close()
		c.slaveConn = nil
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

// Ping checks if the master connection is alive
func (c *MasterSlaveConnection) Ping(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.masterConn == nil {
		return fmt.Errorf("master connection not established")
	}

	return c.masterConn.Ping(ctx)
}

// IsHealthy checks if the connection is healthy
func (c *MasterSlaveConnection) IsHealthy(ctx context.Context) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Check master health
	if c.masterConn != nil && c.masterConn.IsHealthy(ctx) {
		return true
	}

	// If master is down but slave is healthy and auto-failover is enabled
	if c.config.AutoFailover && c.slaveConn != nil && c.slaveConn.IsHealthy(ctx) {
		return true
	}

	return false
}

// BeginTx begins a transaction on the master
func (c *MasterSlaveConnection) BeginTx(ctx context.Context) (_postgres.Transaction, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.masterConn == nil {
		return nil, fmt.Errorf("master connection not established")
	}

	return c.masterConn.BeginTx(ctx)
}

// GetMasterClient returns the master client
func (c *MasterSlaveConnection) GetMasterClient() _postgres.DatabaseClient {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.masterConn
}

// GetSlaveClient returns the slave client
func (c *MasterSlaveConnection) GetSlaveClient() _postgres.DatabaseClient {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.slaveConn
}

// HasSlaveConnected returns true if a slave connection is available
func (c *MasterSlaveConnection) HasSlaveConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.slaveConn != nil
}

// IsMaster returns true if this connection is a master
func (c *MasterSlaveConnection) IsMaster() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.role == "master"
}

// IsSlave returns true if this connection is a slave
func (c *MasterSlaveConnection) IsSlave() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.role == "slave"
}

// GetPool returns the underlying pgx pool (master pool)
func (c *MasterSlaveConnection) GetPool() *pgxpool.Pool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.masterConn != nil {
		return c.masterConn.GetPool()
	}
	return nil
}

// GetConn returns the underlying pgx connection (master connection)
func (c *MasterSlaveConnection) GetConn() *pgx.Conn {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.masterConn != nil {
		return c.masterConn.GetConn()
	}
	return nil
}

// GetMasterPool returns the master pool
func (c *MasterSlaveConnection) GetMasterPool() *pgxpool.Pool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.masterConn != nil {
		return c.masterConn.GetPool()
	}
	return nil
}

// GetSlavePool returns the slave pool
func (c *MasterSlaveConnection) GetSlavePool() *pgxpool.Pool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.slaveConn != nil {
		return c.slaveConn.GetPool()
	}
	return nil
}

// Stats returns connection statistics for the master
func (c *MasterSlaveConnection) Stats() _postgres.ConnectionStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.masterConn != nil {
		return c.masterConn.Stats()
	}
	return _postgres.ConnectionStats{}
}

// InsertModel inserts a model into the master database
func (c *MasterSlaveConnection) InsertModel(ctx context.Context, model any) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.masterConn != nil {
		return c.masterConn.InsertModel(ctx, model)
	}
	return fmt.Errorf("master connection not established")
}

// UpsertModel upserts a model into the master database
func (c *MasterSlaveConnection) UpsertModel(ctx context.Context, model any, primaryKeys ...string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.masterConn != nil {
		return c.masterConn.UpsertModel(ctx, model, primaryKeys...)
	}
	return fmt.Errorf("master connection not established")
}

// BatchInsertModel batch inserts models into the master database
func (c *MasterSlaveConnection) BatchInsertModel(ctx context.Context, models []any, batchSize int) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.masterConn != nil {
		return c.masterConn.BatchInsertModel(ctx, models, batchSize)
	}
	return fmt.Errorf("master connection not established")
}

// startHealthCheck starts a periodic health check
func (c *MasterSlaveConnection) startHealthCheck() {
	c.healthTicker = time.NewTicker(c.config.HealthCheckInterval)
	go func() {
		for {
			select {
			case <-c.stopChan:
				return
			case <-c.healthTicker.C:
				c.checkHealth()
			}
		}
	}()
}

// checkHealth checks the health of master and slave connections
func (c *MasterSlaveConnection) checkHealth() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check master health
	masterHealthy := c.masterConn != nil && c.masterConn.IsHealthy(ctx)
	slaveHealthy := c.slaveConn != nil && c.slaveConn.IsHealthy(ctx)

	// If master is down but slave is healthy and auto-failover is enabled
	if !masterHealthy && slaveHealthy && c.config.AutoFailover {
		// Promote slave to master
		c.role = "slave" // This connection is now operating in slave mode
		fmt.Println("Master connection is down, operating in slave-only mode")
	} else if masterHealthy && !slaveHealthy && c.slaveConn != nil {
		// Try to reconnect to slave
		fmt.Println("Slave connection is down, attempting to reconnect")
		c.attemptSlaveReconnect(ctx)
	} else if !masterHealthy && !slaveHealthy {
		// Both connections are down, try to reconnect to both
		fmt.Println("Both master and slave connections are down, attempting to reconnect")
		c.attemptMasterReconnect(ctx)
		if c.slaveConn != nil {
			c.attemptSlaveReconnect(ctx)
		}
	}
}

// attemptMasterReconnect attempts to reconnect to the master
func (c *MasterSlaveConnection) attemptMasterReconnect(ctx context.Context) {
	for i := 0; i < c.config.FailoverRetries; i++ {
		// Create a new connection
		conn := NewConnection(c.config.Master)
		if err := conn.Connect(ctx); err == nil {
			// Successfully reconnected
			if c.masterConn != nil {
				c.masterConn.Close()
			}
			c.masterConn = conn
			c.role = "master"
			fmt.Println("Successfully reconnected to master")
			return
		}
		time.Sleep(c.config.FailoverInterval)
	}
	fmt.Println("Failed to reconnect to master after multiple attempts")
}

// attemptSlaveReconnect attempts to reconnect to the slave
func (c *MasterSlaveConnection) attemptSlaveReconnect(ctx context.Context) {
	for i := 0; i < c.config.FailoverRetries; i++ {
		// Create a new connection
		conn := NewConnection(c.config.Slave)
		if err := conn.Connect(ctx); err == nil {
			// Successfully reconnected
			if c.slaveConn != nil {
				c.slaveConn.Close()
			}
			c.slaveConn = conn
			fmt.Println("Successfully reconnected to slave")
			return
		}
		time.Sleep(c.config.FailoverInterval)
	}
	fmt.Println("Failed to reconnect to slave after multiple attempts")
}
