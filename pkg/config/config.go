// pkg/config/config.go
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App struct {
		Name string
		Env  string
		Port string
	}
	DB struct {
		Host     string
		Port     int
		User     string
		Password string
		Name     string
		URL      string
	}
	Auth struct {
		AdminUsername string
		AdminPassword string
		JWTSecret     string
		JWTExpiry     time.Duration
	}
	Kafka struct {
		Brokers []string
	}
	Milvus struct {
		Address string
	}
	AI struct {
		GeminiAPIKey         string
		OpenAIAPIKey         string
		ChatEmbeddingProfile string
	}
}

var Current *Config

func Load() (*Config, error) {
	// Try to load .env file, ignore error if it doesn't exist
	_ = godotenv.Load()

	cfg := &Config{}
	cfg.App.Name = getEnv("APP_NAME", "Dan AI")
	cfg.App.Env = getEnv("APP_ENV", "development")
	cfg.App.Port = getEnv("APP_PORT", "8080")

	cfg.DB.Host = getEnv("DB_HOST", "localhost")
	cfg.DB.User = getEnv("DB_USER", "postgres")
	cfg.DB.Password = getEnv("DB_PASSWORD", "postgres")
	cfg.DB.Name = getEnv("DB_NAME", "dan_ai")
	cfg.DB.URL = getEnv("DATABASE_URL", "")

	dbPortStr := getEnv("DB_PORT", "5432")
	dbPort, err := strconv.Atoi(dbPortStr)
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}
	cfg.DB.Port = dbPort

	// Auth config
	cfg.Auth.AdminUsername = getEnv("ADMIN_USERNAME", "admin")
	cfg.Auth.AdminPassword = getEnv("ADMIN_PASSWORD", "admin")
	cfg.Auth.JWTSecret = getEnv("JWT_SECRET", "")
	jwtExpireStr := getEnv("JWT_EXPIRE", "24h")
	jwtExpiry, err := time.ParseDuration(jwtExpireStr)
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_EXPIRE: %w", err)
	}
	cfg.Auth.JWTExpiry = jwtExpiry

	// Kafka config
	cfg.Kafka.Brokers = []string{getEnv("KAFKA_BROKER", "localhost:9092")}

	// Milvus config
	cfg.Milvus.Address = getEnv("MILVUS_ADDRESS", "localhost:19530")

	// AI config
	cfg.AI.GeminiAPIKey = getEnv("GEMINI_API_KEY", "")
	cfg.AI.OpenAIAPIKey = getEnv("OPENAI_API_KEY", "")
	cfg.AI.ChatEmbeddingProfile = getEnv("CHAT_EMBEDDING_PROFILE", "e5")

	Current = cfg
	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}
