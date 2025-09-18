package _rabbitmq

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

// Connection manages the connection to RabbitMQ
type Connection struct {
	config      *Config
	conn        *amqp.Connection
	channel     *amqp.Channel
	isConnected bool
	mu          sync.RWMutex

	// Reconnection
	reconnectCh chan struct{}
	closeCh     chan struct{}

	// Connection status
	connError  error
	connClosed bool
}

// NewConnection creates a new RabbitMQ connection
func NewConnection(cfg *Config) *Connection {
	return &Connection{
		config:      cfg,
		isConnected: false,
		reconnectCh: make(chan struct{}),
		closeCh:     make(chan struct{}),
	}
}

// Connect establishes a connection to RabbitMQ
func (c *Connection) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.config.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	// Connect with timeout
	connectCtx, cancel := context.WithTimeout(ctx, c.config.ConnTimeout)
	defer cancel()

	var err error
	var conn *amqp.Connection

	// Try to connect with retries
	for i := 0; i <= c.config.MaxRetries; i++ {
		select {
		case <-connectCtx.Done():
			return fmt.Errorf("connection timeout: %w", connectCtx.Err())
		default:
			conn, err = amqp.Dial(c.config.GetURI())
			if err == nil {
				break
			}

			if i < c.config.MaxRetries {
				time.Sleep(c.config.RetryTimeout)
			}
		}
	}

	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ after %d retries: %w", c.config.MaxRetries, err)
	}

	c.conn = conn

	// Create channel
	channel, err := conn.Channel()
	if err != nil {
		c.conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}
	c.channel = channel

	c.isConnected = true
	c.connClosed = false

	// Start reconnection goroutine
	go c.handleReconnection()

	// Monitor connection closure
	go func() {
		// Will be closed when connection is closed
		connCloseChan := c.conn.NotifyClose(make(chan *amqp.Error))

		select {
		case err := <-connCloseChan:
			c.mu.Lock()
			c.isConnected = false
			c.connError = err
			c.mu.Unlock()

			// Trigger reconnection
			select {
			case c.reconnectCh <- struct{}{}:
			default:
			}

		case <-c.closeCh:
			// Connection was closed intentionally
			return
		}
	}()

	return nil
}

// handleReconnection attempts to reconnect when the connection is lost
func (c *Connection) handleReconnection() {
	for {
		select {
		case <-c.reconnectCh:
			c.mu.RLock()
			if c.connClosed {
				c.mu.RUnlock()
				return
			}
			c.mu.RUnlock()

			// Try to reconnect
			ctx, cancel := context.WithTimeout(context.Background(), c.config.ConnTimeout)
			err := c.Connect(ctx)
			cancel()

			if err != nil {
				// If reconnection fails, try again after delay
				time.Sleep(c.config.RetryTimeout)
				select {
				case c.reconnectCh <- struct{}{}:
				default:
				}
			}

		case <-c.closeCh:
			return
		}
	}
}

// Close closes the RabbitMQ connection
func (c *Connection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Signal reconnection goroutine to stop
	close(c.closeCh)

	c.connClosed = true

	var err error
	if c.channel != nil {
		err = c.channel.Close()
		c.channel = nil
	}

	if c.conn != nil {
		if closeErr := c.conn.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
		c.conn = nil
	}

	c.isConnected = false
	return err
}

// IsConnected checks if the connection is established
func (c *Connection) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isConnected
}

// GetChannel returns the AMQP channel
func (c *Connection) GetChannel() (*amqp.Channel, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.isConnected {
		return nil, fmt.Errorf("not connected to RabbitMQ")
	}

	return c.channel, nil
}

// CreateChannel creates a new AMQP channel
func (c *Connection) CreateChannel() (*amqp.Channel, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.isConnected {
		return nil, fmt.Errorf("not connected to RabbitMQ")
	}

	return c.conn.Channel()
}

// DeclareExchange declares a new exchange
func (c *Connection) DeclareExchange(cfg ExchangeConfig) error {
	channel, err := c.GetChannel()
	if err != nil {
		return err
	}

	return channel.ExchangeDeclare(
		cfg.Name,
		cfg.Type,
		cfg.Durable,
		cfg.AutoDelete,
		cfg.Internal,
		cfg.NoWait,
		nil, // arguments
	)
}

// DeclareQueue declares a new queue
func (c *Connection) DeclareQueue(cfg QueueConfig) (amqp.Queue, error) {
	channel, err := c.GetChannel()
	if err != nil {
		return amqp.Queue{}, err
	}

	return channel.QueueDeclare(
		cfg.Name,
		cfg.Durable,
		cfg.AutoDelete,
		cfg.Exclusive,
		cfg.NoWait,
		nil, // arguments
	)
}

// BindQueue binds a queue to an exchange
func (c *Connection) BindQueue(cfg BindingConfig) error {
	channel, err := c.GetChannel()
	if err != nil {
		return err
	}

	return channel.QueueBind(
		cfg.Queue,
		cfg.RoutingKey,
		cfg.Exchange,
		cfg.NoWait,
		nil, // arguments
	)
}
