package _rabbitmq

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/streadway/amqp"
)

// PublishConfig holds configuration for publishing messages
type PublishConfig struct {
	Exchange     string
	RoutingKey   string
	Mandatory    bool
	Immediate    bool
	ContentType  string
	DeliveryMode uint8 // 1 = non-persistent, 2 = persistent
	Priority     uint8
	Expiration   string // expiration time in milliseconds as string
}

// DefaultPublishConfig returns default publish configuration
func DefaultPublishConfig() PublishConfig {
	return PublishConfig{
		Exchange:     "",
		RoutingKey:   "",
		Mandatory:    false,
		Immediate:    false,
		ContentType:  "application/json",
		DeliveryMode: 2, // persistent
		Priority:     0,
		Expiration:   "",
	}
}

// Producer handles publishing messages to RabbitMQ
type Producer struct {
	conn *Connection
}

// NewProducer creates a new RabbitMQ producer
func NewProducer(conn *Connection) *Producer {
	return &Producer{
		conn: conn,
	}
}

// PublishResult contains information about the published message
type PublishResult struct {
	MessageID  string
	Exchange   string
	RoutingKey string
	Timestamp  time.Time
}

// Publish publishes a message to RabbitMQ with auto-generated message ID
func (p *Producer) Publish(ctx context.Context, body []byte, config PublishConfig) (*PublishResult, error) {
	return p.PublishWithID(ctx, body, config, uuid.New().String())
}

// PublishWithID publishes a message to RabbitMQ with a custom message ID
func (p *Producer) PublishWithID(ctx context.Context, body []byte, config PublishConfig, messageID string) (*PublishResult, error) {
	if !p.conn.IsConnected() {
		return nil, fmt.Errorf("not connected to RabbitMQ")
	}

	channel, err := p.conn.GetChannel()
	if err != nil {
		return nil, err
	}

	timestamp := time.Now()

	// Create message headers
	headers := amqp.Table{
		"message_id": messageID,
		"timestamp":  timestamp.UnixNano(),
	}

	// Create publishing
	msg := amqp.Publishing{
		Headers:         headers,
		ContentType:     config.ContentType,
		ContentEncoding: "",
		DeliveryMode:    config.DeliveryMode,
		Priority:        config.Priority,
		CorrelationId:   messageID,
		Expiration:      config.Expiration,
		MessageId:       messageID,
		Timestamp:       timestamp,
		Type:            "",
		UserId:          "",
		AppId:           "",
		Body:            body,
	}

	// Use context for timeout
	publishCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Create confirmation channel
	if err := channel.Confirm(false); err != nil {
		return nil, fmt.Errorf("failed to put channel in confirm mode: %w", err)
	}

	confirms := channel.NotifyPublish(make(chan amqp.Confirmation, 1))

	// Publish the message
	if err := channel.Publish(
		config.Exchange,
		config.RoutingKey,
		config.Mandatory,
		config.Immediate,
		msg,
	); err != nil {
		return nil, fmt.Errorf("failed to publish message: %w", err)
	}

	// Wait for confirmation
	select {
	case <-publishCtx.Done():
		return nil, fmt.Errorf("publish confirmation timeout: %w", publishCtx.Err())
	case confirmation := <-confirms:
		if !confirmation.Ack {
			return nil, fmt.Errorf("message not acknowledged by server")
		}
	}

	return &PublishResult{
		MessageID:  messageID,
		Exchange:   config.Exchange,
		RoutingKey: config.RoutingKey,
		Timestamp:  timestamp,
	}, nil
}

// PublishJSON publishes a JSON message to RabbitMQô
func (p *Producer) PublishJSON(ctx context.Context, body []byte, config PublishConfig) (*PublishResult, error) {
	config.ContentType = "application/json"
	return p.Publish(ctx, body, config)
}

// PublishBatch publishes multiple messages in a transaction
func (p *Producer) PublishBatch(ctx context.Context, messages [][]byte, config PublishConfig) ([]*PublishResult, error) {
	if !p.conn.IsConnected() {
		return nil, fmt.Errorf("not connected to RabbitMQ")
	}

	// Create a new channel for the transaction
	channel, err := p.conn.CreateChannel()
	if err != nil {
		return nil, err
	}
	defer channel.Close()

	// Start a transaction
	if err := channel.Tx(); err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}

	results := make([]*PublishResult, 0, len(messages))

	// Publish all messages
	for _, body := range messages {
		timestamp := time.Now()
		messageID := uuid.New().String()

		// Create message headers
		headers := amqp.Table{
			"message_id": messageID,
			"timestamp":  timestamp.UnixNano(),
		}

		// Create publishing
		msg := amqp.Publishing{
			Headers:         headers,
			ContentType:     config.ContentType,
			ContentEncoding: "",
			DeliveryMode:    config.DeliveryMode,
			Priority:        config.Priority,
			CorrelationId:   messageID,
			Expiration:      config.Expiration,
			MessageId:       messageID,
			Timestamp:       timestamp,
			Type:            "",
			UserId:          "",
			AppId:           "",
			Body:            body,
		}

		// Publish the message
		if err := channel.Publish(
			config.Exchange,
			config.RoutingKey,
			config.Mandatory,
			config.Immediate,
			msg,
		); err != nil {
			// Rollback the transaction if any message fails
			channel.TxRollback()
			return nil, fmt.Errorf("failed to publish message: %w", err)
		}

		results = append(results, &PublishResult{
			MessageID:  messageID,
			Exchange:   config.Exchange,
			RoutingKey: config.RoutingKey,
			Timestamp:  timestamp,
		})
	}

	// Commit the transaction
	if err := channel.TxCommit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return results, nil
}
