package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	RabbitMQ RabbitMQConfig
	SMTP     SMTPConfig
	API      APIConfig
	TCP      TCPConfig
}

type RabbitMQConfig struct {
	Host     string
	Port     string
	User     string
	Password string
}

type SMTPConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
}

type APIConfig struct {
	Port string
}

type TCPConfig struct {
	Port       string
	AuthSecret string
}

// LoadConfig loads the configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	godotenv.Load()

	// SMTP Configuration
	smtpPort, err := strconv.Atoi(getEnvWithDefault("SMTP_PORT", "587"))
	if err != nil {
		return nil, fmt.Errorf("invalid SMTP_PORT: %w", err)
	}

	config := &Config{
		SMTP: SMTPConfig{
			Host:     getEnvWithDefault("SMTP_HOST", "smtp.gmail.com"),
			Port:     smtpPort,
			User:     os.Getenv("SMTP_USER"),
			Password: os.Getenv("SMTP_PASSWORD"),
			From:     os.Getenv("SMTP_FROM"),
		},
		RabbitMQ: RabbitMQConfig{
			Host:     getEnvWithDefault("RABBITMQ_HOST", "localhost"),
			Port:     getEnvWithDefault("RABBITMQ_PORT", "5672"),
			User:     getEnvWithDefault("RABBITMQ_USER", "admin"),
			Password: getEnvWithDefault("RABBITMQ_PASSWORD", "admin"),
		},
		API: APIConfig{
			Port: getEnvWithDefault("API_PORT", "8080"),
		},
		TCP: TCPConfig{
			Port:       getEnvWithDefault("TCP_PORT", "9000"),
			AuthSecret: os.Getenv("TCP_AUTH_SECRET"),
		},
	}

	// Validate required environment variables
	if err := config.validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// validate checks if all required environment variables are set
func (c *Config) validate() error {
	missingVars := []string{}

	// Check required SMTP variables
	if c.SMTP.User == "" {
		missingVars = append(missingVars, "SMTP_USER")
	}
	if c.SMTP.Password == "" {
		missingVars = append(missingVars, "SMTP_PASSWORD")
	}
	if c.SMTP.From == "" {
		missingVars = append(missingVars, "SMTP_FROM")
	}

	// Check required TCP variables
	if c.TCP.AuthSecret == "" {
		missingVars = append(missingVars, "TCP_AUTH_SECRET")
	}

	if len(missingVars) > 0 {
		return fmt.Errorf("missing required environment variables: %v", missingVars)
	}

	return nil
}

// getEnvWithDefault returns the environment variable value or the default if not set
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
} 