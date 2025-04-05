package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/IBM/sarama"
	"github.com/Piyushhbhutoria/tote-assignment/internal/models"
)

// MessageHandler is a function that processes a Kafka message
type MessageHandler func(context.Context, *models.Event) error

// Consumer represents a Kafka consumer
type Consumer struct {
	consumer       sarama.ConsumerGroup
	topic          string
	handler        MessageHandler
	ready          chan bool
	commitInterval int64
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(cfg *Config, handler MessageHandler) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	group, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.ConsumerGroup, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %v", err)
	}

	return &Consumer{
		consumer:       group,
		topic:          cfg.Topic,
		handler:        handler,
		ready:          make(chan bool),
		commitInterval: cfg.CommitInterval.Milliseconds(),
	}, nil
}

// Start starts consuming messages
func (c *Consumer) Start(ctx context.Context) error {
	topics := []string{c.topic}
	for {
		err := c.consumer.Consume(ctx, topics, c)
		if err != nil {
			return fmt.Errorf("error from consumer: %v", err)
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}

		c.ready = make(chan bool)
	}
}

// Close closes the consumer
func (c *Consumer) Close() error {
	return c.consumer.Close()
}

// Setup is run at the beginning of a new session
func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	close(c.ready)
	return nil
}

// Cleanup is run at the end of a session
func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim processes messages from a partition
func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				return nil
			}

			var event models.Event
			if err := json.Unmarshal(message.Value, &event); err != nil {
				log.Printf("Error unmarshaling event: %v", err)
				continue
			}

			if err := c.handler(session.Context(), &event); err != nil {
				log.Printf("Error handling event: %v", err)
				continue
			}

			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}
