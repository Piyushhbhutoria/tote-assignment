package database

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDefaultConfig(t *testing.T) {
	// Test with default values
	cfg := NewDefaultConfig()
	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, 5432, cfg.Port)
	assert.Equal(t, "pos", cfg.User)
	assert.Equal(t, "pos123", cfg.Password)
	assert.Equal(t, "pos_system", cfg.Database)

	// Test with environment variables
	os.Setenv("DB_HOST", "testhost")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("DB_NAME", "testdb")
	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
	}()

	cfg = NewDefaultConfig()
	assert.Equal(t, "testhost", cfg.Host)
	assert.Equal(t, 5432, cfg.Port)
	assert.Equal(t, "testuser", cfg.User)
	assert.Equal(t, "testpass", cfg.Password)
	assert.Equal(t, "testdb", cfg.Database)
}

func TestDatabaseConnection(t *testing.T) {
	// Create test database connection
	cfg := NewDefaultConfig()
	db, err := New(context.Background(), cfg)
	assert.NoError(t, err)
	defer db.Close()

	// Test connection by pinging database
	err = db.Pool().Ping(context.Background())
	assert.NoError(t, err)
}

func TestInitSchema(t *testing.T) {
	// Create test database connection
	cfg := NewDefaultConfig()
	db, err := New(context.Background(), cfg)
	assert.NoError(t, err)
	defer db.Close()

	// Initialize schema
	err = db.InitSchema(context.Background())
	assert.NoError(t, err)

	// Verify tables were created
	var tableCount int
	err = db.Pool().QueryRow(context.Background(), `
		SELECT COUNT(*) 
		FROM information_schema.tables 
		WHERE table_schema = 'public'
	`).Scan(&tableCount)
	assert.NoError(t, err)
	assert.Greater(t, tableCount, 0)

	// Verify specific tables exist
	tables := []string{
		"plugins",
		"employees",
		"employee_sessions",
		"customers",
		"items",
		"baskets",
		"basket_items",
		"fraud_alerts",
		"item_recommendations",
	}

	for _, table := range tables {
		var exists bool
		err = db.Pool().QueryRow(context.Background(), `
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = 'public' 
				AND table_name = $1
			)
		`, table).Scan(&exists)
		assert.NoError(t, err)
		assert.True(t, exists, "Table %s should exist", table)
	}

	// Verify indexes were created
	indexes := []struct {
		table string
		name  string
	}{
		{"employee_sessions", "idx_employee_sessions_employee_id"},
		{"basket_items", "idx_basket_items_basket_id"},
		{"fraud_alerts", "idx_fraud_alerts_basket_id"},
		{"item_recommendations", "idx_item_recommendations_source_item"},
	}

	for _, idx := range indexes {
		var exists bool
		err = db.Pool().QueryRow(context.Background(), `
			SELECT EXISTS (
				SELECT FROM pg_indexes
				WHERE schemaname = 'public'
				AND tablename = $1
				AND indexname = $2
			)
		`, idx.table, idx.name).Scan(&exists)
		assert.NoError(t, err)
		assert.True(t, exists, "Index %s on table %s should exist", idx.name, idx.table)
	}
}

func TestGetEnvOrDefault(t *testing.T) {
	// Test with environment variable set
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")
	assert.Equal(t, "test_value", getEnvOrDefault("TEST_VAR", "default"))

	// Test with environment variable not set
	assert.Equal(t, "default", getEnvOrDefault("NON_EXISTENT_VAR", "default"))
}
