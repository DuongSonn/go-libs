# Go-Libs

A collection of reusable Go libraries for common backend development tasks.

## Overview

Go-Libs provides a set of well-tested, production-ready packages that can be imported into your Go projects to accelerate development. Each package is designed to be lightweight, performant, and easy to integrate.

## Packages

### Postgres

Database connection and operation utilities for PostgreSQL with two implementation options:

-   **GORM**: Object-relational mapping based implementation
-   **PGX**: Pure SQL implementation with pgx driver

[Learn more](pkg/postgres/README.md)

### Redis

Redis connection utilities with support for multiple deployment models:

-   **Single Node**: Simple Redis server connection
-   **Cluster**: Support for Redis Cluster with multiple nodes
-   **Sentinel**: High availability with Redis Sentinel
-   **Master-Slave**: Read-write separation for both Cluster and Sentinel

[Learn more](pkg/redis/README.md)

### Kafka

Flexible Kafka client for consuming messages with custom processors:

-   Topic-based message processing
-   Automatic partition management
-   Per-partition goroutines for parallel processing

[Learn more](pkg/kafka/README.md)

### Errors

Multilingual error handling system with:

-   Support for Vietnamese and English error messages
-   HTTP status code mapping
-   Parameterized error messages
-   Module-based error categorization

[Learn more](pkg/errors/README.md)

### Logger

Structured logging with:

-   Multiple output formats (JSON, text)
-   Log level filtering
-   Context-aware logging

[Learn more](pkg/logger/README.md)

## Installation

```bash
# Install the entire library
go get github.com/DuongSonn/go-libs

# Or install specific packages
go get github.com/DuongSonn/go-libs/pkg/postgres
go get github.com/DuongSonn/go-libs/pkg/redis
go get github.com/DuongSonn/go-libs/pkg/kafka
go get github.com/DuongSonn/go-libs/pkg/errors
go get github.com/DuongSonn/go-libs/pkg/logger
```

## Requirements

-   Go 1.24 or higher

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
