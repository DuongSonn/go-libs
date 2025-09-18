# RabbitMQ Package

A flexible RabbitMQ client library providing a simple interface for publishing and consuming messages with RabbitMQ.

## Features

-   Connection management with automatic reconnection
-   Exchange and queue declaration
-   Message publishing with delivery confirmations
-   Message consuming with automatic acknowledgment
-   Batch message publishing and consuming
-   Transaction support for batch operations
-   Message ID tracking and correlation

## Installation

```bash
go get github.com/streadway/amqp
```

## Usage

### Basic Connection

```go
package main

import (
    "context"
    "log"

    rabbitmq "github.com/DuongSonn/go-libs/pkg/rabbitmq"
)

func main() {
    // Create configuration
    config := rabbitmq.DefaultConfig()
    config.Host = "localhost"
    config.Port = 5672
    config.Username = "guest"
    config.Password = "guest"

    // Create connection
    conn := rabbitmq.NewConnection(config)

    // Connect to RabbitMQ
    if err := conn.Connect(context.Background()); err != nil {
        log.Fatalf("Failed to connect to RabbitMQ: %v", err)
    }
    defer conn.Close()

    // Declare exchange
    exchangeConfig := rabbitmq.ExchangeConfig{
        Name:       "my-exchange",
        Type:       "direct",
        Durable:    true,
        AutoDelete: false,
        Internal:   false,
        NoWait:     false,
    }

    if err := conn.DeclareExchange(exchangeConfig); err != nil {
        log.Fatalf("Failed to declare exchange: %v", err)
    }

    // Declare queue
    queueConfig := rabbitmq.QueueConfig{
        Name:       "my-queue",
        Durable:    true,
        AutoDelete: false,
        Exclusive:  false,
        NoWait:     false,
    }

    queue, err := conn.DeclareQueue(queueConfig)
    if err != nil {
        log.Fatalf("Failed to declare queue: %v", err)
    }

    // Bind queue to exchange
    bindingConfig := rabbitmq.BindingConfig{
        Exchange:   "my-exchange",
        Queue:      queue.Name,
        RoutingKey: "my-routing-key",
        NoWait:     false,
    }

    if err := conn.BindQueue(bindingConfig); err != nil {
        log.Fatalf("Failed to bind queue: %v", err)
    }

    log.Println("RabbitMQ setup complete")
}
```

### Publishing Messages

```go
// Create producer
producer := rabbitmq.NewProducer(conn)

// Create publish config
publishConfig := rabbitmq.DefaultPublishConfig()
publishConfig.Exchange = "my-exchange"
publishConfig.RoutingKey = "my-routing-key"
publishConfig.DeliveryMode = 2 // persistent

// Publish message
result, err := producer.Publish(context.Background(), []byte("Hello, RabbitMQ!"), publishConfig)
if err != nil {
    log.Fatalf("Failed to publish message: %v", err)
}

log.Printf("Published message with ID: %s", result.MessageID)

// Publish JSON message
jsonData := []byte(`{"name":"John","age":30}`)
result, err = producer.PublishJSON(context.Background(), jsonData, publishConfig)
if err != nil {
    log.Fatalf("Failed to publish JSON message: %v", err)
}

// Publish batch of messages
messages := [][]byte{
    []byte("Message 1"),
    []byte("Message 2"),
    []byte("Message 3"),
}

results, err := producer.PublishBatch(context.Background(), messages, publishConfig)
if err != nil {
    log.Fatalf("Failed to publish batch: %v", err)
}

log.Printf("Published %d messages", len(results))
```

### Consuming Messages

```go
// Create message processor
type MyMessageProcessor struct {}

func (p *MyMessageProcessor) Process(ctx context.Context, msg *rabbitmq.Message) error {
    log.Printf("Received message: %s", string(msg.Body))
    return nil
}

// Create consumer config
consumeConfig := rabbitmq.DefaultConsumeConfig()
consumeConfig.Queue = "my-queue"
consumeConfig.AutoAck = false
consumeConfig.PrefetchCount = 10

// Create consumer
consumer := rabbitmq.NewConsumer(conn, consumeConfig, &MyMessageProcessor{})

// Start consuming
if err := consumer.Start(context.Background()); err != nil {
    log.Fatalf("Failed to start consumer: %v", err)
}

// To stop consuming
defer consumer.Stop()
```

### Batch Consuming

```go
// Create batch consumer
batchConsumer := rabbitmq.NewBatchConsumer(conn, consumeConfig, 10)

// Start consuming
if err := batchConsumer.Start(context.Background()); err != nil {
    log.Fatalf("Failed to start batch consumer: %v", err)
}

// Get batches of messages
for {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    batch, err := batchConsumer.GetBatch(ctx)
    cancel()

    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            continue
        }
        log.Fatalf("Error getting batch: %v", err)
    }

    log.Printf("Received batch of %d messages", len(batch))

    // Process batch
    for _, msg := range batch {
        // Process message
        log.Printf("Processing message: %s", string(msg.Body))

        // Acknowledge message
        if err := msg.Ack(); err != nil {
            log.Printf("Failed to ack message: %v", err)
        }
    }
}
```

## Configuration

### Connection Config

```go
type Config struct {
    Host         string        // RabbitMQ host
    Port         int           // RabbitMQ port
    Username     string        // RabbitMQ username
    Password     string        // RabbitMQ password
    VHost        string        // RabbitMQ virtual host
    ConnTimeout  time.Duration // Connection timeout
    RetryTimeout time.Duration // Retry timeout
    MaxRetries   int           // Maximum number of connection retries
}
```

### Exchange Config

```go
type ExchangeConfig struct {
    Name       string // Exchange name
    Type       string // Exchange type (direct, fanout, topic, headers)
    Durable    bool   // Survive broker restart
    AutoDelete bool   // Delete when no longer used
    Internal   bool   // Cannot be published to directly
    NoWait     bool   // Don't wait for confirmation
}
```

### Queue Config

```go
type QueueConfig struct {
    Name       string // Queue name
    Durable    bool   // Survive broker restart
    AutoDelete bool   // Delete when no longer used
    Exclusive  bool   // Only accessible by this connection
    NoWait     bool   // Don't wait for confirmation
}
```

## Error Handling

The package includes comprehensive error handling and automatic reconnection. When a connection is lost, the library will attempt to reconnect automatically with configurable retry parameters.

## Thread Safety

All components of this package are thread-safe and can be used concurrently from multiple goroutines.

## License

This library is distributed under the same license as the go-libs project.
