package employee_tracker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Piyushhbhutoria/tote-assignment/internal/models"
	"github.com/Piyushhbhutoria/tote-assignment/pkg/database"
	"github.com/jackc/pgx/v4"
)

// Plugin implements the employee time tracking plugin
type Plugin struct {
	db     *database.Connection
	active bool
	config map[string]interface{}
}

// New creates a new employee time tracker plugin
func New(db *database.Connection) *Plugin {
	return &Plugin{
		db:     db,
		active: false,
		config: make(map[string]interface{}),
	}
}

// Name returns the plugin name
func (p *Plugin) Name() string {
	return "employee_time_tracker"
}

// Description returns the plugin description
func (p *Plugin) Description() string {
	return "Tracks employee login/logout events and calculates time spent at terminals"
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

// ProcessEvent handles employee login/logout events
func (p *Plugin) ProcessEvent(ctx context.Context, event *models.Event) ([]*models.Event, error) {
	if !p.active {
		return nil, nil
	}

	switch event.Type {
	case models.EventEmployeeLogin:
		return p.handleLogin(ctx, event)
	case models.EventEmployeeLogout:
		return p.handleLogout(ctx, event)
	default:
		return nil, nil
	}
}

func (p *Plugin) handleLogin(ctx context.Context, event *models.Event) ([]*models.Event, error) {
	var payload struct {
		EmployeeID string `json:"employee_id"`
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

	// Check if employee is already logged in somewhere
	var currentTerminal string
	err = p.db.Pool().QueryRow(ctx, `
		SELECT current_terminal_id 
		FROM employees 
		WHERE employee_id = $1 AND current_terminal_id IS NOT NULL
	`, payload.EmployeeID).Scan(&currentTerminal)

	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("failed to check employee status: %v", err)
	}

	// If employee is logged in elsewhere, generate auto-logout event
	var events []*models.Event
	if err != pgx.ErrNoRows && currentTerminal != payload.TerminalID {
		autoLogout := &models.Event{
			Type:      models.EventEmployeeLogout,
			Timestamp: event.Timestamp,
			Payload: map[string]interface{}{
				"employee_id": payload.EmployeeID,
				"terminal_id": currentTerminal,
				"store_id":    payload.StoreID,
				"auto_logout": true,
				"reason":      "Login detected at different terminal",
			},
		}
		events = append(events, autoLogout)
	}

	// Start new session
	batch := &pgx.Batch{}

	// Update employee status
	batch.Queue(`
		INSERT INTO employees (employee_id, current_terminal_id, last_login)
		VALUES ($1, $2, $3)
		ON CONFLICT (employee_id) 
		DO UPDATE SET 
			current_terminal_id = $2,
			last_login = $3
	`, payload.EmployeeID, payload.TerminalID, event.Timestamp)

	// Create new session
	batch.Queue(`
		INSERT INTO employee_sessions 
		(employee_id, terminal_id, login_time)
		VALUES ($1, $2, $3)
	`, payload.EmployeeID, payload.TerminalID, event.Timestamp)

	br := p.db.Pool().SendBatch(ctx, batch)
	defer br.Close()

	if err := br.Close(); err != nil {
		return nil, fmt.Errorf("failed to execute batch: %v", err)
	}

	return events, nil
}

func (p *Plugin) handleLogout(ctx context.Context, event *models.Event) ([]*models.Event, error) {
	var payload struct {
		EmployeeID string `json:"employee_id"`
		TerminalID string `json:"terminal_id"`
	}

	data, err := json.Marshal(event.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %v", err)
	}

	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %v", err)
	}

	batch := &pgx.Batch{}

	// Update employee status
	batch.Queue(`
		UPDATE employees 
		SET current_terminal_id = NULL,
			last_logout = $2
		WHERE employee_id = $1
	`, payload.EmployeeID, event.Timestamp)

	// Update session
	batch.Queue(`
		UPDATE employee_sessions 
		SET logout_time = $3,
			duration_minutes = EXTRACT(EPOCH FROM ($3 - login_time))/60
		WHERE employee_id = $1 
		AND terminal_id = $2 
		AND logout_time IS NULL
	`, payload.EmployeeID, payload.TerminalID, event.Timestamp)

	br := p.db.Pool().SendBatch(ctx, batch)
	defer br.Close()

	if err := br.Close(); err != nil {
		return nil, fmt.Errorf("failed to execute batch: %v", err)
	}

	return nil, nil
}
