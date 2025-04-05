package kafka

import (
	"context"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/Piyushhbhutoria/tote-assignment/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSaramaProducer mocks the sarama.SyncProducer interface
type MockSaramaProducer struct {
	mock.Mock
}

func (m *MockSaramaProducer) SendMessage(msg *sarama.ProducerMessage) (partition int32, offset int64, err error) {
	args := m.Called(msg)
	return args.Get(0).(int32), args.Get(1).(int64), args.Error(2)
}

func (m *MockSaramaProducer) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockSaramaConsumerGroup mocks the sarama.ConsumerGroup interface
type MockSaramaConsumerGroup struct {
	mock.Mock
}

func (m *MockSaramaConsumerGroup) Consume(ctx context.Context, topics []string, handler sarama.ConsumerGroupHandler) error {
	args := m.Called(ctx, topics, handler)
	return args.Error(0)
}

func (m *MockSaramaConsumerGroup) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockSaramaConsumerGroup) Errors() <-chan error {
	args := m.Called()
	return args.Get(0).(chan error)
}

func TestProducerSendEvent(t *testing.T) {
	// Create mock Sarama producer
	mockSarama := new(MockSaramaProducer)
	mockSarama.On("SendMessage", mock.Anything).Return(int32(0), int64(1), nil)
	mockSarama.On("Close").Return(nil)

	// Create producer with mock
	producer := &Producer{
		producer: mockSarama,
		topic:    "test_topic",
	}

	// Create test event
	event := &models.Event{
		ID:        "test_event",
		Type:      models.EventEmployeeLogin,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"employee_id": "EMP001",
			"terminal_id": "POS001",
		},
	}

	// Send event
	err := producer.SendEvent(context.Background(), event)
	assert.NoError(t, err)

	// Verify mock was called with correct message
	mockSarama.AssertCalled(t, "SendMessage", mock.MatchedBy(func(msg *sarama.ProducerMessage) bool {
		return msg.Topic == "test_topic"
	}))

	// Close producer
	err = producer.Close()
	assert.NoError(t, err)
	mockSarama.AssertCalled(t, "Close")
}

func TestConsumerStart(t *testing.T) {
	// Create mock Sarama consumer group
	mockSarama := new(MockSaramaConsumerGroup)
	mockSarama.On("Consume", mock.Anything, []string{"test_topic"}, mock.Anything).Return(nil)
	mockSarama.On("Close").Return(nil)
	mockSarama.On("Errors").Return(make(chan error))

	// Create handler function
	handlerCalled := false
	handler := func(ctx context.Context, event *models.Event) error {
		handlerCalled = true
		assert.Equal(t, "test_event", event.ID)
		assert.Equal(t, models.EventEmployeeLogin, event.Type)
		return nil
	}

	// Create consumer with mock
	consumer := &Consumer{
		consumer: mockSarama,
		topic:    "test_topic",
		handler:  handler,
		ready:    make(chan bool),
	}

	// Start consumer in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		err := consumer.Start(ctx)
		assert.NoError(t, err)
	}()

	// Simulate message processing by calling handler directly
	event := &models.Event{
		ID:        "test_event",
		Type:      models.EventEmployeeLogin,
		Timestamp: time.Now(),
	}
	err := handler(context.Background(), event)
	assert.NoError(t, err)
	assert.True(t, handlerCalled)

	// Close consumer
	err = consumer.Close()
	assert.NoError(t, err)
	mockSarama.AssertCalled(t, "Close")
}

func TestConsumerSetup(t *testing.T) {
	consumer := &Consumer{
		ready: make(chan bool),
	}

	err := consumer.Setup(nil)
	assert.NoError(t, err)

	// Verify ready channel is closed
	select {
	case <-consumer.ready:
		// Channel closed as expected
	default:
		t.Error("Ready channel not closed")
	}
}

func TestConsumerCleanup(t *testing.T) {
	consumer := &Consumer{}
	err := consumer.Cleanup(nil)
	assert.NoError(t, err)
}
