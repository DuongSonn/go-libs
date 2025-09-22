# PostgreSQL Package

This package provides utilities for connecting to and interacting with PostgreSQL databases in Go. It supports two main connection approaches: using GORM and using pure pgx.

## Key Features

-   **Single Connection**: Connect to a single PostgreSQL instance
-   **Master-Slave**: Support for master-slave replication with automatic failover
-   **GORM Support**: Integration with GORM ORM for object-oriented data manipulation
-   **pgx Support**: Pure pgx driver implementation for optimal performance
-   **Error Handling and Auto-reconnection**: Automatic handling of connection errors and retry attempts
-   **Flexible Configuration**: Multiple configuration options to optimize connections

## Installation

```bash
go get -u go-libs/pkg/postgres
```

## Package Structure

```
pkg/postgres/
├── config.go           # PostgreSQL connection configuration
├── interface.go        # Common interface definitions
├── gorm/               # GORM implementation
│   ├── connection.go   # GORM connection
│   ├── client.go       # GORM client
│   ├── transaction.go  # Transaction handling with GORM
│   ├── rows.go         # Query result handling with GORM
│   └── master_slave.go # Master-slave support with GORM
└── pgx/                # pgx implementation
    ├── connection.go   # pgx connection
    ├── client.go       # pgx client
    ├── transaction.go  # Transaction handling with pgx
    ├── rows.go         # Query result handling with pgx
    └── master_slave.go # Master-slave support with pgx
```

## Main Interfaces

This package defines the following key interfaces:

-   `DatabaseClient`: Base interface for all database clients
-   `MasterSlaveClient`: Interface for master-slave connections
-   `Transaction`: Interface for database transactions
-   `Rows` and `Row`: Interfaces for query results
-   `GormClient` and `PgxClient`: Specific interfaces for GORM and pgx
-   `GormMasterSlaveClient` and `PgxMasterSlaveClient`: Master-slave interfaces for GORM and pgx

## Usage Guide

### Single Connection with pgx

```go
package main

import (
	"context"
	"fmt"
	"log"

	_postgres "go-libs/pkg/postgres"
	_pgx_postgres "go-libs/pkg/postgres/pgx"
)

func main() {
	// Create configuration
	config := _postgres.DefaultConfig()
	config.Host = "localhost"
	config.Port = 5432
	config.User = "postgres"
	config.Password = "password"
	config.Database = "mydb"
	config.SSLMode = "disable"

	// Create connection
	conn := _pgx_postgres.NewConnection(config)

	// Connect to database
	ctx := context.Background()
	if err := conn.Connect(ctx); err != nil {
		log.Fatalf("Could not connect to PostgreSQL: %v", err)
	}
	defer conn.Close()

	// Check connection
	if conn.IsHealthy(ctx) {
		fmt.Println("Successfully connected to PostgreSQL")
	}

	// Execute query
	rows, err := conn.GetPool().Query(ctx, "SELECT id, name FROM users LIMIT 10")
	if err != nil {
		log.Fatalf("Query error: %v", err)
	}
	defer rows.Close()

	// Process results
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			log.Fatalf("Scan error: %v", err)
		}
		fmt.Printf("ID: %d, Name: %s\n", id, name)
	}
}
```

### Single Connection with GORM

```go
package main

import (
	"context"
	"fmt"
	"log"

	_postgres "go-libs/pkg/postgres"
	_gorm_postgres "go-libs/pkg/postgres/gorm"
)

func main() {
	// Create GORM configuration
	config := _postgres.DefaultGormConfig()
	config.Host = "localhost"
	config.Port = 5432
	config.User = "postgres"
	config.Password = "password"
	config.Database = "mydb"
	config.SSLMode = "disable"

	// Create connection
	conn := _gorm_postgres.NewConnection(config)

	// Connect to database
	ctx := context.Background()
	if err := conn.Connect(ctx); err != nil {
		log.Fatalf("Could not connect to PostgreSQL: %v", err)
	}
	defer conn.Close()

	// Check connection
	if conn.IsHealthy(ctx) {
		fmt.Println("Successfully connected to PostgreSQL with GORM")
	}

	// Define model
	type User struct {
		ID   uint   `gorm:"primaryKey"`
		Name string `gorm:"size:255"`
	}

	// Auto migrate
	if err := conn.GetDB().AutoMigrate(&User{}); err != nil {
		log.Fatalf("Could not auto migrate: %v", err)
	}

	// Create new user
	user := User{Name: "John Doe"}
	if err := conn.GetDB().Create(&user).Error; err != nil {
		log.Fatalf("Could not create user: %v", err)
	}
	fmt.Printf("Created user with ID: %d\n", user.ID)

	// Query users
	var users []User
	if err := conn.GetDB().Find(&users).Error; err != nil {
		log.Fatalf("Could not query users: %v", err)
	}

	for _, u := range users {
		fmt.Printf("User: ID=%d, Name=%s\n", u.ID, u.Name)
	}
}
```

