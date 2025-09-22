package _gorm_postgres

import (
	"context"
	"fmt"
	"sync"
	"time"

	_postgres "go-libs/pkg/postgres"

	"gorm.io/gorm"
)

var _ _postgres.GormMasterSlaveClient = (*MasterSlaveConnection)(nil)

// MasterSlaveConnection implements the GormMasterSlaveClient interface
type MasterSlaveConnection struct {
	config       *_postgres.GormMasterSlaveConfig
	masterConn   *Connection
	slaveConn    *Connection
	role         string // "master" or "slave"
	mu           sync.RWMutex
	healthTicker *time.Ticker
	stopChan     chan struct{}
}

// NewMasterSlaveConnection creates a new master-slave connection
func NewMasterSlaveConnection(cfg *_postgres.GormMasterSlaveConfig) *MasterSlaveConnection {
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
	masterGormConfig := c.config.GetMasterGormConfig()
	c.masterConn = NewConnection(masterGormConfig)
	if err := c.masterConn.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to master: %w", err)
	}

	// Connect to slave if enabled
	if c.config.UseSlaveConnection {
		slaveGormConfig := c.config.GetSlaveGormConfig()
		c.slaveConn = NewConnection(slaveGormConfig)
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

// GetDB returns the underlying GORM DB instance (master DB)
func (c *MasterSlaveConnection) GetDB() *gorm.DB {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.masterConn != nil {
		return c.masterConn.GetDB()
	}
	return nil
}

// GetMasterDB returns the master DB
func (c *MasterSlaveConnection) GetMasterDB() *gorm.DB {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.masterConn != nil {
		return c.masterConn.GetDB()
	}
	return nil
}

// GetSlaveDB returns the slave DB
func (c *MasterSlaveConnection) GetSlaveDB() *gorm.DB {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.slaveConn != nil {
		return c.slaveConn.GetDB()
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
	masterGormConfig := c.config.GetMasterGormConfig()

	for i := 0; i < c.config.FailoverRetries; i++ {
		// Create a new connection
		conn := NewConnection(masterGormConfig)
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
	slaveGormConfig := c.config.GetSlaveGormConfig()

	for i := 0; i < c.config.FailoverRetries; i++ {
		// Create a new connection
		conn := NewConnection(slaveGormConfig)
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
