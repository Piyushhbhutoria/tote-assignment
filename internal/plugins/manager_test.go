package plugins

import (
	"context"
	"testing"

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

func setupTestManager(t *testing.T) *Manager {
	dbCfg := database.NewDefaultConfig()
	db, err := database.New(context.Background(), dbCfg)
	assert.NoError(t, err)
	return NewManager(db)
}

func TestNewManager(t *testing.T) {
	mgr := setupTestManager(t)
	assert.NotNil(t, mgr)
	assert.NotNil(t, mgr.db)
	assert.NotNil(t, mgr.plugins)
	assert.NotNil(t, mgr.pluginOrder)
	assert.Empty(t, mgr.plugins)
	assert.Empty(t, mgr.pluginOrder)
}

func TestManagerRegisterPlugin(t *testing.T) {
	mgr := setupTestManager(t)

	// Create mock plugin
	mockPlugin := new(MockPlugin)
	mockPlugin.On("Name").Return("test_plugin").Maybe()
	mockPlugin.On("Description").Return("Test plugin description").Maybe()
	mockPlugin.On("IsActive").Return(true).Maybe()

	// Test successful registration
	err := mgr.RegisterPlugin(mockPlugin)
	assert.NoError(t, err)
	assert.Len(t, mgr.plugins, 1)
	assert.Len(t, mgr.pluginOrder, 1)

	// Test duplicate registration
	err = mgr.RegisterPlugin(mockPlugin)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")

	mockPlugin.AssertExpectations(t)
}

func TestManagerGetPlugin(t *testing.T) {
	mgr := setupTestManager(t)

	// Create mock plugin
	mockPlugin := new(MockPlugin)
	mockPlugin.On("Name").Return("test_plugin").Maybe()
	mockPlugin.On("Description").Return("Test plugin description").Maybe()
	mockPlugin.On("IsActive").Return(true).Maybe()

	// Register plugin
	err := mgr.RegisterPlugin(mockPlugin)
	assert.NoError(t, err)

	// Test getting existing plugin
	plugin, exists := mgr.GetPlugin("test_plugin")
	assert.True(t, exists)
	assert.NotNil(t, plugin)

	// Test getting non-existent plugin
	plugin, exists = mgr.GetPlugin("non_existent")
	assert.False(t, exists)
	assert.Nil(t, plugin)

	mockPlugin.AssertExpectations(t)
}

func TestManagerListPlugins(t *testing.T) {
	mgr := setupTestManager(t)

	// Create first mock plugin
	plugin1 := new(MockPlugin)
	plugin1.On("Name").Return("plugin1").Maybe()
	plugin1.On("Description").Return("Plugin 1 description").Maybe()
	plugin1.On("IsActive").Return(true).Maybe()

	// Create second mock plugin
	plugin2 := new(MockPlugin)
	plugin2.On("Name").Return("plugin2").Maybe()
	plugin2.On("Description").Return("Plugin 2 description").Maybe()
	plugin2.On("IsActive").Return(false).Maybe()

	// Register plugins
	err := mgr.RegisterPlugin(plugin1)
	assert.NoError(t, err)
	err = mgr.RegisterPlugin(plugin2)
	assert.NoError(t, err)

	// Test listing plugins
	plugins := mgr.ListPlugins()
	assert.Len(t, plugins, 2)
	assert.Equal(t, "plugin1", plugins[0].Name())
	assert.Equal(t, "plugin2", plugins[1].Name())

	plugin1.AssertExpectations(t)
	plugin2.AssertExpectations(t)
}

func TestManagerHandleEvent(t *testing.T) {
	mgr := setupTestManager(t)

	// Create mock plugin
	mockPlugin := new(MockPlugin)
	mockPlugin.On("Name").Return("test_plugin").Maybe()
	mockPlugin.On("Description").Return("Test plugin description").Maybe()
	mockPlugin.On("IsActive").Return(true).Maybe()

	// Setup ProcessEvent expectations
	ctx := context.Background()
	event := &models.Event{
		ID:      "test_event",
		Type:    "test_type",
		Payload: map[string]interface{}{"key": "value"},
	}
	mockPlugin.On("ProcessEvent", ctx, event).Return([]*models.Event{}, nil).Once()

	// Register plugin
	err := mgr.RegisterPlugin(mockPlugin)
	assert.NoError(t, err)

	// Test handling event
	err = mgr.HandleEvent(ctx, event)
	assert.NoError(t, err)

	mockPlugin.AssertExpectations(t)
}

func TestManagerHandleEventWithError(t *testing.T) {
	mgr := setupTestManager(t)

	// Create mock plugin
	mockPlugin := new(MockPlugin)
	mockPlugin.On("Name").Return("test_plugin").Maybe()
	mockPlugin.On("Description").Return("Test plugin description").Maybe()
	mockPlugin.On("IsActive").Return(true).Maybe()

	// Setup ProcessEvent expectations with error
	ctx := context.Background()
	event := &models.Event{
		ID:      "test_event",
		Type:    "test_type",
		Payload: map[string]interface{}{"key": "value"},
	}
	mockPlugin.On("ProcessEvent", ctx, event).Return([]*models.Event{}, assert.AnError).Once()

	// Register plugin
	err := mgr.RegisterPlugin(mockPlugin)
	assert.NoError(t, err)

	// Test handling event with error
	err = mgr.HandleEvent(ctx, event)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plugin errors")

	mockPlugin.AssertExpectations(t)
}

func TestManagerHandleEventWithGeneratedEvents(t *testing.T) {
	mgr := setupTestManager(t)

	// Create mock plugin
	mockPlugin := new(MockPlugin)
	mockPlugin.On("Name").Return("test_plugin").Maybe()
	mockPlugin.On("Description").Return("Test plugin description").Maybe()
	mockPlugin.On("IsActive").Return(true).Maybe()

	// Setup ProcessEvent expectations with generated events
	ctx := context.Background()
	event := &models.Event{
		ID:      "test_event",
		Type:    "test_type",
		Payload: map[string]interface{}{"key": "value"},
	}
	generatedEvent := &models.Event{
		ID:      "generated_event",
		Type:    "generated_type",
		Payload: map[string]interface{}{"generated": "value"},
	}
	mockPlugin.On("ProcessEvent", ctx, event).Return([]*models.Event{generatedEvent}, nil).Once()
	mockPlugin.On("ProcessEvent", ctx, generatedEvent).Return([]*models.Event{}, nil).Once()

	// Register plugin
	err := mgr.RegisterPlugin(mockPlugin)
	assert.NoError(t, err)

	// Test handling event with generated events
	err = mgr.HandleEvent(ctx, event)
	assert.NoError(t, err)

	mockPlugin.AssertExpectations(t)
}
