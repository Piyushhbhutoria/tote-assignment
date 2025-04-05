package customer_lookup

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Piyushhbhutoria/tote-assignment/internal/models"
	"github.com/Piyushhbhutoria/tote-assignment/pkg/database"
)

// Plugin implements the customer lookup plugin
type Plugin struct {
	db     *database.Connection
	active bool
	config map[string]interface{}
}

// New creates a new customer lookup plugin
func New(db *database.Connection) *Plugin {
	return &Plugin{
		db:     db,
		active: false,
		config: make(map[string]interface{}),
	}
}

// Name returns the plugin name
func (p *Plugin) Name() string {
	return "customer_lookup"
}

// Description returns the plugin description
func (p *Plugin) Description() string {
	return "Identifies customer events and enriches them with customer data"
}

// IsActive returns whether the plugin is active
func (p *Plugin) IsActive() bool {
	return p.active
}

// SetActive enables or disables the plugin
func (p *Plugin) SetActive(active bool) {
	p.active = active
}

// Configure applies the plugin configuration
func (p *Plugin) Configure(config map[string]interface{}) error {
	p.config = config
	return nil
}

// ProcessEvent handles customer identification events
func (p *Plugin) ProcessEvent(ctx context.Context, event *models.Event) ([]*models.Event, error) {
	if !p.active {
		return nil, nil
	}

	if event.Type != models.EventCustomerIdentify {
		return nil, nil
	}

	return p.handleCustomerIdentified(ctx, event)
}

func (p *Plugin) handleCustomerIdentified(ctx context.Context, event *models.Event) ([]*models.Event, error) {
	var payload struct {
		CustomerID string `json:"customer_id"`
		BasketID   string `json:"basket_id"`
		TerminalID string `json:"terminal_id"`
		StoreID    string `json:"store_id"`
	}

	data, err := json.Marshal(event.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %v", err)
	}

	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %v", err)
	}

	// Fetch customer data
	customerData, err := p.getCustomerData(ctx, payload.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer data: %v", err)
	}

	// Update last seen timestamp
	if err := p.updateLastSeen(ctx, payload.CustomerID); err != nil {
		return nil, fmt.Errorf("failed to update last seen: %v", err)
	}

	// Create customer data event
	customerEvent := &models.Event{
		Type:      "CUSTOMER_DATA",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"customer_id": payload.CustomerID,
			"basket_id":   payload.BasketID,
			"terminal_id": payload.TerminalID,
			"store_id":    payload.StoreID,
			"data":        customerData,
		},
	}

	return []*models.Event{customerEvent}, nil
}

func (p *Plugin) getCustomerData(ctx context.Context, customerID string) (map[string]interface{}, error) {
	var data json.RawMessage
	err := p.db.Pool().QueryRow(ctx, `
		SELECT data
		FROM customers
		WHERE customer_id = $1
	`, customerID).Scan(&data)

	if err != nil {
		// If customer not found, simulate fetching from remote system
		customerData := p.simulateRemoteLookup(customerID)

		// Store the data for future lookups
		_, err = p.db.Pool().Exec(ctx, `
			INSERT INTO customers (customer_id, data, last_seen)
			VALUES ($1, $2, CURRENT_TIMESTAMP)
			ON CONFLICT (customer_id) 
			DO UPDATE SET 
				data = $2,
				updated_at = CURRENT_TIMESTAMP
		`, customerID, customerData)
		if err != nil {
			return nil, fmt.Errorf("failed to store customer data: %v", err)
		}

		return customerData, nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal customer data: %v", err)
	}

	return result, nil
}

func (p *Plugin) updateLastSeen(ctx context.Context, customerID string) error {
	_, err := p.db.Pool().Exec(ctx, `
		UPDATE customers
		SET last_seen = CURRENT_TIMESTAMP
		WHERE customer_id = $1
	`, customerID)
	return err
}

// simulateRemoteLookup simulates fetching customer data from a remote system
func (p *Plugin) simulateRemoteLookup(customerID string) map[string]interface{} {
	// In a real system, this would make an API call to a remote customer service
	// For now, we'll generate some dummy data
	return map[string]interface{}{
		"name": fmt.Sprintf("Customer %s", customerID[len(customerID)-4:]),
		"tier": "regular",
		"preferences": map[string]interface{}{
			"marketing_emails": true,
			"notifications":    true,
		},
		"last_purchase":   time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
		"total_purchases": 5,
		"average_basket":  45.99,
		"source":          "remote_lookup",
		"fetched_at":      time.Now().Format(time.RFC3339),
	}
}
