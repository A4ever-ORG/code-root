package services

import (
	"fmt"
	"log"
	"telegram-store-hub/internal/models"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

// BotManagerService manages automatic sub-bot creation and management
type BotManagerService struct {
	bot *tgbotapi.BotAPI
	db  *gorm.DB
}

// SubBotConfig contains configuration for creating sub-bots
type SubBotConfig struct {
	StoreID     uint   `json:"store_id"`
	StoreName   string `json:"store_name"`
	Description string `json:"description"`
	PlanType    string `json:"plan_type"`
	OwnerID     uint   `json:"owner_id"`
}

// BotStatus represents the status of a bot
type BotStatus string

const (
	BotStatusCreating BotStatus = "creating"
	BotStatusActive   BotStatus = "active"
	BotStatusInactive BotStatus = "inactive"
	BotStatusError    BotStatus = "error"
)

// NewBotManagerService creates a new bot manager service
func NewBotManagerService(bot *tgbotapi.BotAPI, db *gorm.DB) *BotManagerService {
	return &BotManagerService{
		bot: bot,
		db:  db,
	}
}

// CreateSubBot creates a new sub-bot for a store
func (b *BotManagerService) CreateSubBot(storeID uint) error {
	// Get store information
	var store models.Store
	if err := b.db.Preload("Owner").First(&store, storeID).Error; err != nil {
		return fmt.Errorf("failed to get store: %w", err)
	}

	log.Printf("Creating sub-bot for store: %s (ID: %d)", store.Name, store.ID)

	// Update store status to indicate bot creation is in progress
	if err := b.db.Model(&store).Updates(map[string]interface{}{
		"bot_status": BotStatusCreating,
		"updated_at": time.Now(),
	}).Error; err != nil {
		return fmt.Errorf("failed to update store status: %w", err)
	}

	// Start async bot creation process
	go b.createSubBotAsync(&store)

	return nil
}

// createSubBotAsync creates sub-bot asynchronously
func (b *BotManagerService) createSubBotAsync(store *models.Store) {
	// Generate unique bot username
	botUsername := b.generateBotUsername(store.Name, store.ID)
	
	// Simulate bot creation process (in real implementation, this would involve BotFather API)
	log.Printf("Simulating bot creation for store: %s", store.Name)
	time.Sleep(2 * time.Second) // Simulate creation delay

	// Generate fake bot token for demonstration
	botToken := b.generateFakeBotToken(store.ID)

	// Create bot configuration
	config := &SubBotConfig{
		StoreID:     store.ID,
		StoreName:   store.Name,
		Description: store.Description,
		PlanType:    string(store.PlanType),
		OwnerID:     store.OwnerID,
	}

	// Update store with bot information
	updates := map[string]interface{}{
		"bot_username": botUsername,
		"bot_token":    botToken, // In real implementation, encrypt this
		"bot_status":   BotStatusActive,
		"updated_at":   time.Now(),
	}

	if err := b.db.Model(store).Updates(updates).Error; err != nil {
		log.Printf("Failed to update store with bot info: %v", err)
		b.handleBotCreationError(store.ID, err)
		return
	}

	// Initialize bot settings
	if err := b.initializeBotSettings(store.ID, config); err != nil {
		log.Printf("Failed to initialize bot settings: %v", err)
		b.handleBotCreationError(store.ID, err)
		return
	}

	// Send notification to store owner
	b.notifyBotCreated(store.Owner.TelegramID, store, botUsername)

	log.Printf("Sub-bot created successfully for store %s: @%s", store.Name, botUsername)
}

// generateBotUsername generates a unique bot username
func (b *BotManagerService) generateBotUsername(storeName string, storeID uint) string {
	// Clean store name for username (remove spaces, special chars)
	cleanName := b.cleanStringForUsername(storeName)
	
	// Ensure uniqueness by adding store ID
	return fmt.Sprintf("%s_%d_bot", cleanName, storeID)
}

// generateFakeBotToken generates a fake bot token for demonstration
func (b *BotManagerService) generateFakeBotToken(storeID uint) string {
	// In real implementation, this would be obtained from BotFather
	return fmt.Sprintf("fake_token_%d_%d", storeID, time.Now().Unix())
}

// cleanStringForUsername cleans a string to make it suitable for username
func (b *BotManagerService) cleanStringForUsername(s string) string {
	// Replace spaces and special characters with underscores
	result := ""
	for _, char := range s {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') {
			result += string(char)
		} else if char == ' ' || char == '-' {
			result += "_"
		}
	}
	
	// Limit length to 20 characters
	if len(result) > 20 {
		result = result[:20]
	}
	
	return result
}

