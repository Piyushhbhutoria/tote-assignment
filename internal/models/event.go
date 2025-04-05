package models

import "time"

// EventType represents different types of POS events
type EventType string

const (
	EventEmployeeLogin    EventType = "EMPLOYEE_LOGIN"
	EventEmployeeLogout   EventType = "EMPLOYEE_LOGOUT"
	EventStartBasket      EventType = "START_BASKET"
	EventCustomerIdentify EventType = "CUSTOMER_IDENTIFY"
	EventAddItem          EventType = "ADD_ITEM"
	EventFinalizeSubtotal EventType = "FINALIZE_SUBTOTAL"
	EventPaymentComplete  EventType = "PAYMENT_COMPLETE"
)

// Event represents a base POS event
type Event struct {
	ID        string    `json:"id"`
	Type      EventType `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Payload   any       `json:"payload"`
}

// BasePayload contains common fields for all event payloads
type BasePayload struct {
	TerminalID string `json:"terminal_id"`
	StoreID    string `json:"store_id"`
}
