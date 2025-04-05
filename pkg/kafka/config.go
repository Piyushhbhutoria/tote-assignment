package kafka

import "time"

// Config holds Kafka configuration
type Config struct {
	Brokers        []string
	Topic          string
	ConsumerGroup  string
	RetryAttempts  int
	RetryDelay     time.Duration
	CommitInterval time.Duration
}

// NewDefaultConfig returns a default configuration
func NewDefaultConfig() *Config {
	return &Config{
		Brokers:        []string{"localhost:9092"},
		Topic:          "pos_events",
		ConsumerGroup:  "pos_consumer_group",
		RetryAttempts:  3,
		RetryDelay:     time.Second * 5,
		CommitInterval: time.Second * 1,
	}
}
