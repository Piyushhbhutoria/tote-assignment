package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/Piyushhbhutoria/tote-assignment/internal/models"
	"github.com/Piyushhbhutoria/tote-assignment/pkg/kafka"
	"github.com/google/uuid"
)

// Producer interface for testing
type Producer interface {
	SendEvent(ctx context.Context, event *models.Event) error
	Close() error
}

func main() {
	log.Println("Starting POS Event Producer...")

	// Create Kafka producer
	cfg := kafka.NewDefaultConfig()
	producer, err := kafka.NewProducer(cfg)
	if err != nil {
		log.Fatalf("Failed to create producer: %v", err)
	}
	defer producer.Close()

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	// Start generating events
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		generateEvents(ctx, producer)
	}()

	// Wait for interrupt signal
	<-signals
	log.Println("Shutting down...")
	cancel()
	wg.Wait()
}

func generateEvents(ctx context.Context, producer Producer) {
	terminals := []string{"POS001", "POS002", "POS003"}
	employees := []string{"EMP001", "EMP002", "EMP003"}
	customers := []string{"CUST001", "CUST002", "CUST003"}
	items := []struct {
		id    string
		price float64
	}{
		{"ITEM001", 10.99},
		{"ITEM002", 15.99},
		{"ITEM003", 5.99},
		{"ITEM004", 20.99},
		{"ITEM005", 8.99},
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Simulate a complete shopping session
			terminalID := terminals[randInt(0, len(terminals))]
			employeeID := employees[randInt(0, len(employees))]
			customerID := customers[randInt(0, len(customers))]
			basketID := uuid.New().String()

			// Employee Login
			sendEvent(ctx, producer, models.EventEmployeeLogin, map[string]interface{}{
				"terminal_id": terminalID,
				"store_id":    "STORE001",
				"employee_id": employeeID,
			})

			// Start Basket
			sendEvent(ctx, producer, models.EventStartBasket, map[string]interface{}{
				"terminal_id": terminalID,
				"store_id":    "STORE001",
				"employee_id": employeeID,
				"basket_id":   basketID,
			})

			// Customer Identification
			sendEvent(ctx, producer, models.EventCustomerIdentify, map[string]interface{}{
				"terminal_id": terminalID,
				"store_id":    "STORE001",
				"employee_id": employeeID,
				"basket_id":   basketID,
				"customer_id": customerID,
			})

			// Add 2-4 items
			numItems := randInt(2, 5)
			for i := 0; i < numItems; i++ {
				item := items[randInt(0, len(items))]
				sendEvent(ctx, producer, models.EventAddItem, map[string]interface{}{
					"terminal_id": terminalID,
					"store_id":    "STORE001",
					"employee_id": employeeID,
					"basket_id":   basketID,
					"item_id":     item.id,
					"price":       item.price,
					"quantity":    1,
				})
				time.Sleep(time.Millisecond * 500) // Simulate realistic timing
			}

			// Finalize Subtotal
			sendEvent(ctx, producer, models.EventFinalizeSubtotal, map[string]interface{}{
				"terminal_id": terminalID,
				"store_id":    "STORE001",
				"employee_id": employeeID,
				"basket_id":   basketID,
			})

			// Payment Complete
			sendEvent(ctx, producer, models.EventPaymentComplete, map[string]interface{}{
				"terminal_id":    terminalID,
				"store_id":       "STORE001",
				"employee_id":    employeeID,
				"basket_id":      basketID,
				"payment_method": "CARD",
			})

			// Employee Logout
			sendEvent(ctx, producer, models.EventEmployeeLogout, map[string]interface{}{
				"terminal_id": terminalID,
				"store_id":    "STORE001",
				"employee_id": employeeID,
			})

			// Wait before starting next session
			time.Sleep(time.Second * 5)
		}
	}
}

func sendEvent(ctx context.Context, producer Producer, eventType models.EventType, payload interface{}) {
	event := &models.Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Timestamp: time.Now(),
		Payload:   payload,
	}

	if err := producer.SendEvent(ctx, event); err != nil {
		log.Printf("Failed to send event: %v", err)
		return
	}

	log.Printf("Sent event: %s", eventType)
}

func randInt(min, max int) int {
	return min + time.Now().Nanosecond()%(max-min)
}