// initializeBotSettings initializes settings for the new bot
func (b *BotManagerService) initializeBotSettings(storeID uint, config *SubBotConfig) error {
	// In a real implementation, this would:
	// 1. Set bot description and about text
	// 2. Configure bot commands
	// 3. Set up webhooks
	// 4. Initialize bot database tables
	// 5. Configure payment settings
	// 6. Set up welcome messages based on plan type

	log.Printf("Initializing bot settings for store %d", storeID)

	// Simulate initialization delay
	time.Sleep(1 * time.Second)

	// Here you would configure the bot based on the plan type
	switch config.PlanType {
	case "free":
		// Configure free plan features
		log.Printf("Configuring free plan features for store %d", storeID)
	case "pro":
		// Configure pro plan features (welcome message, etc.)
		log.Printf("Configuring pro plan features for store %d", storeID)
	case "vip":
		// Configure VIP plan features (custom features, etc.)
		log.Printf("Configuring VIP plan features for store %d", storeID)
	}

	return nil
}

// handleBotCreationError handles errors during bot creation
func (b *BotManagerService) handleBotCreationError(storeID uint, err error) {
	log.Printf("Bot creation failed for store %d: %v", storeID, err)

	// Update store status to error
	b.db.Model(&models.Store{}).Where("id = ?", storeID).Updates(map[string]interface{}{
		"bot_status": BotStatusError,
		"updated_at": time.Now(),
	})

	// Notify store owner about the error
	var store models.Store
	if err := b.db.Preload("Owner").First(&store, storeID).Error; err == nil {
		b.notifyBotCreationError(store.Owner.TelegramID, &store, err)
	}
}

// notifyBotCreated sends notification when bot is successfully created
func (b *BotManagerService) notifyBotCreated(ownerTelegramID int64, store *models.Store, botUsername string) {
	text := fmt.Sprintf(`🤖 ربات فروشگاه شما آماده شد!

🏪 فروشگاه: %s
🤖 ربات: @%s

✅ ربات شما با موفقیت ایجاد و پیکربندی شد!

🔗 لینک ربات: https://t.me/%s

اکنون مشتریان می‌توانند از طریق این ربات محصولات شما را مشاهده و خریداری کنند.

💡 نکته: ربات فعلاً در حالت آزمایشی است. برای فعال‌سازی کامل، با پشتیبانی تماس بگیرید.`,
		store.Name,
		botUsername,
		botUsername)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("🔗 باز کردن ربات", fmt.Sprintf("https://t.me/%s", botUsername)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏪 پنل مدیریت", "manage_store"),
		),
	)

	msg := tgbotapi.NewMessage(ownerTelegramID, text)
	msg.ReplyMarkup = keyboard

	b.bot.Send(msg)
}

// notifyBotCreationError sends notification when bot creation fails
func (b *BotManagerService) notifyBotCreationError(ownerTelegramID int64, store *models.Store, err error) {
	text := fmt.Sprintf(`❌ خطا در ایجاد ربات فروشگاه

🏪 فروشگاه: %s

متأسفانه در ایجاد ربات فروشگاه شما خطایی رخ داده است.

🔄 ما در حال بررسی و رفع مشکل هستیم.

💬 در صورت ادامه مشکل، با پشتیبانی تماس بگیرید.`,
		store.Name)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 تلاش مجدد", "retry_bot_creation"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💬 پشتیبانی", "support"),
		),
	)

	msg := tgbotapi.NewMessage(ownerTelegramID, text)
	msg.ReplyMarkup = keyboard

	b.bot.Send(msg)
}

