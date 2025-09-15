package config

import (
	"fmt"
	"time"
)

// Config holds PostgreSQL database configuration
type Config struct {
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	User     string `json:"user" yaml:"user"`
	Password string `json:"password" yaml:"password"`
	Database string `json:"database" yaml:"database"`
	SSLMode  string `json:"ssl_mode" yaml:"ssl_mode"`

	// Connection pool settings
	MaxOpenConns    int           `json:"max_open_conns" yaml:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns" yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime" yaml:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time" yaml:"conn_max_idle_time"`

	// Connection timeout settings
	ConnectTimeout time.Duration `json:"connect_timeout" yaml:"connect_timeout"`
	QueryTimeout   time.Duration `json:"query_timeout" yaml:"query_timeout"`

	// Retry settings
	MaxRetries    int           `json:"max_retries" yaml:"max_retries"`
	RetryInterval time.Duration `json:"retry_interval" yaml:"retry_interval"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "",
		Database: "postgres",
		SSLMode:  "disable",

		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,

		ConnectTimeout: 10 * time.Second,
		QueryTimeout:   30 * time.Second,

		MaxRetries:    3,
		RetryInterval: 1 * time.Second,
	}
}

// DSN returns the PostgreSQL data source name
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode,
	)
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("host is required")
	}
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	if c.User == "" {
		return fmt.Errorf("user is required")
	}
	if c.Database == "" {
		return fmt.Errorf("database is required")
	}
	if c.MaxOpenConns <= 0 {
		return fmt.Errorf("max_open_conns must be greater than 0")
	}
	if c.MaxIdleConns < 0 {
		return fmt.Errorf("max_idle_conns must be greater than or equal to 0")
	}
	if c.MaxIdleConns > c.MaxOpenConns {
		return fmt.Errorf("max_idle_conns cannot be greater than max_open_conns")
	}
	return nil
}

// GormConfig holds GORM-specific configuration
type GormConfig struct {
	*Config

	// GORM-specific settings
	LogLevel                                 int           `json:"log_level" yaml:"log_level"`
	SlowThreshold                            time.Duration `json:"slow_threshold" yaml:"slow_threshold"`
	SkipDefaultTransaction                   bool          `json:"skip_default_transaction" yaml:"skip_default_transaction"`
	PrepareStmt                              bool          `json:"prepare_stmt" yaml:"prepare_stmt"`
	DisableForeignKeyConstraintWhenMigrating bool          `json:"disable_foreign_key_constraint_when_migrating" yaml:"disable_foreign_key_constraint_when_migrating"`
}

// DefaultGormConfig returns GORM configuration with sensible defaults
func DefaultGormConfig() *GormConfig {
	return &GormConfig{
		Config:                                   DefaultConfig(),
		LogLevel:                                 1, // Silent
		SlowThreshold:                            200 * time.Millisecond,
		SkipDefaultTransaction:                   false,
		PrepareStmt:                              true,
		DisableForeignKeyConstraintWhenMigrating: false,
	}
}
