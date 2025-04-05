package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/Piyushhbhutoria/tote-assignment/internal/models"
)

// Producer represents a Kafka producer
type Producer struct {
	producer sarama.SyncProducer
	topic    string
}

// NewProducer creates a new Kafka producer
func NewProducer(cfg *Config) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = cfg.RetryAttempts
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(cfg.Brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %v", err)
	}

	return &Producer{
		producer: producer,
		topic:    cfg.Topic,
	}, nil
}

// Close closes the producer
func (p *Producer) Close() error {
	return p.producer.Close()
}

// SendEvent sends an event to Kafka
func (p *Producer) SendEvent(ctx context.Context, event *models.Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %v", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Value: sarama.StringEncoder(data),
		Key:   sarama.StringEncoder(event.ID),
	}

	_, _, err = p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}

	return nil
}
