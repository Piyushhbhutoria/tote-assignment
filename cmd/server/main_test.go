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

// setupTestServer creates a new test server with mock dependencies
func setupTestServer(t *testing.T) (*server, *gin.Engine) {
	gin.SetMode(gin.TestMode)

	// Create test database connection
	dbCfg := database.NewDefaultConfig()
	db, err := database.New(context.Background(), dbCfg)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Initialize plugin manager
	pluginMgr := plugins.NewManager(db)

	// Create server instance
	srv := &server{
		db:          db,
		pluginMgr:   pluginMgr,
		pluginStats: make(map[string]*models.PluginStats),
	}

	// Setup router
	r := gin.Default()
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
	mockPlugin.On("Name").Return("test_plugin")
	mockPlugin.On("Description").Return("Test plugin description")
	mockPlugin.On("IsActive").Return(true)

	// Register mock plugin
	err := srv.pluginMgr.RegisterPlugin(mockPlugin)
	assert.NoError(t, err)

	// Add some stats
	srv.pluginStats["test_plugin"] = &models.PluginStats{
		EventsProcessed: 10,
		LastProcessed:   &time.Time{},
		ErrorCount:      2,
	}

	// Create test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/plugins", nil)
	r.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response []map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 1)

	plugin := response[0]
	assert.Equal(t, "test_plugin", plugin["name"])
	assert.Equal(t, "Test plugin description", plugin["description"])
	assert.Equal(t, true, plugin["isActive"])

	stats := plugin["stats"].(map[string]interface{})
	assert.Equal(t, float64(10), stats["eventsProcessed"])
	assert.Equal(t, float64(2), stats["errorCount"])
}

func TestHandleUpdatePluginStatus(t *testing.T) {
	srv, r := setupTestServer(t)

	// Create mock plugin
	mockPlugin := new(MockPlugin)
	mockPlugin.On("Name").Return("test_plugin")
	mockPlugin.On("SetActive", true).Once()

	// Register mock plugin
	err := srv.pluginMgr.RegisterPlugin(mockPlugin)
	assert.NoError(t, err)

	// Create test request
	body := map[string]interface{}{
		"isActive": true,
	}
	jsonBody, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/api/plugins/test_plugin/status", bytes.NewBuffer(jsonBody))
	r.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	mockPlugin.AssertExpectations(t)
}

func TestHandleUpdatePluginConfig(t *testing.T) {
	srv, r := setupTestServer(t)

	// Create mock plugin
	mockPlugin := new(MockPlugin)
	mockPlugin.On("Name").Return("test_plugin")
	mockPlugin.On("Configure", mock.Anything).Return(nil)

	// Register mock plugin
	err := srv.pluginMgr.RegisterPlugin(mockPlugin)
	assert.NoError(t, err)

	// Create test request
	config := map[string]interface{}{
		"key": "value",
	}
	body := map[string]interface{}{
		"config": config,
	}
	jsonBody, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/api/plugins/test_plugin/config", bytes.NewBuffer(jsonBody))
	r.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	mockPlugin.AssertExpectations(t)
}

func TestHandleEvent(t *testing.T) {
	srv, _ := setupTestServer(t)

	// Create mock plugin
	mockPlugin := new(MockPlugin)
	mockPlugin.On("Name").Return("test_plugin")
	mockPlugin.On("ProcessEvent", mock.Anything, mock.Anything).Return([]*models.Event{}, nil)

	// Register mock plugin
	err := srv.pluginMgr.RegisterPlugin(mockPlugin)
	assert.NoError(t, err)

	// Create test event
	event := &models.Event{
		ID:        "test_event",
		Type:      models.EventType("test_plugin"),
		Timestamp: time.Now(),
		Payload:   map[string]interface{}{"key": "value"},
	}

	// Process event
	err = srv.handleEvent(context.Background(), event)
	assert.NoError(t, err)

	// Verify stats were updated
	stats := srv.pluginStats["test_plugin"]
	assert.NotNil(t, stats)
	assert.Equal(t, 1, stats.EventsProcessed)
	assert.Equal(t, 0, stats.ErrorCount)
	assert.NotNil(t, stats.LastProcessed)

	mockPlugin.AssertExpectations(t)
}
