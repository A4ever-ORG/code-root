package config

import (
	"log"
	"os"
)

// Config holds all application configuration
type Config struct {
	MotherBotToken     string
	DatabaseURL        string 
	RequiredChannelID  string
	Debug              bool
	Port               string
	AdminUserIDs       []int64
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	cfg := &Config{
		MotherBotToken:    getEnv("MOTHER_BOT_TOKEN", ""),
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://postgres:password@localhost/telegram_store_hub?sslmode=disable"),
		RequiredChannelID: getEnv("REQUIRED_CHANNEL_ID", "@your_channel"),
		Debug:            getEnv("DEBUG", "false") == "true",
		Port:             getEnv("PORT", "3000"),
	}
	
	if cfg.MotherBotToken == "" {
		log.Println("⚠️ MOTHER_BOT_TOKEN not set, using placeholder")
		cfg.MotherBotToken = "YOUR_MOTHER_BOT_TOKEN_HERE"
	}
	
	return cfg, nil
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}