package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Piyushhbhutoria/tote-assignment/internal/models"
	"github.com/Piyushhbhutoria/tote-assignment/internal/plugins"
	"github.com/Piyushhbhutoria/tote-assignment/pkg/database"
	"github.com/gin-gonic/gin"
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

func setupTestServer(t *testing.T) (*server, *gin.Engine) {
	// Initialize database
	dbCfg := database.NewDefaultConfig()
	db, err := database.New(context.Background(), dbCfg)
	assert.NoError(t, err)

	// Initialize plugin manager
	pluginMgr := plugins.NewManager(db)

	// Create server instance
	srv := &server{
		db:          db,
		pluginMgr:   pluginMgr,
		pluginStats: make(map[string]*models.PluginStats),
	}

	// Initialize Gin router
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// API routes
	api := r.Group("/api")
	{
		api.GET("/plugins", srv.handleListPlugins)
		api.PATCH("/plugins/:name/status", srv.handleUpdatePluginStatus)
		api.PATCH("/plugins/:name/config", srv.handleUpdatePluginConfig)
	}

	return srv, r
}

func TestHandleListPlugins(t *testing.T) {
	srv, r := setupTestServer(t)

	// Create mock plugin
	mockPlugin := new(MockPlugin)
	mockPlugin.On("Name").Return("test_plugin").Maybe()
	mockPlugin.On("Description").Return("Test plugin description").Maybe()
	mockPlugin.On("IsActive").Return(true).Maybe()

	// Register plugin
	err := srv.pluginMgr.RegisterPlugin(mockPlugin)
	assert.NoError(t, err)

	// Add some stats
	srv.pluginStats["test_plugin"] = &models.PluginStats{
		EventsProcessed: 10,
		LastProcessed:   &time.Time{},
		ErrorCount:      2,
	}

	// Create request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/plugins", nil)
	r.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response []struct {
		Name        string                 `json:"name"`
		Description string                 `json:"description"`
		IsActive    bool                   `json:"isActive"`
		Config      map[string]interface{} `json:"config"`
		Stats       struct {
			EventsProcessed int    `json:"eventsProcessed"`
			LastProcessed   string `json:"lastProcessed,omitempty"`
			ErrorCount      int    `json:"errorCount"`
		} `json:"stats"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 1)
	assert.Equal(t, "test_plugin", response[0].Name)
	assert.Equal(t, "Test plugin description", response[0].Description)
	assert.True(t, response[0].IsActive)
	assert.Equal(t, 10, response[0].Stats.EventsProcessed)
	assert.Equal(t, 2, response[0].Stats.ErrorCount)

	mockPlugin.AssertExpectations(t)
}

func TestHandleUpdatePluginStatus(t *testing.T) {
	srv, r := setupTestServer(t)

	// Create mock plugin
	mockPlugin := new(MockPlugin)
	mockPlugin.On("Name").Return("test_plugin").Maybe()
	mockPlugin.On("Description").Return("Test plugin description").Maybe()
	mockPlugin.On("IsActive").Return(true).Maybe()
	mockPlugin.On("SetActive", false).Once()

	// Register plugin
	err := srv.pluginMgr.RegisterPlugin(mockPlugin)
	assert.NoError(t, err)

	// Create request
	body := bytes.NewBufferString(`{"isActive": false}`)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/api/plugins/test_plugin/status", body)
	r.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	mockPlugin.AssertExpectations(t)
}

func TestHandleUpdatePluginConfig(t *testing.T) {
	srv, r := setupTestServer(t)

	// Create mock plugin
	mockPlugin := new(MockPlugin)
	mockPlugin.On("Name").Return("test_plugin").Maybe()
	mockPlugin.On("Description").Return("Test plugin description").Maybe()
	mockPlugin.On("IsActive").Return(true).Maybe()
	mockPlugin.On("Configure", mock.Anything).Return(nil).Once()

	// Register plugin
	err := srv.pluginMgr.RegisterPlugin(mockPlugin)
	assert.NoError(t, err)

	// Create request
	config := map[string]interface{}{
		"key": "value",
	}
	body := bytes.NewBuffer(nil)
	err = json.NewEncoder(body).Encode(map[string]interface{}{
		"config": config,
	})
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/api/plugins/test_plugin/config", body)
	r.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	mockPlugin.AssertExpectations(t)
}

func TestHandleEvent(t *testing.T) {
	srv, _ := setupTestServer(t)

	// Create mock plugin
	mockPlugin := new(MockPlugin)
	mockPlugin.On("Name").Return("test_plugin").Maybe()
	mockPlugin.On("Description").Return("Test plugin description").Maybe()
	mockPlugin.On("IsActive").Return(true).Maybe()

	// Setup ProcessEvent expectations
	ctx := context.Background()
	event := &models.Event{
		ID:        "test_event",
		Type:      "test_type",
		Payload:   map[string]interface{}{"key": "value"},
		Timestamp: time.Now(),
	}
	mockPlugin.On("ProcessEvent", ctx, event).Return([]*models.Event{}, nil).Once()

	// Register plugin
	err := srv.pluginMgr.RegisterPlugin(mockPlugin)
	assert.NoError(t, err)

	// Test handling event
	err = srv.handleEvent(ctx, event)
	assert.NoError(t, err)

	// Verify stats were updated
	stats := srv.pluginStats["test_plugin"]
	assert.NotNil(t, stats)
	assert.Equal(t, 1, stats.EventsProcessed)
	assert.Equal(t, 0, stats.ErrorCount)
	assert.Equal(t, &event.Timestamp, stats.LastProcessed)

	mockPlugin.AssertExpectations(t)
}

func TestHandleEventWithError(t *testing.T) {
	srv, _ := setupTestServer(t)

	// Create mock plugin
	mockPlugin := new(MockPlugin)
	mockPlugin.On("Name").Return("test_plugin").Maybe()
	mockPlugin.On("Description").Return("Test plugin description").Maybe()
	mockPlugin.On("IsActive").Return(true).Maybe()

	// Setup ProcessEvent expectations with error
	ctx := context.Background()
	event := &models.Event{
		ID:        "test_event",
		Type:      "test_type",
		Payload:   map[string]interface{}{"key": "value"},
		Timestamp: time.Now(),
	}
	mockPlugin.On("ProcessEvent", ctx, event).Return([]*models.Event{}, assert.AnError).Once()

	// Register plugin
	err := srv.pluginMgr.RegisterPlugin(mockPlugin)
	assert.NoError(t, err)

	// Test handling event
	err = srv.handleEvent(ctx, event)
	assert.NoError(t, err) // Main handler doesn't return error, just logs it

	// Verify stats were updated
	stats := srv.pluginStats["test_plugin"]
	assert.NotNil(t, stats)
	assert.Equal(t, 1, stats.EventsProcessed)
	assert.Equal(t, 1, stats.ErrorCount)
	assert.Equal(t, &event.Timestamp, stats.LastProcessed)

	mockPlugin.AssertExpectations(t)
}

func TestHandleEventWithInactivePlugin(t *testing.T) {
	srv, _ := setupTestServer(t)

	// Create mock plugin
	mockPlugin := new(MockPlugin)
	mockPlugin.On("Name").Return("test_plugin").Maybe()
	mockPlugin.On("Description").Return("Test plugin description").Maybe()
	mockPlugin.On("IsActive").Return(false).Maybe() // Plugin is inactive

	// Register plugin
	err := srv.pluginMgr.RegisterPlugin(mockPlugin)
	assert.NoError(t, err)

	// Test handling event
	ctx := context.Background()
	event := &models.Event{
		ID:        "test_event",
		Type:      "test_type",
		Payload:   map[string]interface{}{"key": "value"},
		Timestamp: time.Now(),
	}
	err = srv.handleEvent(ctx, event)
	assert.NoError(t, err)

	// Verify no stats were updated
	stats := srv.pluginStats["test_plugin"]
	assert.Nil(t, stats)

	mockPlugin.AssertExpectations(t)
}