### Master-Slave Connection with pgx

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	_postgres "go-libs/pkg/postgres"
	_pgx_postgres "go-libs/pkg/postgres/pgx"
)

func main() {
	// Create master configuration
	masterConfig := _postgres.DefaultConfig()
	masterConfig.Host = "master.postgres.example.com" // Change to actual address
	masterConfig.Port = 5432
	masterConfig.User = "postgres"
	masterConfig.Password = "masterpassword"
	masterConfig.Database = "mydb"

	// Create slave configuration
	slaveConfig := _postgres.DefaultConfig()
	slaveConfig.Host = "slave.postgres.example.com" // Change to actual address
	slaveConfig.Port = 5432
	slaveConfig.User = "postgres"
	slaveConfig.Password = "slavepassword"
	slaveConfig.Database = "mydb"

	// Create master-slave configuration
	msConfig := _postgres.DefaultMasterSlaveConfig()
	msConfig.Master = masterConfig
	msConfig.Slave = slaveConfig
	msConfig.UseSlaveConnection = true
	msConfig.SlaveReadOnly = true
	msConfig.AutoFailover = true
	msConfig.HealthCheckEnabled = true
	msConfig.HealthCheckInterval = 30 * time.Second

	// Create master-slave connection
	conn := _pgx_postgres.NewMasterSlaveConnection(msConfig)

	// Connect to database
	ctx := context.Background()
	if err := conn.Connect(ctx); err != nil {
		log.Fatalf("Could not connect to PostgreSQL: %v", err)
	}
	defer conn.Close()

	// Check connection
	if conn.IsHealthy(ctx) {
		fmt.Println("Successfully connected to PostgreSQL master-slave")
	}

	// Get master client and use for write operations
	masterClient := conn.GetMasterClient()

	// Example: Execute INSERT query on master
	tx, err := masterClient.BeginTx(ctx)
	if err != nil {
		log.Fatalf("Could not begin transaction: %v", err)
	}

	err = tx.Exec(ctx, "INSERT INTO users (name, email) VALUES ($1, $2)", "John Doe", "john@example.com")
	if err != nil {
		tx.Rollback()
		log.Fatalf("Could not execute INSERT: %v", err)
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("Could not commit transaction: %v", err)
	}
	fmt.Println("Successfully executed INSERT on master")

	// Check if slave connection is available
	if conn.HasSlaveConnected() {
		// Get slave client and use for read operations
		slaveClient := conn.GetSlaveClient()

		// Example: Execute SELECT query on slave
		rows, err := slaveClient.(*_pgx_postgres.Connection).GetPool().Query(ctx, "SELECT id, name, email FROM users ORDER BY id DESC LIMIT 1")
		if err != nil {
			log.Printf("Could not execute SELECT on slave: %v", err)

			// Try reading from master if reading from slave fails
			rows, err = conn.GetMasterPool().Query(ctx, "SELECT id, name, email FROM users ORDER BY id DESC LIMIT 1")
			if err != nil {
				log.Fatalf("Could not execute SELECT on master: %v", err)
			}
			defer rows.Close()

			fmt.Println("Successfully executed SELECT on master (fallback)")
		} else {
			defer rows.Close()
			fmt.Println("Successfully executed SELECT on slave")
		}

		// Process results
		for rows.Next() {
			var id int
			var name, email string
			if err := rows.Scan(&id, &name, &email); err != nil {
				log.Fatalf("Could not scan result: %v", err)
			}
			fmt.Printf("Result: ID=%d, Name=%s, Email=%s\n", id, name, email)
		}
	}
}
```

### Master-Slave Connection with GORM

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	_postgres "go-libs/pkg/postgres"
	_gorm_postgres "go-libs/pkg/postgres/gorm"
)

func main() {
	// Create master configuration
	masterConfig := _postgres.DefaultConfig()
	masterConfig.Host = "master.postgres.example.com" // Change to actual address
	masterConfig.Port = 5432
	masterConfig.User = "postgres"
	masterConfig.Password = "masterpassword"
	masterConfig.Database = "mydb"

	// Create slave configuration
	slaveConfig := _postgres.DefaultConfig()
	slaveConfig.Host = "slave.postgres.example.com" // Change to actual address
	slaveConfig.Port = 5432
	slaveConfig.User = "postgres"
	slaveConfig.Password = "slavepassword"
	slaveConfig.Database = "mydb"

	// Create master-slave configuration for GORM
	msConfig := _postgres.DefaultGormMasterSlaveConfig()
	msConfig.Master = masterConfig
	msConfig.Slave = slaveConfig
	msConfig.UseSlaveConnection = true
	msConfig.SlaveReadOnly = true
	msConfig.AutoFailover = true
	msConfig.HealthCheckEnabled = true
	msConfig.HealthCheckInterval = 30 * time.Second

	// GORM configuration for master
	msConfig.MasterSkipDefaultTransaction = false
	msConfig.MasterPrepareStmt = true

	// GORM configuration for slave
	msConfig.SlaveSkipDefaultTransaction = true // Typically true for read-only connections
	msConfig.SlavePrepareStmt = true

	// Create master-slave connection
	conn := _gorm_postgres.NewMasterSlaveConnection(msConfig)

	// Connect to database
	ctx := context.Background()
	if err := conn.Connect(ctx); err != nil {
		log.Fatalf("Could not connect to PostgreSQL: %v", err)
	}
	defer conn.Close()

	// Check connection
	if conn.IsHealthy(ctx) {
		fmt.Println("Successfully connected to PostgreSQL master-slave with GORM")
	}

	// Define model
	type User struct {
		ID    uint   `gorm:"primaryKey"`
		Name  string `gorm:"size:255"`
		Email string `gorm:"size:255;uniqueIndex"`
	}

	// Get master DB and use for write operations
	masterDB := conn.GetMasterDB()

	// Auto migrate schema
	if err := masterDB.AutoMigrate(&User{}); err != nil {
		log.Fatalf("Could not auto migrate: %v", err)
	}

	// Example: Add new user
	user := User{
		Name:  "Jane Smith",
		Email: "jane@example.com",
	}

	if err := masterDB.Create(&user).Error; err != nil {
		log.Fatalf("Could not create user: %v", err)
	}
	fmt.Println("Successfully created user on master")

	// Check if slave connection is available
	if conn.HasSlaveConnected() {
		// Get slave DB and use for read operations
		slaveDB := conn.GetSlaveDB()

		// Example: Read users from slave
		var users []User
		if err := slaveDB.Limit(5).Order("id desc").Find(&users).Error; err != nil {
			log.Printf("Could not read users from slave: %v", err)

			// Try reading from master if reading from slave fails
			if err := masterDB.Limit(5).Order("id desc").Find(&users).Error; err != nil {
				log.Fatalf("Could not read users from master: %v", err)
			}
			fmt.Println("Successfully read users from master (fallback)")
		} else {
			fmt.Println("Successfully read users from slave")
		}

		// Display results
		for _, u := range users {
			fmt.Printf("User: ID=%d, Name=%s, Email=%s\n", u.ID, u.Name, u.Email)
		}
	}
}
```

