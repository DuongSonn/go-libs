package _postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gorm.io/gorm"
)

// DatabaseClient defines the common interface for database operations
type DatabaseClient interface {
	// Connection management
	Connect(ctx context.Context) error
	Close() error
	Ping(ctx context.Context) error

	// Health check
	IsHealthy(ctx context.Context) bool

	// Transaction support
	BeginTx(ctx context.Context) (Transaction, error)
}

// Transaction interface for transaction operations
type Transaction interface {
	Commit() error
	Rollback() error
	Exec(ctx context.Context, query string, args ...any) error
	Query(ctx context.Context, query string, args ...any) (Rows, error)
	QueryRow(ctx context.Context, query string, args ...any) Row
}

// Rows interface for query result iteration
type Rows interface {
	Next() bool
	Scan(dest ...any) error
	Close() error
	Err() error
}

// Row interface for single row query results
type Row interface {
	Scan(dest ...any) error
}

// ConnectionStats provides database connection statistics
type ConnectionStats struct {
	OpenConnections   int           `json:"open_connections"`
	InUseConnections  int           `json:"in_use_connections"`
	IdleConnections   int           `json:"idle_connections"`
	WaitCount         int64         `json:"wait_count"`
	WaitDuration      time.Duration `json:"wait_duration"`
	MaxIdleClosed     int64         `json:"max_idle_closed"`
	MaxIdleTimeClosed int64         `json:"max_idle_time_closed"`
	MaxLifetimeClosed int64         `json:"max_lifetime_closed"`
}

// StatsProvider interface for getting connection statistics
type StatsProvider interface {
	Stats() ConnectionStats
}

// GormClient extends DatabaseClient with GORM-specific operations
type GormClient interface {
	DatabaseClient
	StatsProvider

	// GORM-specific methods
	GetDB() *gorm.DB // Returns *gorm.DB
}

// PgxClient extends DatabaseClient with pgx-specific operations
type PgxClient interface {
	DatabaseClient
	StatsProvider

	// pgx-specific methods
	GetPool() *pgxpool.Pool // Returns *pgxpool.Pool
	GetConn() *pgx.Conn     // Returns *pgx.Conn
	InsertModel(ctx context.Context, model any) error
	UpsertModel(ctx context.Context, model any, primaryKeys ...string) error
	BatchInsertModel(ctx context.Context, models []any, batchSize int) error
}
