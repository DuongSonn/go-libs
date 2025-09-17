package _pgx_postgres

import (
	"context"
	"fmt"
	"time"

	_postgres "go-libs/pkg/postgres"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ _postgres.PgxClient = (*Connection)(nil)

// Connection implements the PgxClient interface
type Connection struct {
	pool   *pgxpool.Pool
	conn   *pgx.Conn
	config *_postgres.Config
}

// NewConnection creates a new pgx connection
func NewConnection(cfg *_postgres.Config) *Connection {
	return &Connection{
		config: cfg,
	}
}

// Connect establishes connection to PostgreSQL using pgx
func (c *Connection) Connect(ctx context.Context) error {
	if err := c.config.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	var pool *pgxpool.Pool
	var conn *pgx.Conn
	var err error

	// Retry connection with exponential backoff
	for i := 0; i <= c.config.MaxRetries; i++ {
		pool, conn, err = c.connectWithTimeout(ctx)
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

	c.pool = pool
	c.conn = conn

	return nil
}

func (c *Connection) connectWithTimeout(ctx context.Context) (*pgxpool.Pool, *pgx.Conn, error) {
	connectCtx, cancel := context.WithTimeout(ctx, c.config.ConnectTimeout)
	defer cancel()

	// Create connection pool
	poolConfig, err := pgxpool.ParseConfig(c.config.DSN())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse pool config: %w", err)
	}

	// Configure pool settings
	poolConfig.MaxConns = int32(c.config.MaxOpenConns)
	poolConfig.MinConns = int32(c.config.MaxIdleConns)
	poolConfig.MaxConnLifetime = c.config.ConnMaxLifetime
	poolConfig.MaxConnIdleTime = c.config.ConnMaxIdleTime

	pool, err := pgxpool.NewWithConfig(connectCtx, poolConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test pool connection
	if err := pool.Ping(connectCtx); err != nil {
		pool.Close()
		return nil, nil, fmt.Errorf("failed to ping pool: %w", err)
	}

	// Create direct connection
	conn, err := pgx.Connect(connectCtx, c.config.DSN())
	if err != nil {
		pool.Close()
		return nil, nil, fmt.Errorf("failed to create direct connection: %w", err)
	}

	// Test direct connection
	if err := conn.Ping(connectCtx); err != nil {
		pool.Close()
		conn.Close(connectCtx)
		return nil, nil, fmt.Errorf("failed to ping direct connection: %w", err)
	}

	return pool, conn, nil
}

// Close closes the database connections
func (c *Connection) Close() error {
	if c.pool != nil {
		c.pool.Close()
	}
	if c.conn != nil {
		return c.conn.Close(context.Background())
	}
	return nil
}

// Ping checks if the database connection is alive
func (c *Connection) Ping(ctx context.Context) error {
	if c.pool == nil {
		return fmt.Errorf("database not connected")
	}
	return c.pool.Ping(ctx)
}

// IsHealthy checks if the database connection is healthy
func (c *Connection) IsHealthy(ctx context.Context) bool {
	return c.Ping(ctx) == nil
}

// GetPool returns the underlying pgx pool instance
func (c *Connection) GetPool() *pgxpool.Pool {
	return c.pool
}

// GetConn returns the underlying pgx connection instance
func (c *Connection) GetConn() *pgx.Conn {
	return c.conn
}

// Stats returns connection statistics
func (c *Connection) Stats() _postgres.ConnectionStats {
	if c.pool == nil {
		return _postgres.ConnectionStats{}
	}

	stats := c.pool.Stat()
	return _postgres.ConnectionStats{
		OpenConnections:   int(stats.TotalConns()),
		InUseConnections:  int(stats.AcquiredConns()),
		IdleConnections:   int(stats.IdleConns()),
		WaitCount:         0, // pgx doesn't provide this directly
		WaitDuration:      0, // pgx doesn't provide this directly
		MaxIdleClosed:     0, // pgx doesn't provide this directly
		MaxIdleTimeClosed: 0, // pgx doesn't provide this directly
		MaxLifetimeClosed: 0, // pgx doesn't provide this directly
	}
}
