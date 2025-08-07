package services

import (
	"fmt"
	"log"
	"telegram-store-hub/internal/models"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

// ReminderService manages subscription renewal reminders
type ReminderService struct {
	bot               *tgbotapi.BotAPI
	db                *gorm.DB
	subscriptionSrv   *SubscriptionService
	reminderDays      []int
	isRunning         bool
}

// ReminderType defines the type of reminder
type ReminderType string

const (
	ReminderTypeExpiring ReminderType = "expiring"
	ReminderTypeExpired  ReminderType = "expired"
	ReminderTypeOverdue  ReminderType = "overdue"
)

// ReminderLog tracks sent reminders to avoid duplicates
type ReminderLog struct {
	ID          uint      `gorm:"primaryKey"`
	StoreID     uint      `gorm:"not null"`
	ReminderType ReminderType `gorm:"type:varchar(20);not null"`
	DaysRemaining int     `gorm:"not null"`
	SentAt      time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	Store       models.Store `gorm:"foreignKey:StoreID"`
}

// NewReminderService creates a new reminder service
func NewReminderService(
	bot *tgbotapi.BotAPI,
	db *gorm.DB,
	subscriptionSrv *SubscriptionService,
	reminderDays []int,
) *ReminderService {
	return &ReminderService{
		bot:             bot,
		db:              db,
		subscriptionSrv: subscriptionSrv,
		reminderDays:    reminderDays,
		isRunning:       false,
	}
}

// StartReminderScheduler starts the automatic reminder scheduler
func (r *ReminderService) StartReminderScheduler() {
	if r.isRunning {
		log.Println("Reminder scheduler is already running")
		return
	}

	// Migrate reminder log table
	r.db.AutoMigrate(&ReminderLog{})

	r.isRunning = true
	log.Println("Starting subscription reminder scheduler...")

	// Run initial check
	r.CheckAndSendReminders()

	// Schedule periodic checks
	go func() {
		ticker := time.NewTicker(6 * time.Hour) // Check every 6 hours
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if r.isRunning {
					r.CheckAndSendReminders()
				}
			}
		}
	}()

	log.Println("Subscription reminder scheduler started successfully")
}

// StopReminderScheduler stops the reminder scheduler
func (r *ReminderService) StopReminderScheduler() {
	r.isRunning = false
	log.Println("Reminder scheduler stopped")
}

// CheckAndSendReminders checks for expiring subscriptions and sends reminders
func (r *ReminderService) CheckAndSendReminders() {
	log.Println("Checking for subscription expiry reminders...")

	// Check for expiring subscriptions
	r.checkExpiringSubscriptions()

	// Check for expired subscriptions
	r.checkExpiredSubscriptions()

	// Check for overdue subscriptions (expired more than 3 days ago)
	r.checkOverdueSubscriptions()

	log.Println("Subscription reminder check completed")
}

// checkExpiringSubscriptions checks for subscriptions that will expire soon
func (r *ReminderService) checkExpiringSubscriptions() {
	for _, days := range r.reminderDays {
		reminderDate := time.Now().AddDate(0, 0, days).Truncate(24 * time.Hour)
		
		var stores []models.Store
		err := r.db.Preload("Owner").Where(
			"expires_at::date = ? AND is_active = ?", 
			reminderDate.Format("2006-01-02"), 
			true,
		).Find(&stores).Error
		
		if err != nil {
			log.Printf("Error finding expiring stores for %d days: %v", days, err)
			continue
		}

		for _, store := range stores {
			// Check if reminder already sent for this store and day count
			if r.isReminderAlreadySent(store.ID, ReminderTypeExpiring, days) {
				continue
			}

			// Send expiring reminder
			r.sendExpiryReminder(store.Owner.TelegramID, &store, days)
			
			// Log the reminder
			r.logReminder(store.ID, ReminderTypeExpiring, days)
		}

		if len(stores) > 0 {
			log.Printf("Sent %d expiring reminders for %d days", len(stores), days)
		}
	}
}

// checkExpiredSubscriptions checks for recently expired subscriptions
func (r *ReminderService) checkExpiredSubscriptions() {
	yesterday := time.Now().AddDate(0, 0, -1).Truncate(24 * time.Hour)
	
	var expiredStores []models.Store
	err := r.db.Preload("Owner").Where(
		"expires_at::date = ? AND is_active = ?", 
		yesterday.Format("2006-01-02"), 
		true,
	).Find(&expiredStores).Error
	
	if err != nil {
		log.Printf("Error finding expired stores: %v", err)
		return
	}

	for _, store := range expiredStores {
		// Check if reminder already sent
		if r.isReminderAlreadySent(store.ID, ReminderTypeExpired, 0) {
			continue
		}

		// Deactivate store
		r.db.Model(&store).Update("is_active", false)
		
		// Send expiry notification
		r.sendExpiryNotification(store.Owner.TelegramID, &store)
		
		// Log the reminder
		r.logReminder(store.ID, ReminderTypeExpired, 0)
	}

	if len(expiredStores) > 0 {
		log.Printf("Processed %d expired subscriptions", len(expiredStores))
	}
}

