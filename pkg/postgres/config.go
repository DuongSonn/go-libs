package _postgres

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

// MasterSlaveConfig holds configuration for master-slave setup
type MasterSlaveConfig struct {
	Master *Config `json:"master" yaml:"master"`
	Slave  *Config `json:"slave" yaml:"slave"`

	// Master-slave specific settings
	UseSlaveConnection bool `json:"use_slave_connection" yaml:"use_slave_connection"`
	SlaveReadOnly      bool `json:"slave_read_only" yaml:"slave_read_only"`

	// Failover settings
	AutoFailover        bool          `json:"auto_failover" yaml:"auto_failover"`
	FailoverRetries     int           `json:"failover_retries" yaml:"failover_retries"`
	FailoverInterval    time.Duration `json:"failover_interval" yaml:"failover_interval"`
	HealthCheckEnabled  bool          `json:"health_check_enabled" yaml:"health_check_enabled"`
	HealthCheckInterval time.Duration `json:"health_check_interval" yaml:"health_check_interval"`
}

// DefaultMasterSlaveConfig returns a master-slave configuration with sensible defaults
func DefaultMasterSlaveConfig() *MasterSlaveConfig {
	return &MasterSlaveConfig{
		Master:              DefaultConfig(),
		Slave:               DefaultConfig(),
		UseSlaveConnection:  true,
		SlaveReadOnly:       true,
		AutoFailover:        true,
		FailoverRetries:     3,
		FailoverInterval:    5 * time.Second,
		HealthCheckEnabled:  true,
		HealthCheckInterval: 30 * time.Second,
	}
}

// Validate checks if the master-slave configuration is valid
func (c *MasterSlaveConfig) Validate() error {
	if c.Master == nil {
		return fmt.Errorf("master configuration is required")
	}

	if err := c.Master.Validate(); err != nil {
		return fmt.Errorf("invalid master configuration: %w", err)
	}

	if c.UseSlaveConnection {
		if c.Slave == nil {
			return fmt.Errorf("slave configuration is required when use_slave_connection is true")
		}

		if err := c.Slave.Validate(); err != nil {
			return fmt.Errorf("invalid slave configuration: %w", err)
		}
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

// GormMasterSlaveConfig holds GORM-specific configuration for master-slave setup
type GormMasterSlaveConfig struct {
	*MasterSlaveConfig

	// GORM-specific settings for master
	MasterLogLevel                                 int           `json:"master_log_level" yaml:"master_log_level"`
	MasterSlowThreshold                            time.Duration `json:"master_slow_threshold" yaml:"master_slow_threshold"`
	MasterSkipDefaultTransaction                   bool          `json:"master_skip_default_transaction" yaml:"master_skip_default_transaction"`
	MasterPrepareStmt                              bool          `json:"master_prepare_stmt" yaml:"master_prepare_stmt"`
	MasterDisableForeignKeyConstraintWhenMigrating bool          `json:"master_disable_foreign_key_constraint_when_migrating" yaml:"master_disable_foreign_key_constraint_when_migrating"`

	// GORM-specific settings for slave
	SlaveLogLevel                                 int           `json:"slave_log_level" yaml:"slave_log_level"`
	SlaveSlowThreshold                            time.Duration `json:"slave_slow_threshold" yaml:"slave_slow_threshold"`
	SlaveSkipDefaultTransaction                   bool          `json:"slave_skip_default_transaction" yaml:"slave_skip_default_transaction"`
	SlavePrepareStmt                              bool          `json:"slave_prepare_stmt" yaml:"slave_prepare_stmt"`
	SlaveDisableForeignKeyConstraintWhenMigrating bool          `json:"slave_disable_foreign_key_constraint_when_migrating" yaml:"slave_disable_foreign_key_constraint_when_migrating"`
}

// DefaultGormMasterSlaveConfig returns GORM master-slave configuration with sensible defaults
func DefaultGormMasterSlaveConfig() *GormMasterSlaveConfig {
	return &GormMasterSlaveConfig{
		MasterSlaveConfig:                              DefaultMasterSlaveConfig(),
		MasterLogLevel:                                 1, // Silent
		MasterSlowThreshold:                            200 * time.Millisecond,
		MasterSkipDefaultTransaction:                   false,
		MasterPrepareStmt:                              true,
		MasterDisableForeignKeyConstraintWhenMigrating: false,
		SlaveLogLevel:                                  1, // Silent
		SlaveSlowThreshold:                             200 * time.Millisecond,
		SlaveSkipDefaultTransaction:                    true, // Typically true for read-only connections
		SlavePrepareStmt:                               true,
		SlaveDisableForeignKeyConstraintWhenMigrating:  false,
	}
}

// GetMasterGormConfig returns GORM configuration for the master
func (c *GormMasterSlaveConfig) GetMasterGormConfig() *GormConfig {
	return &GormConfig{
		Config:                                   c.Master,
		LogLevel:                                 c.MasterLogLevel,
		SlowThreshold:                            c.MasterSlowThreshold,
		SkipDefaultTransaction:                   c.MasterSkipDefaultTransaction,
		PrepareStmt:                              c.MasterPrepareStmt,
		DisableForeignKeyConstraintWhenMigrating: c.MasterDisableForeignKeyConstraintWhenMigrating,
	}
}

// GetSlaveGormConfig returns GORM configuration for the slave
func (c *GormMasterSlaveConfig) GetSlaveGormConfig() *GormConfig {
	return &GormConfig{
		Config:                                   c.Slave,
		LogLevel:                                 c.SlaveLogLevel,
		SlowThreshold:                            c.SlaveSlowThreshold,
		SkipDefaultTransaction:                   c.SlaveSkipDefaultTransaction,
		PrepareStmt:                              c.SlavePrepareStmt,
		DisableForeignKeyConstraintWhenMigrating: c.SlaveDisableForeignKeyConstraintWhenMigrating,
	}
}
