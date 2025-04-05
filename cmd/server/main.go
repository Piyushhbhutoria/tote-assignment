package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/Piyushhbhutoria/tote-assignment/internal/models"
	"github.com/Piyushhbhutoria/tote-assignment/internal/plugins"
	"github.com/Piyushhbhutoria/tote-assignment/internal/plugins/customer_lookup"
	"github.com/Piyushhbhutoria/tote-assignment/internal/plugins/employee_tracker"
	"github.com/Piyushhbhutoria/tote-assignment/internal/plugins/purchase_recommender"
	"github.com/Piyushhbhutoria/tote-assignment/pkg/database"
	"github.com/Piyushhbhutoria/tote-assignment/pkg/kafka"
	"github.com/gin-gonic/gin"
)

type server struct {
	db          *database.Connection
	pluginMgr   *plugins.Manager
	pluginStats map[string]*models.PluginStats
	statsMutex  sync.RWMutex
	consumer    *kafka.Consumer
}

func main() {
	log.Println("Starting POS System...")

	// Initialize database
	dbCfg := database.NewDefaultConfig()
	db, err := database.New(context.Background(), dbCfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize database schema
	if err := db.InitSchema(context.Background()); err != nil {
		log.Fatalf("Failed to initialize schema: %v", err)
	}

	// Initialize plugin manager
	pluginMgr := plugins.NewManager(db)

	// Create server instance
	srv := &server{
		db:          db,
		pluginMgr:   pluginMgr,
		pluginStats: make(map[string]*models.PluginStats),
	}

	// Register plugins
	if err := srv.registerPlugins(); err != nil {
		log.Fatalf("Failed to register plugins: %v", err)
	}

	// Create Kafka consumer
	kafkaCfg := kafka.NewDefaultConfig()
	consumer, err := kafka.NewConsumer(kafkaCfg, srv.handleEvent)
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}
	srv.consumer = consumer
	defer consumer.Close()

	// Initialize Gin router
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}
		c.Next()
	})

	// API routes
	api := r.Group("/api")
	{
		api.GET("/plugins", srv.handleListPlugins)
		api.PATCH("/plugins/:name/status", srv.handleUpdatePluginStatus)
		api.PATCH("/plugins/:name/config", srv.handleUpdatePluginConfig)
	}

	// Create HTTP server
	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Start background tasks
	var wg sync.WaitGroup

	// Start HTTP server
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("API server listening on %s", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
			cancel()
		}
	}()

	// Start consuming events
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("Starting event consumer...")
		if err := consumer.Start(ctx); err != nil {
			log.Printf("Consumer error: %v", err)
			cancel()
		}
	}()

	// Wait for interrupt signal
	<-stop
	log.Println("Shutting down...")

	// Shutdown HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error shutting down HTTP server: %v", err)
	}

	// Cancel context to stop consumer
	cancel()

	// Wait for all tasks to complete
	wg.Wait()
}

func (s *server) registerPlugins() error {
	// Register employee time tracker plugin
	empTracker := employee_tracker.New(s.db)
	if err := s.pluginMgr.RegisterPlugin(empTracker); err != nil {
		return fmt.Errorf("failed to register employee tracker: %v", err)
	}

	// Register purchase recommender plugin
	purchaseRecommender := purchase_recommender.New(s.db)
	if err := s.pluginMgr.RegisterPlugin(purchaseRecommender); err != nil {
		return fmt.Errorf("failed to register purchase recommender: %v", err)
	}

	// Register customer lookup plugin
	customerLookup := customer_lookup.New(s.db)
	if err := s.pluginMgr.RegisterPlugin(customerLookup); err != nil {
		return fmt.Errorf("failed to register customer lookup: %v", err)
	}

	return nil
}

func (s *server) handleEvent(ctx context.Context, event *models.Event) error {
	// Process event through all plugins
	plugins := s.pluginMgr.ListPlugins()
	for _, plugin := range plugins {
		// Skip inactive plugins
		if !plugin.IsActive() {
			continue
		}

		// Process event through plugin
		fmt.Println("Processing event through plugin", plugin.Name())
		newEvents, err := plugin.ProcessEvent(ctx, event)

		// Update plugin stats
		s.statsMutex.Lock()
		stats, ok := s.pluginStats[plugin.Name()]
		if !ok {
			stats = &models.PluginStats{}
			s.pluginStats[plugin.Name()] = stats
		}
		stats.EventsProcessed++
		stats.LastProcessed = &event.Timestamp
		if err != nil {
			stats.ErrorCount++
		}
		s.statsMutex.Unlock()

		if err != nil {
			log.Printf("Error processing event in plugin %s: %v", plugin.Name(), err)
			continue
		}

		// Handle any new events generated by the plugin
		for _, newEvent := range newEvents {
			if err := s.handleEvent(ctx, newEvent); err != nil {
				log.Printf("Error handling generated event: %v", err)
			}
		}
	}

	return nil
}

func (s *server) handleListPlugins(c *gin.Context) {
	plugins := s.pluginMgr.ListPlugins()
	response := make([]struct {
		Name        string                 `json:"name"`
		Description string                 `json:"description"`
		IsActive    bool                   `json:"isActive"`
		Config      map[string]interface{} `json:"config"`
		Stats       struct {
			EventsProcessed int    `json:"eventsProcessed"`
			LastProcessed   string `json:"lastProcessed,omitempty"`
			ErrorCount      int    `json:"errorCount"`
		} `json:"stats"`
	}, 0, len(plugins))

	for _, p := range plugins {
		resp := struct {
			Name        string                 `json:"name"`
			Description string                 `json:"description"`
			IsActive    bool                   `json:"isActive"`
			Config      map[string]interface{} `json:"config"`
			Stats       struct {
				EventsProcessed int    `json:"eventsProcessed"`
				LastProcessed   string `json:"lastProcessed,omitempty"`
				ErrorCount      int    `json:"errorCount"`
			} `json:"stats"`
		}{
			Name:        p.Name(),
			Description: p.Description(),
			IsActive:    p.IsActive(),
			Config:      make(map[string]interface{}),
		}

		// Get plugin stats
		s.statsMutex.RLock()
		if stats, ok := s.pluginStats[p.Name()]; ok {
			resp.Stats.EventsProcessed = stats.EventsProcessed
			if stats.LastProcessed != nil {
				resp.Stats.LastProcessed = stats.LastProcessed.Format(time.RFC3339)
			}
			resp.Stats.ErrorCount = stats.ErrorCount
		}
		s.statsMutex.RUnlock()

		response = append(response, resp)
	}

	c.JSON(http.StatusOK, response)
}

func (s *server) handleUpdatePluginStatus(c *gin.Context) {
	name := c.Param("name")

	var req struct {
		IsActive bool `json:"isActive"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	plugin, ok := s.pluginMgr.GetPlugin(name)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Plugin not found"})
		return
	}

	plugin.SetActive(req.IsActive)
	fmt.Println("Plugin status updated to", req.IsActive, plugin.Name())
	c.Status(http.StatusOK)
}

func (s *server) handleUpdatePluginConfig(c *gin.Context) {
	name := c.Param("name")

	var req struct {
		Config map[string]interface{} `json:"config"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	plugin, ok := s.pluginMgr.GetPlugin(name)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Plugin not found"})
		return
	}

	if err := plugin.Configure(req.Config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update config: %v", err)})
		return
	}

	c.Status(http.StatusOK)
}
