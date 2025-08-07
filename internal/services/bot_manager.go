package services

import (
	"fmt"
	"log"
	"telegram-store-hub/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

type BotManagerService struct {
	db           *gorm.DB
	activeBots   map[uint]*tgbotapi.BotAPI // store_id -> bot instance
	runningBots  map[uint]chan struct{}    // store_id -> stop channel
}

func NewBotManagerService(db *gorm.DB) *BotManagerService {
	return &BotManagerService{
		db:          db,
		activeBots:  make(map[uint]*tgbotapi.BotAPI),
		runningBots: make(map[uint]chan struct{}),
	}
}

// StartAllBots starts all active store bots
func (s *BotManagerService) StartAllBots() error {
	var stores []models.Store
	err := s.db.Where("is_active = ? AND bot_token != ''", true).Find(&stores).Error
	if err != nil {
		return err
	}
	
	for _, store := range stores {
		if err := s.StartStoreBot(store.ID); err != nil {
			log.Printf("Failed to start bot for store %d: %v", store.ID, err)
		}
	}
	
	return nil
}

// StartStoreBot starts a bot for a specific store
func (s *BotManagerService) StartStoreBot(storeID uint) error {
	var store models.Store
	if err := s.db.First(&store, storeID).Error; err != nil {
		return err
	}
	
	if store.BotToken == "" {
		return fmt.Errorf("bot token not set for store %d", storeID)
	}
	
	// Check if bot is already running
	if _, exists := s.runningBots[storeID]; exists {
		return fmt.Errorf("bot already running for store %d", storeID)
	}
	
	bot, err := tgbotapi.NewBotAPI(store.BotToken)
	if err != nil {
		return err
	}
	
	bot.Debug = false
	s.activeBots[storeID] = bot
	
	// Create stop channel
	stopChan := make(chan struct{})
	s.runningBots[storeID] = stopChan
	
	// Start bot in goroutine
	go s.runStoreBot(storeID, bot, stopChan)
	
	log.Printf("Started bot for store %d (@%s)", storeID, bot.Self.UserName)
	return nil
}

// StopStoreBot stops a bot for a specific store
func (s *BotManagerService) StopStoreBot(storeID uint) error {
	stopChan, exists := s.runningBots[storeID]
	if !exists {
		return fmt.Errorf("no running bot found for store %d", storeID)
	}
	
	// Signal to stop
	close(stopChan)
	
	// Clean up
	delete(s.activeBots, storeID)
	delete(s.runningBots, storeID)
	
	log.Printf("Stopped bot for store %d", storeID)
	return nil
}

// StopAllBots stops all running bots
func (s *BotManagerService) StopAllBots() {
	for storeID := range s.runningBots {
		s.StopStoreBot(storeID)
	}
}

// RestartStoreBot restarts a bot for a specific store
func (s *BotManagerService) RestartStoreBot(storeID uint) error {
	// Stop if running
	if _, exists := s.runningBots[storeID]; exists {
		if err := s.StopStoreBot(storeID); err != nil {
			return err
		}
	}
	
	// Start again
	return s.StartStoreBot(storeID)
}

// UpdateStoreBotToken updates bot token for a store and restarts bot
func (s *BotManagerService) UpdateStoreBotToken(storeID uint, token string) error {
	// Update token in database
	if err := s.db.Model(&models.Store{}).Where("id = ?", storeID).Update("bot_token", token).Error; err != nil {
		return err
	}
	
	// Restart bot with new token
	return s.RestartStoreBot(storeID)
}

// runStoreBot runs the store bot loop
func (s *BotManagerService) runStoreBot(storeID uint, bot *tgbotapi.BotAPI, stopChan chan struct{}) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	
	updates := bot.GetUpdatesChan(u)
	
	for {
		select {
		case <-stopChan:
			bot.StopReceivingUpdates()
			return
		case update := <-updates:
			// Handle store bot updates
			s.handleStoreBotUpdate(storeID, bot, update)
		}
	}
}

// handleStoreBotUpdate handles updates for store bots
func (s *BotManagerService) handleStoreBotUpdate(storeID uint, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	// Get store information
	var store models.Store
	if err := s.db.Preload("Owner").Preload("Products").First(&store, storeID).Error; err != nil {
		log.Printf("Error getting store %d: %v", storeID, err)
		return
	}
	
	// Check if store is active
	if !store.IsActive {
		// Send inactive message if someone tries to use the bot
		if update.Message != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "❌ این فروشگاه در حال حاضر غیرفعال است")
			bot.Send(msg)
		}
		return
	}
	
	if update.Message != nil {
		s.handleStoreBotMessage(store, bot, update.Message)
	} else if update.CallbackQuery != nil {
		s.handleStoreBotCallback(store, bot, update.CallbackQuery)
	}
}

// handleStoreBotMessage handles text messages for store bots
func (s *BotManagerService) handleStoreBotMessage(store models.Store, bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	chatID := message.Chat.ID
	text := message.Text
	
	switch text {
	case "/start":
		s.sendStoreWelcome(store, bot, chatID)
	case "/products", "📦 محصولات":
		s.sendStoreProducts(store, bot, chatID)
	case "/contact", "📞 تماس":
		s.sendStoreContact(store, bot, chatID)
	case "/help", "ℹ️ راهنما":
		s.sendStoreHelp(store, bot, chatID)
	default:
		s.sendStoreWelcome(store, bot, chatID)
	}
}

