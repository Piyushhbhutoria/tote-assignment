package plugins

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/Piyushhbhutoria/tote-assignment/internal/models"
	"github.com/Piyushhbhutoria/tote-assignment/pkg/database"
)

// Manager handles plugin registration and event routing
type Manager struct {
	db      *database.Connection
	plugins map[string]Plugin
	// Keep track of plugin order
	pluginOrder []string
	mu          sync.RWMutex
}

// NewManager creates a new plugin manager
func NewManager(db *database.Connection) *Manager {
	return &Manager{
		db:          db,
		plugins:     make(map[string]Plugin),
		pluginOrder: make([]string, 0),
	}
}

// RegisterPlugin registers a plugin with the manager
func (m *Manager) RegisterPlugin(p Plugin) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.plugins[p.Name()]; exists {
		return fmt.Errorf("plugin %s already registered", p.Name())
	}

	m.plugins[p.Name()] = p
	m.pluginOrder = append(m.pluginOrder, p.Name())
	log.Printf("Registered plugin: %s", p.Name())
	return nil
}

// GetPlugin returns a plugin by name
func (m *Manager) GetPlugin(name string) (Plugin, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p, ok := m.plugins[name]
	return p, ok
}

// ListPlugins returns all registered plugins in registration order
func (m *Manager) ListPlugins() []Plugin {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugins := make([]Plugin, len(m.pluginOrder))
	for i, name := range m.pluginOrder {
		plugins[i] = m.plugins[name]
	}
	return plugins
}

// HandleEvent processes an event through all active plugins
func (m *Manager) HandleEvent(ctx context.Context, event *models.Event) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var wg sync.WaitGroup
	errCh := make(chan error, len(m.plugins))

	for _, p := range m.plugins {
		wg.Add(1)
		go func(p Plugin) {
			defer wg.Done()

			// Process event through plugin
			newEvents, err := p.ProcessEvent(ctx, event)
			if err != nil {
				errCh <- fmt.Errorf("plugin %s error: %v", p.Name(), err)
				return
			}

			// Process any new events generated by the plugin
			for _, newEvent := range newEvents {
				if err := m.HandleEvent(ctx, newEvent); err != nil {
					errCh <- fmt.Errorf("error processing generated event: %v", err)
					return
				}
			}
		}(p)
	}

	// Wait for all plugins to finish
	wg.Wait()
	close(errCh)

	// Collect any errors
	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("plugin errors: %v", errs)
	}

	return nil
}
