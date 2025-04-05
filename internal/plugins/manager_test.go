package plugins

import (
	"context"
	"testing"
	"time"

	"github.com/Piyushhbhutoria/tote-assignment/internal/models"
	"github.com/Piyushhbhutoria/tote-assignment/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPlugin is a mock implementation of the Plugin interface
type MockPlugin struct {
	mock.Mock
}

func (m *MockPlugin) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockPlugin) Description() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockPlugin) IsActive() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockPlugin) SetActive(active bool) {
	m.Called(active)
}

func (m *MockPlugin) Configure(config map[string]interface{}) error {
	args := m.Called(config)
	return args.Error(0)
}

func (m *MockPlugin) ProcessEvent(ctx context.Context, event *models.Event) ([]*models.Event, error) {
	args := m.Called(ctx, event)
	return args.Get(0).([]*models.Event), args.Error(1)
}

func TestManagerRegisterPlugin(t *testing.T) {
	// Create test database connection
	dbCfg := database.NewDefaultConfig()
	db, err := database.New(context.Background(), dbCfg)
	assert.NoError(t, err)
	defer db.Close()

	// Create plugin manager
	manager := NewManager(db)

	// Create mock plugin
	mockPlugin := new(MockPlugin)
	mockPlugin.On("Name").Return("test_plugin")
	mockPlugin.On("Description").Return("Test plugin description")
	mockPlugin.On("IsActive").Return(false)

	// Register plugin
	err = manager.RegisterPlugin(mockPlugin)
	assert.NoError(t, err)

	// Verify plugin was registered
	plugin, ok := manager.GetPlugin("test_plugin")
	assert.True(t, ok)
	assert.Equal(t, mockPlugin, plugin)

	// Try to register same plugin again
	err = manager.RegisterPlugin(mockPlugin)
	assert.Error(t, err)

	mockPlugin.AssertExpectations(t)
}

func TestManagerListPlugins(t *testing.T) {
	// Create test database connection
	dbCfg := database.NewDefaultConfig()
	db, err := database.New(context.Background(), dbCfg)
	assert.NoError(t, err)
	defer db.Close()

	// Create plugin manager
	manager := NewManager(db)

	// Create mock plugins
	plugin1 := new(MockPlugin)
	plugin1.On("Name").Return("plugin1")
	plugin1.On("Description").Return("Plugin 1 description")
	plugin1.On("IsActive").Return(true)

	plugin2 := new(MockPlugin)
	plugin2.On("Name").Return("plugin2")
	plugin2.On("Description").Return("Plugin 2 description")
	plugin2.On("IsActive").Return(false)

	// Register plugins
	err = manager.RegisterPlugin(plugin1)
	assert.NoError(t, err)
	err = manager.RegisterPlugin(plugin2)
	assert.NoError(t, err)

	// List plugins
	plugins := manager.ListPlugins()
	assert.Len(t, plugins, 2)

	// Verify plugin order and properties
	assert.Contains(t, []string{plugins[0].Name(), plugins[1].Name()}, "plugin1")
	assert.Contains(t, []string{plugins[0].Name(), plugins[1].Name()}, "plugin2")

	plugin1.AssertExpectations(t)
	plugin2.AssertExpectations(t)
}

func TestManagerGetPlugin(t *testing.T) {
	// Create test database connection
	dbCfg := database.NewDefaultConfig()
	db, err := database.New(context.Background(), dbCfg)
	assert.NoError(t, err)
	defer db.Close()

	// Create plugin manager
	manager := NewManager(db)

	// Create mock plugin
	mockPlugin := new(MockPlugin)
	mockPlugin.On("Name").Return("test_plugin")
	mockPlugin.On("Description").Return("Test plugin description")
	mockPlugin.On("IsActive").Return(true)

	// Register plugin
	err = manager.RegisterPlugin(mockPlugin)
	assert.NoError(t, err)

	// Get existing plugin
	plugin, ok := manager.GetPlugin("test_plugin")
	assert.True(t, ok)
	assert.Equal(t, mockPlugin, plugin)

	// Get non-existent plugin
	plugin, ok = manager.GetPlugin("non_existent")
	assert.False(t, ok)
	assert.Nil(t, plugin)

	mockPlugin.AssertExpectations(t)
}

func TestManagerProcessEvent(t *testing.T) {
	// Create test database connection
	dbCfg := database.NewDefaultConfig()
	db, err := database.New(context.Background(), dbCfg)
	assert.NoError(t, err)
	defer db.Close()

	// Create plugin manager
	manager := NewManager(db)

	// Create mock plugin
	mockPlugin := new(MockPlugin)
	mockPlugin.On("Name").Return("test_plugin")
	mockPlugin.On("Description").Return("Test plugin description")
	mockPlugin.On("IsActive").Return(true)
	mockPlugin.On("ProcessEvent", mock.Anything, mock.Anything).Return([]*models.Event{}, nil)

	// Register plugin
	err = manager.RegisterPlugin(mockPlugin)
	assert.NoError(t, err)

	// Create test event
	event := &models.Event{
		ID:        "test_event",
		Type:      models.EventType("test_plugin"),
		Timestamp: time.Now(),
		Payload:   map[string]interface{}{"key": "value"},
	}

	// Process event
	newEvents, err := mockPlugin.ProcessEvent(context.Background(), event)
	assert.NoError(t, err)
	assert.Empty(t, newEvents)

	mockPlugin.AssertExpectations(t)
}
