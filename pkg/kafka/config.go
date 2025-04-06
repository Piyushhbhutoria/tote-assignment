package kafka

import (
	"os"
	"strings"
	"time"
)

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
		Brokers:        strings.Split(getEnvOrDefault("KAFKA_BROKERS", "localhost:9092"), ","),
		Topic:          getEnvOrDefault("KAFKA_TOPIC", "pos_events"),
		ConsumerGroup:  getEnvOrDefault("KAFKA_CONSUMER_GROUP", "pos_consumer_group"),
		RetryAttempts:  3,
		RetryDelay:     time.Second * 5,
		CommitInterval: time.Second * 1,
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
