package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	// Bot Configuration
	MotherBotToken    string `json:"mother_bot_token"`
	Debug             bool   `json:"debug"`
	
	// Database Configuration
	DatabaseURL       string `json:"database_url"`
	
	// Admin Configuration
	AdminChatID       int64  `json:"admin_chat_id"`
	
	// Channel Configuration (for forced join)
	RequiredChannelID       string `json:"required_channel_id"`
	RequiredChannelUsername string `json:"required_channel_username"`
	
	// Payment Configuration
	PaymentCardNumber string `json:"payment_card_number"`
	PaymentCardHolder string `json:"payment_card_holder"`
	
	// Plan Pricing (in Toman)
	FreePlanPrice int64 `json:"free_plan_price"`
	ProPlanPrice  int64 `json:"pro_plan_price"`
	VIPPlanPrice  int64 `json:"vip_plan_price"`
	
	// Commission Rates (percentage)
	FreePlanCommission int `json:"free_plan_commission"`
	ProPlanCommission  int `json:"pro_plan_commission"`
	VIPPlanCommission  int `json:"vip_plan_commission"`
	
	// Feature Limits
	FreePlanProductLimit int `json:"free_plan_product_limit"`
	ProPlanProductLimit  int `json:"pro_plan_product_limit"`
	VIPPlanProductLimit  int `json:"vip_plan_product_limit"` // -1 for unlimited
	
	// Subscription Reminder Settings
	ReminderDaysBeforeExpiry []int `json:"reminder_days_before_expiry"`
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Try to load .env file (ignore error if file doesn't exist)
	_ = godotenv.Load()
	
	cfg := &Config{
		// Default values
		Debug:                    getEnvBool("DEBUG", false),
		FreePlanPrice:           0,
		ProPlanPrice:            50000,  // 50,000 Toman
		VIPPlanPrice:            150000, // 150,000 Toman
		FreePlanCommission:      5,      // 5%
		ProPlanCommission:       5,      // 5%
		VIPPlanCommission:       0,      // 0% for VIP
		FreePlanProductLimit:    10,     // 10 products
		ProPlanProductLimit:     200,    // 200 products
		VIPPlanProductLimit:     -1,     // Unlimited
		ReminderDaysBeforeExpiry: []int{7, 3, 1}, // Remind 7, 3, and 1 days before expiry
	}
	
	// Required environment variables
	cfg.MotherBotToken = os.Getenv("BOT_TOKEN")
	if cfg.MotherBotToken == "" {
		return nil, fmt.Errorf("BOT_TOKEN environment variable is required")
	}
	
	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}
	
	// Admin chat ID
	adminChatIDStr := os.Getenv("ADMIN_CHAT_ID")
	if adminChatIDStr == "" {
		return nil, fmt.Errorf("ADMIN_CHAT_ID environment variable is required")
	}
	
	adminChatID, err := strconv.ParseInt(adminChatIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid ADMIN_CHAT_ID: %v", err)
	}
	cfg.AdminChatID = adminChatID
	
	// Optional configuration
	cfg.RequiredChannelID = os.Getenv("FORCE_JOIN_CHANNEL_ID")
	cfg.RequiredChannelUsername = os.Getenv("FORCE_JOIN_CHANNEL_USERNAME")
	cfg.PaymentCardNumber = getEnv("PAYMENT_CARD_NUMBER", "1234-5678-9012-3456")
	cfg.PaymentCardHolder = getEnv("PAYMENT_CARD_HOLDER", "فروشگاه CodeRoot")
	
	// Override pricing if environment variables are set
	if price := getEnvInt64("FREE_PLAN_PRICE", -1); price != -1 {
		cfg.FreePlanPrice = price
	}
	if price := getEnvInt64("PRO_PLAN_PRICE", -1); price != -1 {
		cfg.ProPlanPrice = price
	}
	if price := getEnvInt64("VIP_PLAN_PRICE", -1); price != -1 {
		cfg.VIPPlanPrice = price
	}
	
	// Override commission rates if environment variables are set
	if rate := getEnvInt("FREE_PLAN_COMMISSION", -1); rate != -1 {
		cfg.FreePlanCommission = rate
	}
	if rate := getEnvInt("PRO_PLAN_COMMISSION", -1); rate != -1 {
		cfg.ProPlanCommission = rate
	}
	if rate := getEnvInt("VIP_PLAN_COMMISSION", -1); rate != -1 {
		cfg.VIPPlanCommission = rate
	}
	
	return cfg, nil
}

// Helper functions for environment variable parsing

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if valueStr := os.Getenv(key); valueStr != "" {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if valueStr := os.Getenv(key); valueStr != "" {
		if value, err := strconv.ParseInt(valueStr, 10, 64); err == nil {
			return value
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if valueStr := os.Getenv(key); valueStr != "" {
		if value, err := strconv.ParseBool(valueStr); err == nil {
			return value
		}
	}
	return defaultValue
}

// GetPlanLimit returns the product limit for a given plan
func (c *Config) GetPlanLimit(planType string) int {
	switch planType {
	case "free":
		return c.FreePlanProductLimit
	case "pro":
		return c.ProPlanProductLimit
	case "vip":
		return c.VIPPlanProductLimit
	default:
		return 0
	}
}

// GetPlanPrice returns the price for a given plan
func (c *Config) GetPlanPrice(planType string) int64 {
	switch planType {
	case "free":
		return c.FreePlanPrice
	case "pro":
		return c.ProPlanPrice
	case "vip":
		return c.VIPPlanPrice
	default:
		return 0
	}
}

// GetPlanCommission returns the commission rate for a given plan
func (c *Config) GetPlanCommission(planType string) int {
	switch planType {
	case "free":
		return c.FreePlanCommission
	case "pro":
		return c.ProPlanCommission
	case "vip":
		return c.VIPPlanCommission
	default:
		return 0
	}
}

// IsChannelJoinRequired returns true if channel join is required
func (c *Config) IsChannelJoinRequired() bool {
	return c.RequiredChannelID != "" && c.RequiredChannelUsername != ""
}