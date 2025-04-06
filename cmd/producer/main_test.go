package main

import (
	"context"
	"testing"
	"time"

	"github.com/Piyushhbhutoria/tote-assignment/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProducer is a mock implementation of the kafka.Producer
type MockProducer struct {
	mock.Mock
}

func (m *MockProducer) SendEvent(ctx context.Context, event *models.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockProducer) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestSendEvent(t *testing.T) {
	// Create mock producer
	mockProducer := &MockProducer{}

	// Create test event
	ctx := context.Background()
	eventType := models.EventEmployeeLogin
	payload := map[string]interface{}{
		"terminal_id": "POS001",
		"store_id":    "STORE001",
		"employee_id": "EMP001",
	}

	// Setup expectations
	mockProducer.On("SendEvent", ctx, mock.MatchedBy(func(e *models.Event) bool {
		return e.Type == eventType &&
			e.Payload.(map[string]interface{})["terminal_id"] == "POS001" &&
			e.Payload.(map[string]interface{})["store_id"] == "STORE001" &&
			e.Payload.(map[string]interface{})["employee_id"] == "EMP001"
	})).Return(nil).Once()

	// Call function
	sendEvent(ctx, mockProducer, eventType, payload)

	// Assert expectations
	mockProducer.AssertExpectations(t)
}

func TestGenerateEvents(t *testing.T) {
	// Create mock producer
	mockProducer := &MockProducer{}

	// Setup context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Setup expectations for all event types
	mockProducer.On("SendEvent", mock.Anything, mock.MatchedBy(func(e *models.Event) bool {
		return e.Type == models.EventEmployeeLogin
	})).Return(nil).Maybe()

	mockProducer.On("SendEvent", mock.Anything, mock.MatchedBy(func(e *models.Event) bool {
		return e.Type == models.EventStartBasket
	})).Return(nil).Maybe()

	mockProducer.On("SendEvent", mock.Anything, mock.MatchedBy(func(e *models.Event) bool {
		return e.Type == models.EventCustomerIdentify
	})).Return(nil).Maybe()

	mockProducer.On("SendEvent", mock.Anything, mock.MatchedBy(func(e *models.Event) bool {
		return e.Type == models.EventAddItem
	})).Return(nil).Maybe()

	mockProducer.On("SendEvent", mock.Anything, mock.MatchedBy(func(e *models.Event) bool {
		return e.Type == models.EventFinalizeSubtotal
	})).Return(nil).Maybe()

	mockProducer.On("SendEvent", mock.Anything, mock.MatchedBy(func(e *models.Event) bool {
		return e.Type == models.EventPaymentComplete
	})).Return(nil).Maybe()

	mockProducer.On("SendEvent", mock.Anything, mock.MatchedBy(func(e *models.Event) bool {
		return e.Type == models.EventEmployeeLogout
	})).Return(nil).Maybe()

	// Run event generation
	generateEvents(ctx, mockProducer)

	// Assert expectations
	mockProducer.AssertExpectations(t)
}

func TestGenerateEventsWithError(t *testing.T) {
	// Create mock producer
	mockProducer := &MockProducer{}

	// Setup context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Setup expectations to return error
	mockProducer.On("SendEvent", mock.Anything, mock.Anything).Return(assert.AnError).Maybe()

	// Run event generation
	generateEvents(ctx, mockProducer)

	// Assert expectations
	mockProducer.AssertExpectations(t)
}

func TestRandInt(t *testing.T) {
	// Test min and max bounds
	min := 1
	max := 5
	for i := 0; i < 100; i++ {
		result := randInt(min, max)
		assert.GreaterOrEqual(t, result, min)
		assert.Less(t, result, max)
	}
}

func TestSendEventWithError(t *testing.T) {
	// Create mock producer
	mockProducer := &MockProducer{}

	// Create test event
	ctx := context.Background()
	payload := map[string]interface{}{
		"terminal_id": "POS001",
		"store_id":    "STORE001",
		"employee_id": "EMP001",
	}

	// Setup expectations
	mockProducer.On("SendEvent", ctx, mock.Anything).Return(assert.AnError).Once()

	// Call function
	sendEvent(ctx, mockProducer, models.EventEmployeeLogin, payload)

	// Assert expectations
	mockProducer.AssertExpectations(t)
}

func TestSendEventWithCancelledContext(t *testing.T) {
	// Create mock producer
	mockProducer := &MockProducer{}

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Create test event
	payload := map[string]interface{}{
		"terminal_id": "POS001",
		"store_id":    "STORE001",
		"employee_id": "EMP001",
	}

	// Setup expectations
	mockProducer.On("SendEvent", ctx, mock.Anything).Return(context.Canceled).Once()

	// Call function
	sendEvent(ctx, mockProducer, models.EventEmployeeLogin, payload)

	// Assert expectations
	mockProducer.AssertExpectations(t)
}

func TestGenerateEventsWithCancelledContext(t *testing.T) {
	// Create mock producer
	mockProducer := &MockProducer{}

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Run event generation
	generateEvents(ctx, mockProducer)

	// Assert that no events were sent
	mockProducer.AssertNotCalled(t, "SendEvent")
}
