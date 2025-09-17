package _kafka

import (
	"context"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
)

// Connection manages Kafka connections and service registrations for topics
type Connection struct {
	config   Config                       // Kafka connection configuration
	services map[string]IMessageProcessor // Map of topic to message processor services
	clients  map[string]*kgo.Client       // Map of topic to Kafka clients
}

// NewConnection creates a new Kafka connection with the provided configuration
func NewConnection(cfg Config) *Connection {
	return &Connection{
		config:   cfg,
		services: make(map[string]IMessageProcessor),
		clients:  make(map[string]*kgo.Client),
	}
}

// RegisterService registers a message processor service for a specific topic
func (c *Connection) RegisterService(topic string, service IMessageProcessor) {
	c.services[topic] = service
}

// Connect establishes connections to Kafka for all registered services
// It validates that all topics with registered services are in the configuration
func (c *Connection) Connect(ctx context.Context) error {
	if err := c.config.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	// Ensure at least one service is registered
	if len(c.services) == 0 {
		return fmt.Errorf("no services registered")
	}

	// Create a map of valid topics from config for quick lookup
	validTopics := make(map[string]bool)
	for _, t := range c.config.Topics {
		validTopics[t] = true
	}

	// Create clients for each topic with a registered service
	for topic, service := range c.services {
		if !validTopics[topic] {
			return fmt.Errorf("topic %s is not in config", topic)
		}

		s := &splitConsume{
			consumers: make(map[tp]*pconsumer),
			service:   service,
		}

		opts := []kgo.Opt{
			kgo.SeedBrokers(c.config.Brokers...),
			kgo.ConsumerGroup(c.config.Group),
			kgo.ConsumeTopics(topic),
			kgo.OnPartitionsAssigned(s.assigned),
			kgo.OnPartitionsRevoked(s.lost),
			kgo.OnPartitionsLost(s.lost),
			kgo.DisableAutoCommit(),
			kgo.BlockRebalanceOnPoll(),
		}

		cl, err := kgo.NewClient(opts...)
		if err != nil {
			// Clean up any clients already created if we encounter an error
			for _, client := range c.clients {
				client.Close()
			}
			return err
		}
		if err = cl.Ping(ctx); err != nil {
			// Clean up any clients already created if we encounter an error
			for _, client := range c.clients {
				client.Close()
			}
			return err
		}

		c.clients[topic] = cl
		go s.poll(cl) // Start polling for messages in a separate goroutine
	}

	return nil
}

// Close closes all Kafka client connections and cleans up resources
func (c *Connection) Close() {
	for _, client := range c.clients {
		client.Close()
	}

	c.clients = make(map[string]*kgo.Client)
}
