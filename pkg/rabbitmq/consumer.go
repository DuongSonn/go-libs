package _rabbitmq

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

// ConsumeConfig holds configuration for consuming messages
type ConsumeConfig struct {
	Queue         string
	ConsumerTag   string
	AutoAck       bool
	Exclusive     bool
	NoLocal       bool
	NoWait        bool
	PrefetchCount int
	PrefetchSize  int
	Global        bool
}

// DefaultConsumeConfig returns default consume configuration
func DefaultConsumeConfig() ConsumeConfig {
	return ConsumeConfig{
		Queue:         "",
		ConsumerTag:   "",
		AutoAck:       false,
		Exclusive:     false,
		NoLocal:       false,
		NoWait:        false,
		PrefetchCount: 1,
		PrefetchSize:  0,
		Global:        false,
	}
}

// Message represents a consumed message
type Message struct {
	Body             []byte
	Headers          map[string]interface{}
	DeliveryTag      uint64
	MessageID        string
	RoutingKey       string
	Exchange         string
	RedeliveredCount int
	Timestamp        time.Time
	ContentType      string
	delivery         amqp.Delivery
}

// Ack acknowledges the message
func (m *Message) Ack() error {
	return m.delivery.Ack(false)
}

// Nack negatively acknowledges the message
func (m *Message) Nack(requeue bool) error {
	return m.delivery.Nack(false, requeue)
}

// Reject rejects the message
func (m *Message) Reject(requeue bool) error {
	return m.delivery.Reject(requeue)
}

// MessageProcessor is an interface for processing messages
type MessageProcessor interface {
	Process(ctx context.Context, msg *Message) error
}

// Consumer handles consuming messages from RabbitMQ
type Consumer struct {
	conn      *Connection
	config    ConsumeConfig
	processor MessageProcessor

	// Control channels
	stopCh chan struct{}
	doneCh chan struct{}

	// State
	consuming bool
	mu        sync.RWMutex
}

// NewConsumer creates a new RabbitMQ consumer
func NewConsumer(conn *Connection, config ConsumeConfig, processor MessageProcessor) *Consumer {
	return &Consumer{
		conn:      conn,
		config:    config,
		processor: processor,
		stopCh:    make(chan struct{}),
		doneCh:    make(chan struct{}),
		consuming: false,
	}
}

// Start begins consuming messages
func (c *Consumer) Start(ctx context.Context) error {
	c.mu.Lock()
	if c.consuming {
		c.mu.Unlock()
		return fmt.Errorf("consumer already started")
	}
	c.consuming = true
	c.mu.Unlock()

	// Start consuming in a goroutine
	go c.consume(ctx)

	return nil
}

// Stop stops consuming messages
func (c *Consumer) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.consuming {
		return nil
	}

	// Signal consume loop to stop
	close(c.stopCh)

	// Wait for consume loop to finish
	<-c.doneCh

	c.consuming = false

	// Reset channels for potential restart
	c.stopCh = make(chan struct{})
	c.doneCh = make(chan struct{})

	return nil
}

// IsConsuming checks if the consumer is active
func (c *Consumer) IsConsuming() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.consuming
}

// consume is the main consume loop
func (c *Consumer) consume(ctx context.Context) {
	defer close(c.doneCh)

	for {
		// Check if we should stop
		select {
		case <-c.stopCh:
			return
		default:
			// Continue consuming
		}

		// Check if connection is available
		if !c.conn.IsConnected() {
			time.Sleep(1 * time.Second)
			continue
		}

		// Get channel
		channel, err := c.conn.GetChannel()
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		// Set QoS
		if err := channel.Qos(
			c.config.PrefetchCount,
			c.config.PrefetchSize,
			c.config.Global,
		); err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		// Start consuming
		deliveries, err := channel.Consume(
			c.config.Queue,
			c.config.ConsumerTag,
			c.config.AutoAck,
			c.config.Exclusive,
			c.config.NoLocal,
			c.config.NoWait,
			nil, // arguments
		)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		// Process messages
		for {
			select {
			case <-c.stopCh:
				return
			case delivery, ok := <-deliveries:
				if !ok {
					// Channel closed, try to reconnect
					break
				}

				// Process the message
				c.processDelivery(ctx, delivery)
			}
		}
	}
}

