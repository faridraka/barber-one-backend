package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

var envValues map[string]string

func init() {
	_ = godotenv.Load()
	
	// Load .env values into memory
	if envValues == nil {
		envValues = make(map[string]string)
		// Read .env file directly
		if content, err := os.ReadFile(".env"); err == nil {
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])
					envValues[key] = value
				}
			}
		}
	}
}

type Config struct {
	ServerPort      string
	DatabaseURL     string
	FrontendOrigin  string
	MidtransKey     string
	MidtransEnv     string
	SecretKey       string
}

func LoadConfig() (*Config, error) {
	config := &Config{
		ServerPort:      getEnvValue("SERVER_PORT", "3000"),
		DatabaseURL:     getEnvValue("DATABASE_URL", ""),
		FrontendOrigin:  getEnvValue("FRONTEND_ORIGIN", "http://localhost:5173"),
		MidtransKey:     getEnvValue("MIDTRANS_SERVER_KEY", ""),
		MidtransEnv:     strings.ToLower(getEnvValue("MIDTRANS_ENVIRONMENT", "sandbox")),
		SecretKey:       getEnvValue("SECRET_KEY", "default-secret-key"),
	}

	if config.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}

	if config.MidtransKey == "" {
		fmt.Println("Warning: MIDTRANS_SERVER_KEY is not set")
	}

	return config, nil
}

func getEnvValue(key, fallback string) string {
	// Deployment platforms (including Render) provide configuration through
	// the process environment. Prefer it over values loaded from a local .env.
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	if value, exists := envValues[key]; exists && value != "" {
		return strings.TrimSpace(value)
	}
	return fallback
}