// RetryBotCreation retries bot creation for a store
func (b *BotManagerService) RetryBotCreation(storeID uint) error {
	log.Printf("Retrying bot creation for store %d", storeID)
	return b.CreateSubBot(storeID)
}

// DeactivateBot deactivates a store's bot
func (b *BotManagerService) DeactivateBot(storeID uint) error {
	// Update store status
	if err := b.db.Model(&models.Store{}).Where("id = ?", storeID).Updates(map[string]interface{}{
		"bot_status": BotStatusInactive,
		"updated_at": time.Now(),
	}).Error; err != nil {
		return fmt.Errorf("failed to deactivate bot: %w", err)
	}

	log.Printf("Bot deactivated for store %d", storeID)
	return nil
}

// ReactivateBot reactivates a store's bot
func (b *BotManagerService) ReactivateBot(storeID uint) error {
	// Update store status
	if err := b.db.Model(&models.Store{}).Where("id = ?", storeID).Updates(map[string]interface{}{
		"bot_status": BotStatusActive,
		"updated_at": time.Now(),
	}).Error; err != nil {
		return fmt.Errorf("failed to reactivate bot: %w", err)
	}

	log.Printf("Bot reactivated for store %d", storeID)
	return nil
}

// GetBotStatus returns the status of a store's bot
func (b *BotManagerService) GetBotStatus(storeID uint) (BotStatus, error) {
	var store models.Store
	if err := b.db.Select("bot_status").First(&store, storeID).Error; err != nil {
		return "", fmt.Errorf("failed to get store: %w", err)
	}

	return BotStatus(store.BotStatus), nil
}

// GetStoreBots returns all bots with their status
func (b *BotManagerService) GetStoreBots() ([]models.Store, error) {
	var stores []models.Store
	if err := b.db.Preload("Owner").Where("bot_username IS NOT NULL").Find(&stores).Error; err != nil {
		return nil, fmt.Errorf("failed to get store bots: %w", err)
	}

	return stores, nil
}

// UpdateBotConfiguration updates bot configuration based on plan changes
func (b *BotManagerService) UpdateBotConfiguration(storeID uint, newPlanType string) error {
	log.Printf("Updating bot configuration for store %d to plan %s", storeID, newPlanType)

	// In a real implementation, this would:
	// 1. Update bot features based on new plan
	// 2. Enable/disable certain commands
	// 3. Update welcome messages
	// 4. Modify payment settings
	// 5. Update commission rates

	// Simulate configuration update
	time.Sleep(500 * time.Millisecond)

	log.Printf("Bot configuration updated for store %d", storeID)
	return nil
}

// MonitorBots monitors all bots and ensures they're running properly
func (b *BotManagerService) MonitorBots() {
	log.Println("Starting bot monitoring routine...")

	// Get all active bots
	var stores []models.Store
	if err := b.db.Where("bot_status = ?", BotStatusActive).Find(&stores).Error; err != nil {
		log.Printf("Error getting active bots: %v", err)
		return
	}

	for _, store := range stores {
		// In a real implementation, this would:
		// 1. Check if bot is responding
		// 2. Verify bot configuration
		// 3. Check for errors or issues
		// 4. Update bot status if needed

		log.Printf("Monitoring bot for store %s (@%s)", store.Name, store.BotUsername)
	}

	log.Printf("Monitored %d active bots", len(stores))
}

// StartBotMonitoringRoutine starts a routine to monitor bots periodically
func (b *BotManagerService) StartBotMonitoringRoutine() {
	go func() {
		ticker := time.NewTicker(30 * time.Minute) // Monitor every 30 minutes
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				b.MonitorBots()
			}
		}
	}()

	log.Println("Bot monitoring routine started")
}