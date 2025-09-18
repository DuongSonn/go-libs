package _kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
)

// Producer handles producing messages to Kafka topics
type Producer struct {
	config Config
}

// NewProducer creates a new Kafka producer with the provided configuration
func NewProducer(cfg Config) *Producer {
	return &Producer{
		config: cfg,
	}
}

// ProduceResult contains information about the produced message
type ProduceResult struct {
	MessageID string
	Topic     string
	Partition int32
	Offset    int64
}

// Produce sends a message to the specified topic with auto-generated message ID
func (p *Producer) Produce(ctx context.Context, topic string, key []byte, value []byte) (*ProduceResult, error) {
	return p.ProduceWithID(ctx, topic, key, value, uuid.New().String())
}

// ProduceWithID sends a message to the specified topic with a custom message ID
func (p *Producer) ProduceWithID(ctx context.Context, topic string, key []byte, value []byte, messageID string) (*ProduceResult, error) {
	client, err := kgo.NewClient(kgo.SeedBrokers(p.config.Brokers...))
	if err != nil {
		return nil, err
	}
	defer client.Close()

	if _, err = kadm.NewClient(client).CreateTopic(ctx, 1, -1, nil, topic); err != nil {
		return nil, err
	}

	// Create headers with message ID
	headers := []kgo.RecordHeader{
		{
			Key:   "message_id",
			Value: []byte(messageID),
		},
		{
			Key:   "timestamp",
			Value: []byte(fmt.Sprintf("%d", time.Now().UnixNano())),
		},
	}

	record := &kgo.Record{
		Key:       key,
		Topic:     topic,
		Timestamp: time.Now(),
		Value:     value,
		Headers:   headers,
	}

	result := &ProduceResult{
		MessageID: messageID,
		Topic:     topic,
	}

	// Create a channel to wait for the produce callback
	done := make(chan error, 1)

	client.Produce(ctx, record, func(r *kgo.Record, err error) {
		if err != nil {
			done <- err
			return
		}

		// Store partition and offset information
		result.Partition = r.Partition
		result.Offset = r.Offset
		done <- nil
	})

	// Wait for the produce callback
	if err := <-done; err != nil {
		return nil, fmt.Errorf("failed to produce message: %w", err)
	}

	return result, nil
}
