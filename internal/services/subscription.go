package services

import (
	"fmt"
	"log"
	"telegram-store-hub/internal/models"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

type SubscriptionService struct {
	db  *gorm.DB
	bot *tgbotapi.BotAPI
}

func NewSubscriptionService(db *gorm.DB) *SubscriptionService {
	return &SubscriptionService{db: db}
}

// SetBot sets the bot instance for sending notifications
func (s *SubscriptionService) SetBot(bot *tgbotapi.BotAPI) {
	s.bot = bot
}

// CheckExpiredStores checks and handles expired stores
func (s *SubscriptionService) CheckExpiredStores() error {
	var expiredStores []models.Store
	err := s.db.Where("expires_at < ? AND is_active = ?", time.Now(), true).
		Preload("Owner").Find(&expiredStores).Error
	if err != nil {
		return err
	}
	
	for _, store := range expiredStores {
		// Deactivate store
		s.db.Model(&store).Update("is_active", false)
		
		// Send notification to owner
		if s.bot != nil {
			s.sendExpirationNotification(store)
		}
		
		log.Printf("Store %s (ID: %d) has been deactivated due to expiration", store.Name, store.ID)
	}
	
	return nil
}

// CheckExpiringSoonStores checks stores expiring within specified days and sends reminders
func (s *SubscriptionService) CheckExpiringSoonStores(days int) error {
	endDate := time.Now().AddDate(0, 0, days)
	
	var expiringSoonStores []models.Store
	err := s.db.Where("expires_at BETWEEN ? AND ? AND is_active = ?", time.Now(), endDate, true).
		Preload("Owner").Find(&expiringSoonStores).Error
	if err != nil {
		return err
	}
	
	for _, store := range expiringSoonStores {
		// Send reminder notification
		if s.bot != nil {
			s.sendExpirationReminder(store, days)
		}
		
		log.Printf("Sent expiration reminder for store %s (ID: %d)", store.Name, store.ID)
	}
	
	return nil
}

// UpgradeStorePlan upgrades a store to a higher plan
func (s *SubscriptionService) UpgradeStorePlan(storeID uint, newPlan models.PlanType, months int) error {
	var store models.Store
	if err := s.db.First(&store, storeID).Error; err != nil {
		return err
	}
	
	// Update plan details
	var productLimit, commissionRate int
	switch newPlan {
	case models.PlanFree:
		productLimit = 10
		commissionRate = 5
	case models.PlanPro:
		productLimit = 200
		commissionRate = 5
	case models.PlanVIP:
		productLimit = -1 // unlimited
		commissionRate = 0
	}
	
	// Calculate new expiry date
	baseTime := time.Now()
	if store.ExpiresAt.After(time.Now()) {
		baseTime = store.ExpiresAt
	}
	newExpiryDate := baseTime.AddDate(0, months, 0)
	
	// Update store
	updates := map[string]interface{}{
		"plan_type":       newPlan,
		"product_limit":   productLimit,
		"commission_rate": commissionRate,
		"expires_at":      newExpiryDate,
		"is_active":       true,
	}
	
	return s.db.Model(&store).Updates(updates).Error
}

// RenewStorePlan renews a store plan for specified months
func (s *SubscriptionService) RenewStorePlan(storeID uint, months int) error {
	var store models.Store
	if err := s.db.First(&store, storeID).Error; err != nil {
		return err
	}
	
	// Calculate new expiry date
	baseTime := store.ExpiresAt
	if time.Now().After(store.ExpiresAt) {
		baseTime = time.Now()
	}
	newExpiryDate := baseTime.AddDate(0, months, 0)
	
	// Update store
	updates := map[string]interface{}{
		"expires_at": newExpiryDate,
		"is_active":  true,
	}
	
	return s.db.Model(&store).Updates(updates).Error
}

// GetPlanLimits returns the limits for a specific plan
func (s *SubscriptionService) GetPlanLimits(planType models.PlanType) map[string]interface{} {
	switch planType {
	case models.PlanFree:
		return map[string]interface{}{
			"product_limit":   10,
			"commission_rate": 5,
			"features": []string{
				"حداکثر ۱۰ محصول",
				"دکمه‌های ثابت",
				"پشتیبانی پایه",
			},
		}
	case models.PlanPro:
		return map[string]interface{}{
			"product_limit":   200,
			"commission_rate": 5,
			"features": []string{
				"تا ۲۰۰ محصول",
				"گزارش‌های پیشرفته",
				"پیام خوش‌آمدگویی",
				"تبلیغات دلخواه",
				"پشتیبانی اولویت‌دار",
			},
		}
	case models.PlanVIP:
		return map[string]interface{}{
			"product_limit":   -1, // unlimited
			"commission_rate": 0,
			"features": []string{
				"محصولات نامحدود",
				"درگاه پرداخت اختصاصی",
				"بدون کارمزد",
				"تبلیغات ویژه",
				"دکمه‌های شخصی‌سازی شده",
				"پشتیبانی ۲۴/۷",
			},
		}
	default:
		return map[string]interface{}{}
	}
}

// GetPlanPrice returns the price for a specific plan
func (s *SubscriptionService) GetPlanPrice(planType models.PlanType) int64 {
	switch planType {
	case models.PlanFree:
		return 0
	case models.PlanPro:
		return 50000 // 50,000 Toman per month
	case models.PlanVIP:
		return 150000 // 150,000 Toman per month
	default:
		return 0
	}
}

// GetSubscriptionStats gets subscription statistics
func (s *SubscriptionService) GetSubscriptionStats() (map[string]interface{}, error) {
	var freeCount, proCount, vipCount, activeCount, expiredCount int64
	
	s.db.Model(&models.Store{}).Where("plan_type = ?", models.PlanFree).Count(&freeCount)
	s.db.Model(&models.Store{}).Where("plan_type = ?", models.PlanPro).Count(&proCount)
	s.db.Model(&models.Store{}).Where("plan_type = ?", models.PlanVIP).Count(&vipCount)
	s.db.Model(&models.Store{}).Where("is_active = ?", true).Count(&activeCount)
	s.db.Model(&models.Store{}).Where("expires_at < ?", time.Now()).Count(&expiredCount)
	
	return map[string]interface{}{
		"free_stores":    freeCount,
		"pro_stores":     proCount,
		"vip_stores":     vipCount,
		"active_stores":  activeCount,
		"expired_stores": expiredCount,
	}, nil
}

// sendExpirationNotification sends notification when store expires
func (s *SubscriptionService) sendExpirationNotification(store models.Store) {
	text := fmt.Sprintf(`⚠️ پلن فروشگاه شما منقضی شد

🏪 فروشگاه: %s
📅 تاریخ انقضا: %s

❌ فروشگاه شما غیرفعال شده است
💡 برای فعال‌سازی مجدد، پلن خود را تمدید کنید`, 
		store.Name,
		store.ExpiresAt.Format("2006/01/02"))
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 تمدید پلن", fmt.Sprintf("renew_plan_%d", store.ID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📞 پشتیبانی", "support"),
		),
	)
	
	msg := tgbotapi.NewMessage(store.Owner.TelegramID, text)
	msg.ReplyMarkup = keyboard
	
	s.bot.Send(msg)
}

