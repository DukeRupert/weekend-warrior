package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for our application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Supabase SupabaseConfig
	Redis    RedisConfig
}

type ServerConfig struct {
	Port         string
	Environment  string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type SupabaseConfig struct {
	Url      string
	Anon_key string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// LoadConfig loads configuration from environment variables
func LoadConfig(envFile string) (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(envFile); err != nil {
		// Only return error if file exists but couldn't be loaded
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("error loading env file: %w", err)
		}
	}

	config := &Config{}

	// Load server configuration
	config.Server = ServerConfig{
		Port:         getEnv("SERVER_PORT", "3000"),
		Environment:  getEnv("ENVIRONMENT", "development"),
		ReadTimeout:  getDurationEnv("SERVER_READ_TIMEOUT", 10*time.Second),
		WriteTimeout: getDurationEnv("SERVER_WRITE_TIMEOUT", 10*time.Second),
	}

	// Load database configuration
	config.Database = DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", ""),
		Name:     getEnv("DB_NAME", ""),
		SSLMode:  getEnv("DB_SSL_MODE", "disable"),
	}

	// Load Supabase configuration
	config.Supabase = SupabaseConfig{
		Url:      getEnv("SUPABASE_URL", ""),
		Anon_key: getEnv("SUPABASE_ANON_KEY", ""),
	}

	// Load Redis configuration
	config.Redis = RedisConfig{
		Host:     getEnv("REDIS_HOST", "localhost"),
		Port:     getEnv("REDIS_PORT", "6379"),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       getIntEnv("REDIS_DB", 0),
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation error: %w", err)
	}

	return config, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("database user is required")
	}
	return nil
}

// GetDatabaseURL returns the formatted database connection string
func (c *Config) GetDatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.Name,
		c.Database.SSLMode,
	)
}

// Helper functions for getting environment variables with defaults
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