// checkOverdueSubscriptions checks for overdue subscriptions (expired > 3 days)
func (r *ReminderService) checkOverdueSubscriptions() {
	overdueDate := time.Now().AddDate(0, 0, -3).Truncate(24 * time.Hour)
	
	var overdueStores []models.Store
	err := r.db.Preload("Owner").Where(
		"expires_at::date = ? AND is_active = ?", 
		overdueDate.Format("2006-01-02"), 
		false,
	).Find(&overdueStores).Error
	
	if err != nil {
		log.Printf("Error finding overdue stores: %v", err)
		return
	}

	for _, store := range overdueStores {
		// Check if reminder already sent
		if r.isReminderAlreadySent(store.ID, ReminderTypeOverdue, 0) {
			continue
		}

		// Send overdue reminder
		r.sendOverdueReminder(store.Owner.TelegramID, &store)
		
		// Log the reminder
		r.logReminder(store.ID, ReminderTypeOverdue, 0)
	}

	if len(overdueStores) > 0 {
		log.Printf("Sent %d overdue reminders", len(overdueStores))
	}
}

// sendExpiryReminder sends subscription expiry reminder
func (r *ReminderService) sendExpiryReminder(chatID int64, store *models.Store, daysRemaining int) {
	var text string
	var urgency string

	switch daysRemaining {
	case 1:
		urgency = "⚠️ هشدار فوری!"
		text = fmt.Sprintf(`%s

پلن فروشگاه "%s" فردا منقضی می‌شود!

🚨 فروشگاه شما تا 24 ساعت دیگر غیرفعال خواهد شد.

برای جلوگیری از قطع سرویس، همین الان پلن خود را تمدید کنید.`,
			urgency, store.Name)
	case 3:
		urgency = "⚠️ هشدار مهم!"
		text = fmt.Sprintf(`%s

پلن فروشگاه "%s" 3 روز دیگر منقضی می‌شود.

⏰ برای جلوگیری از قطع سرویس، هر چه زودتر پلن خود را تمدید کنید.`,
			urgency, store.Name)
	case 7:
		urgency = "📅 یادآوری"
		text = fmt.Sprintf(`%s

پلن فروشگاه "%s" یک هفته دیگر منقضی می‌شود.

💡 توصیه می‌کنیم پلن خود را زودتر تمدید کنید تا سرویس شما قطع نشود.`,
			urgency, store.Name)
	default:
		text = fmt.Sprintf(`📅 یادآوری

پلن فروشگاه "%s" %d روز دیگر منقضی می‌شود.

برای تمدید پلن روی دکمه زیر کلیک کنید.`,
			store.Name, daysRemaining)
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 تمدید فوری", "renew_plan"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏪 پنل مدیریت", "manage_store"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💬 پشتیبانی", "support"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard

	if _, err := r.bot.Send(msg); err != nil {
		log.Printf("Error sending expiry reminder to %d: %v", chatID, err)
	}
}

// sendExpiryNotification sends subscription expired notification
func (r *ReminderService) sendExpiryNotification(chatID int64, store *models.Store) {
	text := fmt.Sprintf(`❌ پلن فروشگاه منقضی شد!

🏪 فروشگاه: "%s"

فروشگاه شما غیرفعال شده و مشتریان نمی‌توانند سفارش ثبت کنند.

🔄 برای فعال‌سازی مجدد، پلن خود را تمدید کنید.

⚡ پس از تمدید، فروشگاه شما فوراً فعال خواهد شد.`,
		store.Name)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 تمدید فوری", "renew_plan"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💬 پشتیبانی", "support"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard

	if _, err := r.bot.Send(msg); err != nil {
		log.Printf("Error sending expiry notification to %d: %v", chatID, err)
	}
}

// sendOverdueReminder sends overdue subscription reminder
func (r *ReminderService) sendOverdueReminder(chatID int64, store *models.Store) {
	text := fmt.Sprintf(`🚨 پلن فروشگاه شما 3 روز است که منقضی شده!

🏪 فروشگاه: "%s"

⚠️ فروشگاه شما در حال حاضر غیرفعال است.

💰 تمام فروش‌ها و سفارشات متوقف شده‌اند.

🔄 برای بازگرداندن فروشگاه به حالت فعال، پلن خود را تمدید کنید.

📞 در صورت نیاز به راهنمایی، با پشتیبانی تماس بگیرید.`,
		store.Name)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 تمدید فوری", "renew_plan"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📞 تماس با پشتیبانی", "support"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard

	if _, err := r.bot.Send(msg); err != nil {
		log.Printf("Error sending overdue reminder to %d: %v", chatID, err)
	}
}

// isReminderAlreadySent checks if a reminder was already sent
func (r *ReminderService) isReminderAlreadySent(storeID uint, reminderType ReminderType, daysRemaining int) bool {
	var count int64
	
	// Check if reminder was sent in the last 24 hours
	cutoff := time.Now().Add(-24 * time.Hour)
	
	r.db.Model(&ReminderLog{}).Where(
		"store_id = ? AND reminder_type = ? AND days_remaining = ? AND sent_at > ?",
		storeID, reminderType, daysRemaining, cutoff,
	).Count(&count)
	
	return count > 0
}

// logReminder logs a sent reminder
func (r *ReminderService) logReminder(storeID uint, reminderType ReminderType, daysRemaining int) {
	reminderLog := &ReminderLog{
		StoreID:       storeID,
		ReminderType:  reminderType,
		DaysRemaining: daysRemaining,
		SentAt:        time.Now(),
	}
	
	if err := r.db.Create(reminderLog).Error; err != nil {
		log.Printf("Error logging reminder: %v", err)
	}
}

// GetReminderStats returns reminder statistics
func (r *ReminderService) GetReminderStats(days int) (map[string]int64, error) {
	stats := make(map[string]int64)
	
	cutoff := time.Now().AddDate(0, 0, -days)
	
	// Count reminders sent in the last N days
	r.db.Model(&ReminderLog{}).Where("sent_at > ?", cutoff).Count(&stats["total_sent"])
	
	// Count by type
	r.db.Model(&ReminderLog{}).Where("sent_at > ? AND reminder_type = ?", cutoff, ReminderTypeExpiring).Count(&stats["expiring"])
	r.db.Model(&ReminderLog{}).Where("sent_at > ? AND reminder_type = ?", cutoff, ReminderTypeExpired).Count(&stats["expired"])
	r.db.Model(&ReminderLog{}).Where("sent_at > ? AND reminder_type = ?", cutoff, ReminderTypeOverdue).Count(&stats["overdue"])
	
	return stats, nil
}

// CleanupOldReminderLogs removes old reminder logs
func (r *ReminderService) CleanupOldReminderLogs() error {
	// Keep logs for 30 days
	cutoff := time.Now().AddDate(0, 0, -30)
	
	result := r.db.Where("sent_at < ?", cutoff).Delete(&ReminderLog{})
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup old reminder logs: %w", result.Error)
	}
	
	if result.RowsAffected > 0 {
		log.Printf("Cleaned up %d old reminder logs", result.RowsAffected)
	}
	
	return nil
}