// handleStoreBotCallback handles callback queries for store bots
func (s *BotManagerService) handleStoreBotCallback(store models.Store, bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
	chatID := callback.Message.Chat.ID
	data := callback.Data
	
	// Answer callback query
	bot.Request(tgbotapi.NewCallback(callback.ID, ""))
	
	// Handle different callback types
	switch {
	case data == "view_products":
		s.sendStoreProducts(store, bot, chatID)
	case data == "contact_store":
		s.sendStoreContact(store, bot, chatID)
	case data == "store_info":
		s.sendStoreInfo(store, bot, chatID)
	}
}

// sendStoreWelcome sends welcome message for store bot
func (s *BotManagerService) sendStoreWelcome(store models.Store, bot *tgbotapi.BotAPI, chatID int64) {
	welcomeText := store.WelcomeMessage
	if welcomeText == "" {
		welcomeText = fmt.Sprintf(`🛍 به فروشگاه %s خوش آمدید!

%s

برای مشاهده محصولات و خرید، از دکمه‌های زیر استفاده کنید:`, store.Name, store.Description)
	}
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📦 مشاهده محصولات", "view_products"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📞 تماس با فروشگاه", "contact_store"),
			tgbotapi.NewInlineKeyboardButtonData("ℹ️ درباره فروشگاه", "store_info"),
		),
	)
	
	msg := tgbotapi.NewMessage(chatID, welcomeText)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

// sendStoreProducts sends list of store products
func (s *BotManagerService) sendStoreProducts(store models.Store, bot *tgbotapi.BotAPI, chatID int64) {
	products, err := s.getActiveStoreProducts(store.ID)
	if err != nil || len(products) == 0 {
		msg := tgbotapi.NewMessage(chatID, "❌ در حال حاضر محصولی موجود نیست")
		bot.Send(msg)
		return
	}
	
	text := fmt.Sprintf("📦 محصولات فروشگاه %s:\n\n", store.Name)
	keyboard := tgbotapi.NewInlineKeyboardMarkup()
	
	for _, product := range products {
		text += fmt.Sprintf("🔹 %s\n", product.Name)
		text += fmt.Sprintf("💰 قیمت: %s تومان\n", s.formatPrice(product.Price))
		text += fmt.Sprintf("📝 %s\n\n", product.Description)
		
		// Add product button
		row := tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("🛒 خرید %s", product.Name),
				fmt.Sprintf("buy_%d", product.ID)),
		)
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
	}
	
	// Add back button
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard,
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 بازگشت", "back_main"),
		))
	
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

// sendStoreContact sends store contact information
func (s *BotManagerService) sendStoreContact(store models.Store, bot *tgbotapi.BotAPI, chatID int64) {
	contactText := store.SupportContact
	if contactText == "" {
		contactText = fmt.Sprintf(`📞 اطلاعات تماس فروشگاه %s:

👤 نام فروشنده: %s %s
📱 تلگرام: @%s

برای سفارش و پشتیبانی با ما در ارتباط باشید.`, 
			store.Name,
			store.Owner.FirstName,
			store.Owner.LastName,
			store.Owner.Username)
	}
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 بازگشت", "back_main"),
		),
	)
	
	msg := tgbotapi.NewMessage(chatID, contactText)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

// sendStoreInfo sends store information
func (s *BotManagerService) sendStoreInfo(store models.Store, bot *tgbotapi.BotAPI, chatID int64) {
	text := fmt.Sprintf(`ℹ️ درباره فروشگاه %s:

📋 توضیحات:
%s

📊 آمار:
🏪 نام فروشگاه: %s
📦 تعداد محصولات: %d
📅 تاریخ شروع: %s
💎 سطح خدمات: %s`, 
		store.Name,
		store.Description,
		store.Name,
		len(store.Products),
		store.CreatedAt.Format("2006/01/02"),
		string(store.PlanType))
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 بازگشت", "back_main"),
		),
	)
	
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

// sendStoreHelp sends help message
func (s *BotManagerService) sendStoreHelp(store models.Store, bot *tgbotapi.BotAPI, chatID int64) {
	text := `ℹ️ راهنمای استفاده:

📦 /products - مشاهده محصولات
📞 /contact - اطلاعات تماس
🏪 /start - صفحه اصلی
ℹ️ /help - راهنما

برای خرید محصولات، از بخش محصولات استفاده کنید.`
	
	msg := tgbotapi.NewMessage(chatID, text)
	bot.Send(msg)
}

// Helper functions
func (s *BotManagerService) getActiveStoreProducts(storeID uint) ([]models.Product, error) {
	var products []models.Product
	err := s.db.Where("store_id = ? AND is_available = ?", storeID, true).Find(&products).Error
	return products, err
}

func (s *BotManagerService) formatPrice(price int64) string {
	// Convert to string with thousands separator
	priceStr := fmt.Sprintf("%d", price)
	if len(priceStr) > 3 {
		// Add comma separators
		formatted := ""
		for i, digit := range priceStr {
			if i > 0 && (len(priceStr)-i)%3 == 0 {
				formatted += ","
			}
			formatted += string(digit)
		}
		return formatted
	}
	return priceStr
}

// GetBotStatus returns status of all bots
func (s *BotManagerService) GetBotStatus() map[string]interface{} {
	return map[string]interface{}{
		"active_bots":  len(s.activeBots),
		"running_bots": len(s.runningBots),
		"store_bots":   s.getStoreBotsList(),
	}
}

func (s *BotManagerService) getStoreBotsList() []map[string]interface{} {
	var result []map[string]interface{}
	
	for storeID := range s.runningBots {
		var store models.Store
		if err := s.db.First(&store, storeID).Error; err == nil {
			result = append(result, map[string]interface{}{
				"store_id":   storeID,
				"store_name": store.Name,
				"is_active":  store.IsActive,
			})
		}
	}
	
	return result
}