package plugins

import (
	"context"

	"github.com/Piyushhbhutoria/tote-assignment/internal/models"
)

// Plugin defines the interface that all plugins must implement
type Plugin interface {
	// Name returns the unique identifier of the plugin
	Name() string

	// Description returns a human-readable description of the plugin
	Description() string

	// IsActive returns whether the plugin is currently active
	IsActive() bool

	// SetActive enables or disables the plugin
	SetActive(active bool)

	// Configure applies the plugin configuration
	Configure(config map[string]interface{}) error

	// ProcessEvent processes an event and returns any resulting events
	ProcessEvent(ctx context.Context, event *models.Event) ([]*models.Event, error)
}

// BasePlugin provides a basic implementation of the Plugin interface
type BasePlugin struct {
	name        string
	description string
	active      bool
	config      map[string]interface{}
}

func NewBasePlugin(name, description string) *BasePlugin {
	return &BasePlugin{
		name:        name,
		description: description,
		active:      false,
		config:      make(map[string]interface{}),
	}
}

func (p *BasePlugin) Name() string {
	return p.name
}

func (p *BasePlugin) Description() string {
	return p.description
}

func (p *BasePlugin) IsActive() bool {
	return p.active
}

func (p *BasePlugin) SetActive(active bool) {
	p.active = active
}

func (p *BasePlugin) Configure(config map[string]interface{}) error {
	p.config = config
	return nil
}