// sendExpirationReminder sends reminder before expiration
func (s *SubscriptionService) sendExpirationReminder(store models.Store, daysLeft int) {
	text := fmt.Sprintf(`🔔 یادآوری تمدید پلن

🏪 فروشگاه: %s
⏰ روزهای باقی‌مانده: %d روز
📅 تاریخ انقضا: %s

💡 برای جلوگیری از قطع سرویس، پلن خود را هرچه سریع‌تر تمدید کنید`, 
		store.Name,
		daysLeft,
		store.ExpiresAt.Format("2006/01/02"))
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 تمدید پلن", fmt.Sprintf("renew_plan_%d", store.ID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 پنل فروشگاه", fmt.Sprintf("manage_store_%d", store.ID)),
		),
	)
	
	msg := tgbotapi.NewMessage(store.Owner.TelegramID, text)
	msg.ReplyMarkup = keyboard
	
	s.bot.Send(msg)
}

// StartSubscriptionChecker starts a goroutine to periodically check subscriptions
func (s *SubscriptionService) StartSubscriptionChecker() {
	go func() {
		ticker := time.NewTicker(24 * time.Hour) // Check daily
		defer ticker.Stop()
		
		for range ticker.C {
			// Check expired stores
			if err := s.CheckExpiredStores(); err != nil {
				log.Printf("Error checking expired stores: %v", err)
			}
			
			// Check stores expiring in 3 days
			if err := s.CheckExpiringSoonStores(3); err != nil {
				log.Printf("Error checking expiring stores: %v", err)
			}
			
			// Check stores expiring in 7 days
			if err := s.CheckExpiringSoonStores(7); err != nil {
				log.Printf("Error checking expiring stores (7 days): %v", err)
			}
		}
	}()
	
	log.Println("✅ Subscription checker started")
}