package database

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
)

// Config holds the database configuration
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

// NewDefaultConfig returns a default configuration
func NewDefaultConfig() *Config {
	return &Config{
		Host:     getEnvOrDefault("DB_HOST", "localhost"),
		Port:     5432,
		User:     getEnvOrDefault("DB_USER", "pos"),
		Password: getEnvOrDefault("DB_PASSWORD", "pos123"),
		Database: getEnvOrDefault("DB_NAME", "pos_system"),
	}
}

// Connection represents a database connection pool
type Connection struct {
	pool *pgxpool.Pool
}

// New creates a new database connection
func New(ctx context.Context, cfg *Config) (*Connection, error) {
	connString := fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)

	pool, err := pgxpool.Connect(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}

	return &Connection{pool: pool}, nil
}

// Close closes the database connection
func (c *Connection) Close() {
	if c.pool != nil {
		c.pool.Close()
	}
}

// Pool returns the underlying connection pool
func (c *Connection) Pool() *pgxpool.Pool {
	return c.pool
}

// InitSchema initializes the database schema
func (c *Connection) InitSchema(ctx context.Context) error {
	schema, err := os.ReadFile("internal/models/schema.sql")
	if err != nil {
		return fmt.Errorf("unable to read schema file: %v", err)
	}

	_, err = c.pool.Exec(ctx, string(schema))
	if err != nil {
		return fmt.Errorf("unable to execute schema: %v", err)
	}

	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
