package services

import (
	"fmt"
	"log"
	"telegram-store-hub/internal/messages"
	"telegram-store-hub/internal/models"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

// SubscriptionService manages subscription plans and renewals
type SubscriptionService struct {
	bot               *tgbotapi.BotAPI
	db                *gorm.DB
	paymentCardNumber string
	paymentCardHolder string
}

// PlanDetails contains plan information
type PlanDetails struct {
	Type           string `json:"type"`
	Price          int64  `json:"price"`
	ProductLimit   int    `json:"product_limit"`
	CommissionRate int    `json:"commission_rate"`
	Features       []string `json:"features"`
}

// SubscriptionRenewalData contains renewal information
type SubscriptionRenewalData struct {
	StoreID  uint   `json:"store_id"`
	PlanType string `json:"plan_type"`
	Price    int64  `json:"price"`
}

// NewSubscriptionService creates a new subscription service
func NewSubscriptionService(
	bot *tgbotapi.BotAPI,
	db *gorm.DB,
	paymentCardNumber string,
	paymentCardHolder string,
) *SubscriptionService {
	return &SubscriptionService{
		bot:               bot,
		db:                db,
		paymentCardNumber: paymentCardNumber,
		paymentCardHolder: paymentCardHolder,
	}
}

// GetAvailablePlans returns all available subscription plans
func (s *SubscriptionService) GetAvailablePlans() []PlanDetails {
	return []PlanDetails{
		{
			Type:           "free",
			Price:          0,
			ProductLimit:   10,
			CommissionRate: 5,
			Features: []string{
				"10 محصول",
				"کارمزد 5%",
				"پشتیبانی عمومی",
			},
		},
		{
			Type:           "pro",
			Price:          50000, // 50,000 Toman
			ProductLimit:   200,
			CommissionRate: 5,
			Features: []string{
				"200 محصول",
				"کارمزد 5%",
				"پیام خوشامدگویی",
				"پشتیبانی اولویت‌دار",
			},
		},
		{
			Type:           "vip",
			Price:          150000, // 150,000 Toman
			ProductLimit:   -1,     // unlimited
			CommissionRate: 0,
			Features: []string{
				"محصولات نامحدود",
				"بدون کارمزد",
				"ویژگی‌های اختصاصی",
				"پشتیبانی VIP",
				"قابلیت‌های پیشرفته",
			},
		},
	}
}

// GetPlanByType returns plan details by type
func (s *SubscriptionService) GetPlanByType(planType string) *PlanDetails {
	plans := s.GetAvailablePlans()
	for _, plan := range plans {
		if plan.Type == planType {
			return &plan
		}
	}
	return nil
}

// ShowPlanRenewal displays plan renewal options
func (s *SubscriptionService) ShowPlanRenewal(chatID int64) {
	// Get user's store
	store, err := s.getUserStore(chatID)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, messages.ErrorNoStore)
		s.bot.Send(msg)
		return
	}

	// Calculate remaining days
	daysRemaining := int(time.Until(store.ExpiresAt).Hours() / 24)
	if daysRemaining < 0 {
		daysRemaining = 0
	}

	currentPlan := s.GetPlanByType(string(store.PlanType))
	if currentPlan == nil {
		msg := tgbotapi.NewMessage(chatID, messages.ErrorDatabaseError)
		s.bot.Send(msg)
		return
	}

	text := fmt.Sprintf(`🔄 تمدید پلن فروشگاه

پلن فعلی: %s
باقیمانده: %d روز

برای تمدید پلن، یکی از گزینه‌های زیر را انتخاب کنید:`,
		string(store.PlanType),
		daysRemaining)

	plans := s.GetAvailablePlans()
	var keyboard [][]tgbotapi.InlineKeyboardButton

	for _, plan := range plans {
		planName := s.getPlanDisplayName(plan.Type)
		priceText := "رایگان"
		if plan.Price > 0 {
			priceText = fmt.Sprintf("%s تومان", formatPrice(plan.Price))
		}

		buttonText := fmt.Sprintf("%s - %s", planName, priceText)
		
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(buttonText, fmt.Sprintf("renew_%s", plan.Type)),
		))
	}

	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🔙 بازگشت", "store_settings"),
	))

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)

	s.bot.Send(msg)
}

// HandlePlanRenewal handles plan renewal request
func (s *SubscriptionService) HandlePlanRenewal(chatID int64, planType string) {
	// Get user's store
	store, err := s.getUserStore(chatID)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, messages.ErrorNoStore)
		s.bot.Send(msg)
		return
	}

	plan := s.GetPlanByType(planType)
	if plan == nil {
		msg := tgbotapi.NewMessage(chatID, "❌ پلن انتخابی معتبر نیست.")
		s.bot.Send(msg)
		return
	}

	if plan.Price == 0 {
		// Free plan - process immediately
		s.ProcessPlanRenewal(store.ID, planType)
		s.SendRenewalSuccess(chatID, planType)
		return
	}

	// Paid plan - show payment instructions
	s.ShowPaymentInstructions(chatID, planType, plan.Price, store.ID)
}

// ShowPaymentInstructions displays payment instructions for plan renewal
func (s *SubscriptionService) ShowPaymentInstructions(chatID int64, planType string, price int64, storeID uint) {
	planName := s.getPlanDisplayName(planType)

	text := fmt.Sprintf(`💳 پرداخت تمدید پلن %s

مبلغ قابل پرداخت: %s تومان

%s

%s

پس از پرداخت، روی دکمه "تایید پرداخت" کلیک کنید.`,
		planName,
		formatPrice(price),
		fmt.Sprintf(messages.PaymentCardInfo, s.paymentCardNumber, s.paymentCardHolder, formatPrice(price)),
		messages.PaymentInstructions)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ تایید پرداخت", fmt.Sprintf("confirm_renewal_%s_%d", planType, storeID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 بازگشت", "renew_plan"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard

	s.bot.Send(msg)
}