// SendManualReminder sends a manual reminder to a specific store
func (r *ReminderService) SendManualReminder(storeID uint, message string) error {
	var store models.Store
	if err := r.db.Preload("Owner").First(&store, storeID).Error; err != nil {
		return fmt.Errorf("failed to get store: %w", err)
	}
	
	text := fmt.Sprintf(`📢 پیام از مدیریت سیستم

🏪 فروشگاه: %s

%s

📞 در صورت نیاز به راهنمایی، با پشتیبانی تماس بگیرید.`,
		store.Name, message)
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏪 پنل مدیریت", "manage_store"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💬 پشتیبانی", "support"),
		),
	)
	
	msg := tgbotapi.NewMessage(store.Owner.TelegramID, text)
	msg.ReplyMarkup = keyboard
	
	if _, err := r.bot.Send(msg); err != nil {
		return fmt.Errorf("failed to send manual reminder: %w", err)
	}
	
	// Log the manual reminder
	r.logReminder(storeID, "manual", 0)
	
	return nil
}

// GetUpcomingExpirations returns stores expiring in the next N days
func (r *ReminderService) GetUpcomingExpirations(days int) ([]models.Store, error) {
	expiryDate := time.Now().AddDate(0, 0, days)
	
	var stores []models.Store
	err := r.db.Preload("Owner").Where(
		"expires_at <= ? AND is_active = ?", 
		expiryDate, 
		true,
	).Order("expires_at ASC").Find(&stores).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to get upcoming expirations: %w", err)
	}
	
	return stores, nil
}