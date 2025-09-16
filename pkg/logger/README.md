# Logger Package

A Go logging package built on top of Go's standard `slog` library with additional features.

## Features

-   Built on Go's standard `slog` library
-   Always shows source code location for easy debugging
-   Custom time format (MM/DD/YYYY HH:mm:ss)
-   Masks sensitive fields with asterisks (`***`)
-   Supports both JSON and text output formats
-   Configurable output destination (stdout, stderr, or file)
-   Context-aware logging

## Installation

```bash
go get -u go-libs/pkg/logger
```

## Quick Start

```go
package main

import (
    "context"
    "go-libs/pkg/logger"
)

func main() {
    ctx := context.Background()

    // Use the default logger
    logger.Info(ctx, "Application started")

    // Log with additional fields
    logger.Info(ctx, "User logged in", "userId", "12345", "username", "john")

    // Log errors
    err := someFunction()
    if err != nil {
        logger.Error(ctx, "Operation failed", "error", err)
    }
}
```

## Configuration

### Basic Configuration

```go
// Create a custom logger
customLogger := logger.New(logger.Config{
    Level:      "debug",    // debug, info, warn, error
    Output:     "stdout",   // stdout, stderr, or file path
    Format:     "json",     // json or text
    HideFields: []string{"password", "token"}, // Fields to mask with ***
})

// Use the custom logger
customLogger.Info(ctx, "Custom logger message")
```

### Changing the Default Logger

```go
// Create a new logger
newDefaultLogger := logger.New(logger.Config{
    Level:  "debug",
    Output: "logs.txt", // Write to file
    Format: "text",
})

// Set as the default logger
logger.SetDefault(newDefaultLogger)

// Now all global logger calls will use this configuration
logger.Debug(ctx, "This will be written to logs.txt")
```

## Masking Sensitive Data

The logger automatically masks sensitive fields specified in the `HideFields` configuration:

```go
// Configure which fields should be masked
cfg := logger.Config{
    Level:      "info",
    Output:     "stdout",
    Format:     "json",
    HideFields: []string{"password", "token", "creditCard"},
}
customLogger := logger.New(cfg)

// When these fields are logged, their values will be masked with ***
customLogger.Info(ctx, "Payment processed",
    "userId", "12345",
    "amount", 100.50,
    "creditCard", "4111-1111-1111-1111") // Will be masked as "***"
```

Output:

```json
{
    "time": "09/16/2025 12:34:32",
    "level": "INFO",
    "msg": "Payment processed",
    "userId": "12345",
    "amount": 100.5,
    "creditCard": "***"
}
```

## Log Levels

The logger supports the following log levels:

-   `debug`: Detailed information for debugging
-   `info`: General information about application progress
-   `warn`: Warning situations that might cause problems
-   `error`: Error events that might still allow the application to continue
-   `fatal`: Severe error events that cause the application to terminate

```go
logger.Debug(ctx, "Debug message")
logger.Info(ctx, "Info message")
logger.Warn(ctx, "Warning message")
logger.Error(ctx, "Error message")
logger.Fatal(ctx, "Fatal message") // This will exit the application
```

## Using with Context

All logging methods accept a context as their first parameter:

```go
// Create a context with request ID
ctx := context.WithValue(context.Background(), "requestId", "req-123")

// Log with this context
logger.Info(ctx, "Processing request")
```

## Advanced Usage

### Structured Logging

```go
// Log structured data
logger.Info(ctx, "Order created",
    "orderId", "ORD-12345",
    "customer", "John Doe",
    "items", 3,
    "total", 59.99)
```

### Error Logging

```go
err := errors.New("database connection failed")
logger.Error(ctx, "Failed to process request",
    "error", err,
    "retryCount", 3,
    "operation", "getUserProfile")
```

## Best Practices

1. **Use appropriate log levels**: Reserve `Error` for actual errors and `Debug` for detailed troubleshooting information.

2. **Include context**: Always pass the context to logging functions to maintain request tracing.

3. **Structured logging**: Use key-value pairs for structured logging rather than formatting strings.

4. **Mask sensitive data**: Add sensitive field names to the `HideFields` configuration.

5. **Be consistent**: Use the same logger configuration across your application for consistent logs.

## License

[Your License Here]