## Error Handling and Failover

This package provides automatic error handling and failover mechanisms:

### Connection Error Handling

-   **Automatic Retry**: When a connection fails, the package automatically retries according to the `MaxRetries` and `RetryInterval` configuration
-   **Timeouts**: `ConnectTimeout` and `QueryTimeout` configurations help prevent hanging queries

### Failover in Master-Slave

-   **Health Checking**: Automatically checks the health of connections according to the `HealthCheckInterval`
-   **Automatic Failover**: When the master fails, the slave can be used for read operations
-   **Automatic Reconnection**: Attempts to reconnect to master/slave when errors are detected

### Failover Configuration

```go
msConfig := _postgres.DefaultMasterSlaveConfig()
msConfig.AutoFailover = true                 // Enable automatic failover
msConfig.FailoverRetries = 3                 // Number of reconnection attempts
msConfig.FailoverInterval = 5 * time.Second  // Interval between attempts
msConfig.HealthCheckEnabled = true           // Enable health checking
msConfig.HealthCheckInterval = 30 * time.Second // Health check interval
```

## Performance Optimization

### Connection Pool Configuration

```go
config := _postgres.DefaultConfig()
config.MaxOpenConns = 50      // Maximum number of open connections
config.MaxIdleConns = 10      // Maximum number of idle connections
config.ConnMaxLifetime = 5 * time.Minute // Maximum connection lifetime
config.ConnMaxIdleTime = 5 * time.Minute // Maximum idle connection time
```

### GORM Configuration

```go
gormConfig := _postgres.DefaultGormConfig()
gormConfig.PrepareStmt = true                // Prepare statements
gormConfig.SkipDefaultTransaction = true     // Skip default transaction
gormConfig.SlowThreshold = 200 * time.Millisecond // Slow query threshold
```

## Contributing

Contributions are welcome! Please feel free to submit an issue or pull request to the repository.
