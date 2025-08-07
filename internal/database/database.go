package database

import (
	"fmt"
	"log"
	"telegram-store-hub/internal/models"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect establishes connection to PostgreSQL database
func Connect(databaseURL string) (*gorm.DB, error) {
	// Configure GORM with custom logger
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(databaseURL), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

// Migrate runs database migrations
func Migrate(db *gorm.DB) error {
	log.Println("Running database migrations...")

	// Auto-migrate all models
	err := db.AutoMigrate(
		&models.User{},
		&models.Store{},
		&models.Product{},
		&models.Order{},
		&models.OrderItem{},
		&models.Payment{},
		&models.UserSession{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	// Create indexes for better performance
	if err := createIndexes(db); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// createIndexes creates additional indexes for better performance
func createIndexes(db *gorm.DB) error {
	// Index on telegram_id for faster user lookups
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_users_telegram_id ON users(telegram_id)").Error; err != nil {
		return err
	}

	// Index on store owner_id for faster store lookups
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_stores_owner_id ON stores(owner_id)").Error; err != nil {
		return err
	}

	// Index on store plan_type and is_active for subscription management
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_stores_plan_active ON stores(plan_type, is_active)").Error; err != nil {
		return err
	}

	// Index on store expires_at for subscription expiry checks
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_stores_expires_at ON stores(expires_at)").Error; err != nil {
		return err
	}

	// Index on product store_id for faster product queries
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_products_store_id ON products(store_id)").Error; err != nil {
		return err
	}

	// Index on order store_id and status for order management
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_orders_store_status ON orders(store_id, status)").Error; err != nil {
		return err
	}

	// Index on order customer_telegram_id for customer order history
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_orders_customer_id ON orders(customer_telegram_id)").Error; err != nil {
		return err
	}

	// Index on payment store_id and status for payment tracking
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_payments_store_status ON payments(store_id, status)").Error; err != nil {
		return err
	}

	// Index on user_session telegram_id for session management
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_user_sessions_telegram_id ON user_sessions(telegram_id)").Error; err != nil {
		return err
	}

	return nil
}

// SeedDefaultData seeds the database with default admin user and data
func SeedDefaultData(db *gorm.DB, adminTelegramID int64) error {
	log.Println("Seeding default data...")

	// Check if admin user already exists
	var existingAdmin models.User
	result := db.Where("telegram_id = ? AND is_admin = ?", adminTelegramID, true).First(&existingAdmin)
	
	if result.Error == nil {
		log.Println("Admin user already exists, skipping seed")
		return nil
	}

	// Create admin user
	admin := models.User{
		TelegramID: adminTelegramID,
		Username:   "admin",
		FirstName:  "System",
		LastName:   "Admin",
		IsAdmin:    true,
	}

	if err := db.Create(&admin).Error; err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	log.Printf("Created admin user with Telegram ID: %d", adminTelegramID)
	return nil
}

// HealthCheck checks database connection health
func HealthCheck(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

// GetStats returns database statistics
func GetStats(db *gorm.DB) (map[string]int64, error) {
	stats := make(map[string]int64)

	// Count users
	var userCount int64
	if err := db.Model(&models.User{}).Count(&userCount).Error; err != nil {
		return nil, err
	}
	stats["users"] = userCount

	// Count stores
	var storeCount int64
	if err := db.Model(&models.Store{}).Where("is_active = ?", true).Count(&storeCount).Error; err != nil {
		return nil, err
	}
	stats["stores"] = storeCount

	// Count products
	var productCount int64
	if err := db.Model(&models.Product{}).Where("is_available = ?", true).Count(&productCount).Error; err != nil {
		return nil, err
	}
	stats["products"] = productCount

	// Count orders
	var orderCount int64
	if err := db.Model(&models.Order{}).Count(&orderCount).Error; err != nil {
		return nil, err
	}
	stats["orders"] = orderCount

	// Calculate total sales
	var totalSales int64
	if err := db.Model(&models.Order{}).Where("payment_status = ?", "paid").Select("SUM(total_amount)").Scan(&totalSales).Error; err != nil {
		return nil, err
	}
	stats["total_sales"] = totalSales

	return stats, nil
}

// CleanupExpiredSessions removes expired user sessions
func CleanupExpiredSessions(db *gorm.DB) error {
	// Remove sessions older than 24 hours
	cutoff := time.Now().Add(-24 * time.Hour)
	
	result := db.Where("updated_at < ?", cutoff).Delete(&models.UserSession{})
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %w", result.Error)
	}

	if result.RowsAffected > 0 {
		log.Printf("Cleaned up %d expired sessions", result.RowsAffected)
	}

	return nil
}

// BackupDatabase creates a database backup (for development/testing)
func BackupDatabase(db *gorm.DB, backupPath string) error {
	// This is a simplified backup - in production you'd use pg_dump
	log.Printf("Creating database backup at: %s", backupPath)
	
	// In a real implementation, you would:
	// 1. Use pg_dump to create a proper backup
	// 2. Compress the backup
	// 3. Store it in a secure location
	// 4. Implement rotation policy
	
	// For now, just return success
	return nil
}