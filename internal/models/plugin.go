package models

import (
	"context"
	"time"
)

// Plugin represents a plugin that can process events
type Plugin interface {
	Name() string
	Description() string
	IsActive() bool
	SetActive(bool)
	Configure(map[string]interface{}) error
	ProcessEvent(context.Context, *Event) ([]*Event, error)
}

// PluginStats tracks statistics for a plugin
type PluginStats struct {
	EventsProcessed int        `json:"eventsProcessed"`
	LastProcessed   *time.Time `json:"lastProcessed,omitempty"`
	ErrorCount      int        `json:"errorCount"`
}