// ProcessPlanRenewal processes the plan renewal
func (s *SubscriptionService) ProcessPlanRenewal(storeID uint, planType string) error {
	plan := s.GetPlanByType(planType)
	if plan == nil {
		return fmt.Errorf("invalid plan type: %s", planType)
	}

	// Update store with new plan
	updates := map[string]interface{}{
		"plan_type":       models.PlanType(planType),
		"product_limit":   plan.ProductLimit,
		"commission_rate": plan.CommissionRate,
		"is_active":       true,
		"expires_at":      time.Now().AddDate(0, 1, 0), // Extend by 1 month
		"updated_at":      time.Now(),
	}

	if err := s.db.Model(&models.Store{}).Where("id = ?", storeID).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update store plan: %w", err)
	}

	log.Printf("Plan renewed for store %d to %s", storeID, planType)
	return nil
}

// SendRenewalSuccess sends renewal success message
func (s *SubscriptionService) SendRenewalSuccess(chatID int64, planType string) {
	planName := s.getPlanDisplayName(planType)
	plan := s.GetPlanByType(planType)

	text := fmt.Sprintf(`🎉 تبریک! پلن شما با موفقیت تمدید شد!

📋 اطلاعات پلن جدید:
• نوع: %s
• محصولات مجاز: %s
• کارمزد: %d%%
• مدت اعتبار: 1 ماه

✨ ویژگی‌های پلن:
%s`,
		planName,
		func() string {
			if plan.ProductLimit == -1 {
				return "نامحدود"
			}
			return fmt.Sprintf("%d", plan.ProductLimit)
		}(),
		plan.CommissionRate,
		s.formatFeatures(plan.Features))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏪 پنل مدیریت", "manage_store"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 منوی اصلی", "back_main"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard

	s.bot.Send(msg)
}

// CheckExpiringSubscriptions checks for expiring subscriptions and sends reminders
func (s *SubscriptionService) CheckExpiringSubscriptions() {
	reminderDays := []int{7, 3, 1} // Remind 7, 3, and 1 days before expiry

	for _, days := range reminderDays {
		reminderDate := time.Now().AddDate(0, 0, days)
		
		var stores []models.Store
		err := s.db.Preload("Owner").Where("expires_at::date = ? AND is_active = ?", reminderDate.Format("2006-01-02"), true).Find(&stores).Error
		if err != nil {
			log.Printf("Error finding expiring stores: %v", err)
			continue
		}

		for _, store := range stores {
			s.SendExpiryReminder(store.Owner.TelegramID, &store, days)
		}
	}
}

// SendExpiryReminder sends subscription expiry reminder
func (s *SubscriptionService) SendExpiryReminder(chatID int64, store *models.Store, daysRemaining int) {
	var text string
	
	if daysRemaining == 1 {
		text = fmt.Sprintf(`⚠️ هشدار! پلن فروشگاه "%s" فردا منقضی می‌شود!

برای جلوگیری از قطع سرویس، همین حالا پلن خود را تمدید کنید.`, store.Name)
	} else {
		text = fmt.Sprintf(`⏰ یادآوری: پلن فروشگاه "%s" %d روز دیگر منقضی می‌شود.

برای تمدید پلن روی دکمه زیر کلیک کنید.`, store.Name, daysRemaining)
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 تمدید پلن", "renew_plan"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏪 پنل مدیریت", "manage_store"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard

	s.bot.Send(msg)
}

// DeactivateExpiredSubscriptions deactivates expired subscriptions
func (s *SubscriptionService) DeactivateExpiredSubscriptions() {
	yesterday := time.Now().AddDate(0, 0, -1)
	
	var expiredStores []models.Store
	err := s.db.Preload("Owner").Where("expires_at < ? AND is_active = ?", yesterday, true).Find(&expiredStores).Error
	if err != nil {
		log.Printf("Error finding expired stores: %v", err)
		return
	}

	for _, store := range expiredStores {
		// Deactivate store
		s.db.Model(&store).Update("is_active", false)
		
		// Send expiry notification
		s.SendExpiryNotification(store.Owner.TelegramID, &store)
		
		log.Printf("Deactivated expired store: %s (ID: %d)", store.Name, store.ID)
	}
}

// SendExpiryNotification sends subscription expired notification
func (s *SubscriptionService) SendExpiryNotification(chatID int64, store *models.Store) {
	text := fmt.Sprintf(`❌ پلن فروشگاه "%s" منقضی شده است!

فروشگاه شما غیرفعال شده و مشتریان نمی‌توانند سفارش ثبت کنند.

برای فعال‌سازی مجدد، پلن خود را تمدید کنید.`, store.Name)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 تمدید پلن", "renew_plan"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 منوی اصلی", "back_main"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard

	s.bot.Send(msg)
}

// Helper methods

func (s *SubscriptionService) getUserStore(chatID int64) (*models.Store, error) {
	var user models.User
	if err := s.db.Where("telegram_id = ?", chatID).First(&user).Error; err != nil {
		return nil, err
	}

	var store models.Store
	if err := s.db.Where("owner_id = ?", user.ID).First(&store).Error; err != nil {
		return nil, err
	}

	return &store, nil
}

func (s *SubscriptionService) getPlanDisplayName(planType string) string {
	switch planType {
	case "free":
		return "رایگان"
	case "pro":
		return "حرفه‌ای"
	case "vip":
		return "ویژه"
	default:
		return planType
	}
}

func (s *SubscriptionService) formatFeatures(features []string) string {
	result := ""
	for _, feature := range features {
		result += "• " + feature + "\n"
	}
	return result
}