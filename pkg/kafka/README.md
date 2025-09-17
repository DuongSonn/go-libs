# Kafka Package

A flexible Kafka client library built on top of [franz-go](https://github.com/twmb/franz-go), providing a simple interface for consuming messages from Kafka topics with custom message processors.

## Features

-   Topic-based message processing with custom service implementations
-   Automatic partition management and distribution
-   Per-partition goroutines for parallel processing
-   Clean connection lifecycle management
-   Manual offset commit control

## Installation

```bash
go get github.com/twmb/franz-go
```

## Usage

### Basic Usage

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"

    kafka "_kafka"  // Import the kafka package
)

// Define a message processor that implements the IMessageProcessor interface
type MyMessageProcessor struct {
    // Add any fields you need
}

// Implement the Process method required by the IMessageProcessor interface
func (p *MyMessageProcessor) Process(ctx context.Context, msg *kgo.Record) error {
    log.Printf("Received message: Topic=%s, Partition=%d, Offset=%d, Key=%s, Value=%s",
        msg.Topic, msg.Partition, msg.Offset, string(msg.Key), string(msg.Value))
    // Process the message...
    return nil
}

func main() {
    // Create a configuration
    cfg := kafka.Config{
        Brokers: []string{"localhost:9092"},
        Group:   "my-consumer-group",
        Topics:  []string{"my-topic", "another-topic"},
    }

    // Create a new connection
    conn := kafka.NewConnection(cfg)

    // Register message processors for each topic
    conn.RegisterService("my-topic", &MyMessageProcessor{})
    conn.RegisterService("another-topic", &MyMessageProcessor{})

    // Connect to Kafka
    if err := conn.Connect(context.Background()); err != nil {
        log.Fatalf("Failed to connect to Kafka: %v", err)
    }
    defer conn.Close()

    // Wait for termination signal
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    <-sigChan

    log.Println("Shutting down...")
}
```

### Advanced Usage

#### Different Processors for Different Topics

```go
// Order processor
type OrderProcessor struct {}

func (p *OrderProcessor) Process(ctx context.Context, msg *kgo.Record) error {
    // Process order messages
    return nil
}

// Payment processor
type PaymentProcessor struct {}

func (p *PaymentProcessor) Process(ctx context.Context, msg *kgo.Record) error {
    // Process payment messages
    return nil
}

// In your main function
conn := kafka.NewConnection(cfg)
conn.RegisterService("orders", &OrderProcessor{})
conn.RegisterService("payments", &PaymentProcessor{})
```

## Configuration Options

The `Config` struct provides the following options:

```go
type Config struct {
    Brokers []string // Kafka broker addresses
    Group   string   // Consumer group ID
    Topics  []string // List of topics to consume
}
```

## Interface

To process messages, implement the `IMessageProcessor` interface:

```go
type IMessageProcessor interface {
    Process(ctx context.Context, msg *kgo.Record) error
}
```

## Error Handling

Errors during message processing are logged but don't stop the consumer. The message will be committed even if processing fails, so implement your own retry logic if needed.

## Graceful Shutdown

Call the `Close()` method to gracefully shut down all Kafka connections:

```go
conn.Close()
```

## Thread Safety

The `RegisterService` method is not thread-safe and should only be called before `Connect()`.

## License

This library is distributed under the same license as the go-libs project.