// processDelivery handles a single delivery
func (c *Consumer) processDelivery(ctx context.Context, delivery amqp.Delivery) {
	// Create message from delivery
	msg := &Message{
		Body:        delivery.Body,
		Headers:     make(map[string]interface{}),
		DeliveryTag: delivery.DeliveryTag,
		MessageID:   delivery.MessageId,
		RoutingKey:  delivery.RoutingKey,
		Exchange:    delivery.Exchange,
		Timestamp:   delivery.Timestamp,
		ContentType: delivery.ContentType,
		delivery:    delivery,
	}

	// Copy headers
	for k, v := range delivery.Headers {
		msg.Headers[k] = v
	}

	// Get redelivered count
	if delivery.Redelivered {
		if count, ok := delivery.Headers["x-redelivered-count"]; ok {
			if countInt, ok := count.(int); ok {
				msg.RedeliveredCount = countInt
			}
		} else {
			msg.RedeliveredCount = 1
		}
	}

	// Process the message
	if err := c.processor.Process(ctx, msg); err != nil {
		// If processing fails, nack the message
		if nackErr := msg.Nack(true); nackErr != nil {
			// Log error
			fmt.Printf("Failed to nack message: %v\n", nackErr)
		}
	} else if !c.config.AutoAck {
		// If processing succeeds and not auto-ack, ack the message
		if ackErr := msg.Ack(); ackErr != nil {
			// Log error
			fmt.Printf("Failed to ack message: %v\n", ackErr)
		}
	}
}

// BatchConsumer consumes messages in batches
type BatchConsumer struct {
	Consumer
	batchSize int
	batchCh   chan []*Message
}

// NewBatchConsumer creates a new batch consumer
func NewBatchConsumer(conn *Connection, config ConsumeConfig, batchSize int) *BatchConsumer {
	return &BatchConsumer{
		Consumer: Consumer{
			conn:      conn,
			config:    config,
			stopCh:    make(chan struct{}),
			doneCh:    make(chan struct{}),
			consuming: false,
		},
		batchSize: batchSize,
		batchCh:   make(chan []*Message, 1),
	}
}

// GetBatch waits for a batch of messages
func (bc *BatchConsumer) GetBatch(ctx context.Context) ([]*Message, error) {
	select {
	case batch := <-bc.batchCh:
		return batch, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// consume is the main consume loop for batch consumer
func (bc *BatchConsumer) consume(ctx context.Context) {
	defer close(bc.doneCh)

	batch := make([]*Message, 0, bc.batchSize)

	for {
		// Check if we should stop
		select {
		case <-bc.stopCh:
			return
		default:
			// Continue consuming
		}

		// Check if connection is available
		if !bc.conn.IsConnected() {
			time.Sleep(1 * time.Second)
			continue
		}

		// Get channel
		channel, err := bc.conn.GetChannel()
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		// Set QoS
		if err := channel.Qos(
			bc.config.PrefetchCount,
			bc.config.PrefetchSize,
			bc.config.Global,
		); err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		// Start consuming
		deliveries, err := channel.Consume(
			bc.config.Queue,
			bc.config.ConsumerTag,
			bc.config.AutoAck,
			bc.config.Exclusive,
			bc.config.NoLocal,
			bc.config.NoWait,
			nil, // arguments
		)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		// Process messages
		for {
			select {
			case <-bc.stopCh:
				return
			case delivery, ok := <-deliveries:
				if !ok {
					// Channel closed, try to reconnect
					break
				}

				// Create message from delivery
				msg := &Message{
					Body:        delivery.Body,
					Headers:     make(map[string]interface{}),
					DeliveryTag: delivery.DeliveryTag,
					MessageID:   delivery.MessageId,
					RoutingKey:  delivery.RoutingKey,
					Exchange:    delivery.Exchange,
					Timestamp:   delivery.Timestamp,
					ContentType: delivery.ContentType,
					delivery:    delivery,
				}

				// Copy headers
				for k, v := range delivery.Headers {
					msg.Headers[k] = v
				}

				// Add to batch
				batch = append(batch, msg)

				// If batch is full, send it
				if len(batch) >= bc.batchSize {
					bc.batchCh <- batch
					batch = make([]*Message, 0, bc.batchSize)
				}
			}
		}
	}
}
