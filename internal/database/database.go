package database

import (
        "fmt"
        "log"
        "telegram-store-hub/internal/models"
        
        "gorm.io/driver/postgres"
        "gorm.io/gorm"
        "gorm.io/gorm/logger"
)

// Connect connects to the database
func Connect(databaseURL string) (*gorm.DB, error) {
        db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
                Logger: logger.Default.LogMode(logger.Info),
        })
        if err != nil {
                return nil, fmt.Errorf("failed to connect to database: %w", err)
        }
        
        return db, nil
}

// Migrate runs database migrations
func Migrate(db *gorm.DB) error {
        return db.AutoMigrate(
                &models.User{},
                &models.Store{},
                &models.Product{},
                &models.Order{},
                &models.OrderItem{},
                &models.Payment{},
                &models.UserSession{},
        )
}

// SeedDefaultData creates default admin user
func SeedDefaultData(db *gorm.DB, adminChatID int64) error {
        if adminChatID == 0 {
                return fmt.Errorf("admin chat ID not provided")
        }
        
        // Check if admin already exists
        var existingUser models.User
        err := db.Where("telegram_id = ?", adminChatID).First(&existingUser).Error
        if err == nil {
                // Admin already exists
                return nil
        }
        
        // Create admin user
        admin := models.User{
                TelegramID: adminChatID,
                Username:   "admin",
                FirstName:  "System Admin",
                IsAdmin:    true,
        }
        
        return db.Create(&admin).Error
}

func Initialize(databaseURL string) (*gorm.DB, error) {
        db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
                Logger: logger.Default.LogMode(logger.Info),
        })
        if err != nil {
                return nil, err
        }

        // Auto migrate all models
        err = db.AutoMigrate(
                &models.User{},
                &models.Store{},
                &models.Product{},
                &models.Order{},
                &models.OrderItem{},
                &models.Payment{},
                &models.UserSession{},
        )
        if err != nil {
                return nil, err
        }

        log.Println("✅ Database connected and migrated successfully")
        return db, nil
}