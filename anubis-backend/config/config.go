package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server Configuration
	Env  string
	Port string

	// Database Configuration
	Database DatabaseConfig

	// JWT Configuration
	JWT JWTConfig

	// ThreeFold Grid Configuration
	TFGrid TFGridConfig

	// API Configuration
	API APIConfig

	// Logging Configuration
	Logging LoggingConfig

	// CORS Configuration
	CORS CORSConfig
}

type DatabaseConfig struct {
	Type     string
	Host     string
	Port     int
	Name     string
	User     string
	Password string
	SSLMode  string
	SQLitePath string
}

type JWTConfig struct {
	Secret string
	Expiry time.Duration
}

type TFGridConfig struct {
	Network string
}

type APIConfig struct {
	RateLimit int
	Timeout   time.Duration
}

type LoggingConfig struct {
	Level  string
	Format string
}

type CORSConfig struct {
	Origins []string
	Methods []string
	Headers []string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	config := &Config{
		Env:  getEnv("ENV", "development"),
		Port: getEnv("PORT", "8080"),

		Database: DatabaseConfig{
			Type:       getEnv("DB_TYPE", "sqlite"),
			Host:       getEnv("DB_HOST", "localhost"),
			Port:       getEnvAsInt("DB_PORT", 5432),
			Name:       getEnv("DB_NAME", "anubis"),
			User:       getEnv("DB_USER", "anubis"),
			Password:   getEnv("DB_PASSWORD", "password"),
			SSLMode:    getEnv("DB_SSL_MODE", "disable"),
			SQLitePath: getEnv("SQLITE_PATH", "./data/anubis.db"),
		},

		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-this-in-production"),
			Expiry: getEnvAsDuration("JWT_EXPIRY", "24h"),
		},

		TFGrid: TFGridConfig{
			Network: getEnv("TFGRID_NETWORK", "main"),
		},

		API: APIConfig{
			RateLimit: getEnvAsInt("API_RATE_LIMIT", 100),
			Timeout:   getEnvAsDuration("API_TIMEOUT", "30s"),
		},

		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},

		CORS: CORSConfig{
			Origins: getEnvAsSlice("CORS_ORIGINS", []string{"http://localhost:3000", "http://localhost:8080"}),
			Methods: getEnvAsSlice("CORS_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
			Headers: getEnvAsSlice("CORS_HEADERS", []string{"Content-Type", "Authorization"}),
		},
	}

	return config
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue string) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	duration, _ := time.ParseDuration(defaultValue)
	return duration
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// Simple split by comma - could be enhanced for more complex parsing
		var result []string
		for _, item := range strings.Split(value, ",") {
			if trimmed := strings.TrimSpace(item); trimmed != "" {
				result = append(result, trimmed)
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return defaultValue
}
