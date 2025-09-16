package gorm

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_config "go-libs/pkg/database/postgres/config"
	"go-libs/pkg/database/postgres/interfaces"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var _ interfaces.GormClient = (*Connection)(nil)

// Connection implements the GormClient interface
type Connection struct {
	db     *gorm.DB
	sqlDB  *sql.DB
	config *_config.GormConfig
}

// NewConnection creates a new GORM connection
func NewConnection(cfg *_config.GormConfig) *Connection {
	return &Connection{
		config: cfg,
	}
}

// Connect establishes connection to PostgreSQL using GORM
func (c *Connection) Connect(ctx context.Context) error {
	if err := c.config.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	var db *gorm.DB
	var err error

	// Retry connection with exponential backoff
	for i := 0; i <= c.config.MaxRetries; i++ {
		db, err = c.connectWithTimeout(ctx)
		if err == nil {
			break
		}

		if i < c.config.MaxRetries {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(c.config.RetryInterval * time.Duration(i+1)):
				continue
			}
		}
	}

	if err != nil {
		return fmt.Errorf("failed to connect after %d retries: %w", c.config.MaxRetries, err)
	}

	c.db = db

	// Get underlying SQL DB for connection pool configuration
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	c.sqlDB = sqlDB

	// Configure connection pool
	c.configureConnectionPool()

	return nil
}

func (c *Connection) connectWithTimeout(ctx context.Context) (*gorm.DB, error) {
	connectCtx, cancel := context.WithTimeout(ctx, c.config.ConnectTimeout)
	defer cancel()

	gormConfig := &gorm.Config{
		Logger:                                   c.getLogger(),
		SkipDefaultTransaction:                   c.config.SkipDefaultTransaction,
		PrepareStmt:                              c.config.PrepareStmt,
		DisableForeignKeyConstraintWhenMigrating: c.config.DisableForeignKeyConstraintWhenMigrating,
	}

	db, err := gorm.Open(postgres.Open(c.config.DSN()), gormConfig)
	if err != nil {
		return nil, err
	}

	// Test connection
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	if err := sqlDB.PingContext(connectCtx); err != nil {
		sqlDB.Close()
		return nil, err
	}

	return db, nil
}

func (c *Connection) configureConnectionPool() {
	c.sqlDB.SetMaxOpenConns(c.config.MaxOpenConns)
	c.sqlDB.SetMaxIdleConns(c.config.MaxIdleConns)
	c.sqlDB.SetConnMaxLifetime(c.config.ConnMaxLifetime)
	c.sqlDB.SetConnMaxIdleTime(c.config.ConnMaxIdleTime)
}

func (c *Connection) getLogger() logger.Interface {
	logLevel := logger.LogLevel(c.config.LogLevel)
	return logger.New(
		nil, // Use default writer (stdout)
		logger.Config{
			SlowThreshold:             c.config.SlowThreshold,
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)
}

// Close closes the database connection
func (c *Connection) Close() error {
	if c.sqlDB != nil {
		return c.sqlDB.Close()
	}
	return nil
}

// Ping checks if the database connection is alive
func (c *Connection) Ping(ctx context.Context) error {
	if c.sqlDB == nil {
		return fmt.Errorf("database not connected")
	}
	return c.sqlDB.PingContext(ctx)
}

// IsHealthy checks if the database connection is healthy
func (c *Connection) IsHealthy(ctx context.Context) bool {
	return c.Ping(ctx) == nil
}

// GetDB returns the underlying GORM DB instance
func (c *Connection) GetDB() *gorm.DB {
	return c.db
}

// Stats returns connection statistics
func (c *Connection) Stats() interfaces.ConnectionStats {
	if c.sqlDB == nil {
		return interfaces.ConnectionStats{}
	}

	stats := c.sqlDB.Stats()
	return interfaces.ConnectionStats{
		OpenConnections:   stats.OpenConnections,
		InUseConnections:  stats.InUse,
		IdleConnections:   stats.Idle,
		WaitCount:         stats.WaitCount,
		WaitDuration:      stats.WaitDuration,
		MaxIdleClosed:     stats.MaxIdleClosed,
		MaxIdleTimeClosed: stats.MaxIdleTimeClosed,
		MaxLifetimeClosed: stats.MaxLifetimeClosed,
	}
}
