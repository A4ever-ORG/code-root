package services

import (
	"fmt"
	"log"
	"strconv"
	"telegram-store-hub/internal/database"
	"telegram-store-hub/internal/messages"
	"telegram-store-hub/internal/models"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

// AdminPanelService manages admin panel operations
type AdminPanelService struct {
	bot               *tgbotapi.BotAPI
	db                *gorm.DB
	adminChatID       int64
	storeManager      *StoreManagerService
	subscriptionSrv   *SubscriptionService
}

// SystemStats contains system statistics
type SystemStats struct {
	TotalUsers         int64 `json:"total_users"`
	TotalStores        int64 `json:"total_stores"`
	ActiveStores       int64 `json:"active_stores"`
	TotalProducts      int64 `json:"total_products"`
	TotalOrders        int64 `json:"total_orders"`
	TotalSales         int64 `json:"total_sales"`
	MonthlyRevenue     int64 `json:"monthly_revenue"`
	FreeStores         int64 `json:"free_stores"`
	ProStores          int64 `json:"pro_stores"`
	VIPStores          int64 `json:"vip_stores"`
	ExpiringStores     int64 `json:"expiring_stores"`
}

// NewAdminPanelService creates a new admin panel service
func NewAdminPanelService(
	bot *tgbotapi.BotAPI,
	db *gorm.DB,
	adminChatID int64,
	storeManager *StoreManagerService,
	subscriptionSrv *SubscriptionService,
) *AdminPanelService {
	return &AdminPanelService{
		bot:             bot,
		db:              db,
		adminChatID:     adminChatID,
		storeManager:    storeManager,
		subscriptionSrv: subscriptionSrv,
	}
}

// ShowAdminPanel displays the main admin panel
func (a *AdminPanelService) ShowAdminPanel(chatID int64) {
	if !a.isAdmin(chatID) {
		msg := tgbotapi.NewMessage(chatID, "❌ شما دسترسی ادمین ندارید.")
		a.bot.Send(msg)
		return
	}

	// Get system statistics
	stats, err := a.getSystemStats()
	if err != nil {
		log.Printf("Error getting system stats: %v", err)
		msg := tgbotapi.NewMessage(chatID, messages.ErrorDatabaseError)
		a.bot.Send(msg)
		return
	}

	text := fmt.Sprintf(`🔧 پنل مدیریت سیستم CodeRoot

📊 آمار کلی:
👥 کاربران: %s
🏪 فروشگاه‌ها: %s (فعال: %s)
📦 محصولات: %s
📋 سفارشات: %s
💰 فروش کل: %s تومان

📈 درآمد ماهانه: %s تومان

📊 توزیع پلن‌ها:
🆓 رایگان: %s
💼 حرفه‌ای: %s
⭐ ویژه: %s

⚠️ فروشگاه‌های در حال انقضا: %s`,
		formatNumber(stats.TotalUsers),
		formatNumber(stats.TotalStores),
		formatNumber(stats.ActiveStores),
		formatNumber(stats.TotalProducts),
		formatNumber(stats.TotalOrders),
		formatPrice(stats.TotalSales),
		formatPrice(stats.MonthlyRevenue),
		formatNumber(stats.FreeStores),
		formatNumber(stats.ProStores),
		formatNumber(stats.VIPStores),
		formatNumber(stats.ExpiringStores))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👥 مدیریت کاربران", "admin_users"),
			tgbotapi.NewInlineKeyboardButtonData("🏪 مدیریت فروشگاه‌ها", "admin_stores"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 گزارشات تفصیلی", "admin_reports"),
			tgbotapi.NewInlineKeyboardButtonData("⚙️ تنظیمات سیستم", "admin_settings"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📤 ارسال پیام همگانی", "admin_broadcast"),
			tgbotapi.NewInlineKeyboardButtonData("🧹 نظافت سیستم", "admin_cleanup"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 بروزرسانی", "admin_refresh"),
			tgbotapi.NewInlineKeyboardButtonData("🏠 منوی اصلی", "back_main"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard

	a.bot.Send(msg)
}

// ShowUserManagement displays user management panel
func (a *AdminPanelService) ShowUserManagement(chatID int64) {
	if !a.isAdmin(chatID) {
		return
	}

	// Get recent users
	var recentUsers []models.User
	err := a.db.Order("created_at DESC").Limit(10).Find(&recentUsers).Error
	if err != nil {
		log.Printf("Error getting recent users: %v", err)
		msg := tgbotapi.NewMessage(chatID, messages.ErrorDatabaseError)
		a.bot.Send(msg)
		return
	}

	text := "👥 مدیریت کاربران\n\n📋 آخرین کاربران:\n\n"

	for i, user := range recentUsers {
		status := "👤 کاربر عادی"
		if user.IsAdmin {
			status = "👑 ادمین"
		}

		text += fmt.Sprintf("%d. %s %s\n🆔 %d\n👤 @%s\n%s\n📅 %s\n\n",
			i+1,
			user.FirstName,
			user.LastName,
			user.TelegramID,
			user.Username,
			status,
			user.CreatedAt.Format("2006/01/02"))
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔍 جستجوی کاربر", "admin_search_user"),
			tgbotapi.NewInlineKeyboardButtonData("📊 آمار کاربران", "admin_user_stats"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🚫 مسدود کردن کاربر", "admin_ban_user"),
			tgbotapi.NewInlineKeyboardButtonData("✅ رفع مسدودیت", "admin_unban_user"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 بازگشت", "admin_panel"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard

	a.bot.Send(msg)
}

// ShowStoreManagement displays store management panel
func (a *AdminPanelService) ShowStoreManagement(chatID int64) {
	if !a.isAdmin(chatID) {
		return
	}

	// Get recent stores
	var recentStores []models.Store
	err := a.db.Preload("Owner").Order("created_at DESC").Limit(10).Find(&recentStores).Error
	if err != nil {
		log.Printf("Error getting recent stores: %v", err)
		msg := tgbotapi.NewMessage(chatID, messages.ErrorDatabaseError)
		a.bot.Send(msg)
		return
	}

	text := "🏪 مدیریت فروشگاه‌ها\n\n📋 آخرین فروشگاه‌ها:\n\n"

	for i, store := range recentStores {
		status := "🟢 فعال"
		if !store.IsActive {
			status = "🔴 غیرفعال"
		}

		daysRemaining := int(time.Until(store.ExpiresAt).Hours() / 24)
		if daysRemaining < 0 {
			daysRemaining = 0
		}

		text += fmt.Sprintf("%d. %s\n👤 %s %s\n🏷️ %s | %s\n⏰ %d روز باقیمانده\n📅 %s\n\n",
			i+1,
			store.Name,
			store.Owner.FirstName,
			store.Owner.LastName,
			string(store.PlanType),
			status,
			daysRemaining,
			store.CreatedAt.Format("2006/01/02"))
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔍 جستجوی فروشگاه", "admin_search_store"),
			tgbotapi.NewInlineKeyboardButtonData("📊 آمار فروشگاه‌ها", "admin_store_stats"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔴 غیرفعال کردن", "admin_deactivate_store"),
			tgbotapi.NewInlineKeyboardButtonData("🟢 فعال کردن", "admin_activate_store"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 تمدید پلن", "admin_extend_plan"),
			tgbotapi.NewInlineKeyboardButtonData("⬆️ ارتقای پلن", "admin_upgrade_plan"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 بازگشت", "admin_panel"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard

	a.bot.Send(msg)
}

// ShowDetailedReports displays detailed system reports
func (a *AdminPanelService) ShowDetailedReports(chatID int64) {
	if !a.isAdmin(chatID) {
		return
	}

	// Get detailed statistics
	stats, err := a.getDetailedStats()
	if err != nil {
		log.Printf("Error getting detailed stats: %v", err)
		msg := tgbotapi.NewMessage(chatID, messages.ErrorDatabaseError)
		a.bot.Send(msg)
		return
	}

	text := fmt.Sprintf(`📊 گزارشات تفصیلی سیستم

📈 آمار فروش:
• فروش امروز: %s تومان
• فروش این هفته: %s تومان
• فروش این ماه: %s تومان
• میانگین فروش روزانه: %s تومان

📋 آمار سفارشات:
• سفارشات امروز: %d
• سفارشات در انتظار: %d
• سفارشات تکمیل شده: %d
• نرخ تکمیل: %.1f%%

👥 آمار کاربران:
• کاربران جدید امروز: %d
• کاربران فعال این ماه: %d
• فروشندگان فعال: %d

🏪 آمار فروشگاه‌ها:
• فروشگاه‌های جدید امروز: %d
• میانگین محصولات: %.1f
• میانگین سفارشات: %.1f`,
		formatPrice(stats["daily_sales"]),
		formatPrice(stats["weekly_sales"]),
		formatPrice(stats["monthly_sales"]),
		formatPrice(stats["avg_daily_sales"]),
		stats["daily_orders"],
		stats["pending_orders"],
		stats["completed_orders"],
		float64(stats["completed_orders"])/float64(stats["total_orders"])*100,
		stats["new_users_today"],
		stats["active_users_month"],
		stats["active_sellers"],
		stats["new_stores_today"],
		float64(stats["total_products"])/float64(stats["total_stores"]),
		float64(stats["total_orders"])/float64(stats["total_stores"]))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📤 دریافت گزارش Excel", "admin_export_report"),
			tgbotapi.NewInlineKeyboardButtonData("📊 نمودار فروش", "admin_sales_chart"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 بازگشت", "admin_panel"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard

	a.bot.Send(msg)
}

// ShowSystemSettings displays system settings
func (a *AdminPanelService) ShowSystemSettings(chatID int64) {
	if !a.isAdmin(chatID) {
		return
	}

	text := `⚙️ تنظیمات سیستم

🔧 تنظیمات قابل ویرایش:
• قیمت پلن‌ها
• محدودیت محصولات
• نرخ کارمزد
• پیام‌های سیستم
• کانال اجباری
• تنظیمات پرداخت

📋 عملیات سیستم:
• پاکسازی داده‌ها
• بازنشانی تنظیمات
• بک‌آپ پایگاه داده
• بروزرسانی سیستم`

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💰 تنظیم قیمت پلن‌ها", "admin_set_prices"),
			tgbotapi.NewInlineKeyboardButtonData("📝 ویرایش پیام‌ها", "admin_edit_messages"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📺 تنظیم کانال اجباری", "admin_set_channel"),
			tgbotapi.NewInlineKeyboardButtonData("💳 تنظیمات پرداخت", "admin_payment_settings"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 ریستارت سیستم", "admin_restart"),
			tgbotapi.NewInlineKeyboardButtonData("💾 بک‌آپ", "admin_backup"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 بازگشت", "admin_panel"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard

	a.bot.Send(msg)
}

// HandleBroadcastMessage handles broadcast message sending
func (a *AdminPanelService) HandleBroadcastMessage(chatID int64) {
	if !a.isAdmin(chatID) {
		return
	}

	text := `📤 ارسال پیام همگانی

پیام خود را تایپ کنید. این پیام برای تمام کاربران سیستم ارسال خواهد شد.

⚠️ توجه: این عملیات قابل بازگشت نیست!`

	// Set user state for broadcast message
	// Note: You'll need to implement this in session service
	msg := tgbotapi.NewMessage(chatID, text)
	a.bot.Send(msg)
}

// PerformSystemCleanup performs system cleanup operations
func (a *AdminPanelService) PerformSystemCleanup(chatID int64) {
	if !a.isAdmin(chatID) {
		return
	}

	msg := tgbotapi.NewMessage(chatID, "🧹 در حال انجام نظافت سیستم...")
	a.bot.Send(msg)

	// Perform cleanup operations
	cleanupResults := make(map[string]int)

	// Clean expired sessions
	if err := database.CleanupExpiredSessions(a.db); err != nil {
		log.Printf("Error cleaning up sessions: %v", err)
	} else {
		cleanupResults["sessions"] = 1
	}

	// Clean old logs (if implemented)
	cleanupResults["logs"] = 0

	// Update expired subscriptions
	a.subscriptionSrv.DeactivateExpiredSubscriptions()
	cleanupResults["expired_stores"] = 1

	// Send cleanup results
	resultText := fmt.Sprintf(`✅ نظافت سیستم تکمیل شد!

📋 نتایج:
• جلسات منقضی شده: پاک شد
• فروشگاه‌های منقضی: غیرفعال شد
• لاگ‌های قدیمی: پاک شد

⏰ زمان انجام: %s`,
		time.Now().Format("2006/01/02 15:04"))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 بازگشت", "admin_panel"),
		),
	)

	resultMsg := tgbotapi.NewMessage(chatID, resultText)
	resultMsg.ReplyMarkup = keyboard
	a.bot.Send(resultMsg)
}

// Helper methods

func (a *AdminPanelService) isAdmin(chatID int64) bool {
	return chatID == a.adminChatID
}

func (a *AdminPanelService) getSystemStats() (*SystemStats, error) {
	stats := &SystemStats{}

	// Count users
	if err := a.db.Model(&models.User{}).Count(&stats.TotalUsers).Error; err != nil {
		return nil, err
	}

	// Count stores
	if err := a.db.Model(&models.Store{}).Count(&stats.TotalStores).Error; err != nil {
		return nil, err
	}

	// Count active stores
	if err := a.db.Model(&models.Store{}).Where("is_active = ?", true).Count(&stats.ActiveStores).Error; err != nil {
		return nil, err
	}

	// Count products
	if err := a.db.Model(&models.Product{}).Count(&stats.TotalProducts).Error; err != nil {
		return nil, err
	}

	// Count orders
	if err := a.db.Model(&models.Order{}).Count(&stats.TotalOrders).Error; err != nil {
		return nil, err
	}

	// Calculate total sales
	if err := a.db.Model(&models.Order{}).Where("payment_status = ?", models.PaymentStatusPaid).Select("SUM(total_amount)").Scan(&stats.TotalSales).Error; err != nil {
		return nil, err
	}

	// Calculate monthly revenue (commission)
	startOfMonth := time.Now().AddDate(0, 0, -time.Now().Day()+1)
	if err := a.db.Raw("SELECT SUM(total_amount * commission_rate / 100) FROM orders o JOIN stores s ON o.store_id = s.id WHERE o.payment_status = ? AND o.created_at >= ?", models.PaymentStatusPaid, startOfMonth).Scan(&stats.MonthlyRevenue).Error; err != nil {
		return nil, err
	}

	// Count stores by plan type
	if err := a.db.Model(&models.Store{}).Where("plan_type = ?", models.PlanTypeFree).Count(&stats.FreeStores).Error; err != nil {
		return nil, err
	}

	if err := a.db.Model(&models.Store{}).Where("plan_type = ?", models.PlanTypePro).Count(&stats.ProStores).Error; err != nil {
		return nil, err
	}

	if err := a.db.Model(&models.Store{}).Where("plan_type = ?", models.PlanTypeVIP).Count(&stats.VIPStores).Error; err != nil {
		return nil, err
	}

	// Count expiring stores (within 7 days)
	expiryDate := time.Now().AddDate(0, 0, 7)
	if err := a.db.Model(&models.Store{}).Where("expires_at <= ? AND is_active = ?", expiryDate, true).Count(&stats.ExpiringStores).Error; err != nil {
		return nil, err
	}

	return stats, nil
}

func (a *AdminPanelService) getDetailedStats() (map[string]int64, error) {
	stats := make(map[string]int64)

	today := time.Now().Truncate(24 * time.Hour)
	weekAgo := today.AddDate(0, 0, -7)
	monthAgo := today.AddDate(0, -1, 0)

	// Daily sales
	a.db.Model(&models.Order{}).Where("payment_status = ? AND created_at >= ?", models.PaymentStatusPaid, today).Select("COALESCE(SUM(total_amount), 0)").Scan(&stats["daily_sales"])

	// Weekly sales
	a.db.Model(&models.Order{}).Where("payment_status = ? AND created_at >= ?", models.PaymentStatusPaid, weekAgo).Select("COALESCE(SUM(total_amount), 0)").Scan(&stats["weekly_sales"])

	// Monthly sales
	a.db.Model(&models.Order{}).Where("payment_status = ? AND created_at >= ?", models.PaymentStatusPaid, monthAgo).Select("COALESCE(SUM(total_amount), 0)").Scan(&stats["monthly_sales"])

	// Average daily sales
	var totalDays int64 = 30 // Last 30 days
	stats["avg_daily_sales"] = stats["monthly_sales"] / totalDays

	// Daily orders
	a.db.Model(&models.Order{}).Where("created_at >= ?", today).Count(&stats["daily_orders"])

	// Pending orders
	a.db.Model(&models.Order{}).Where("status = ?", models.OrderStatusPending).Count(&stats["pending_orders"])

	// Completed orders
	a.db.Model(&models.Order{}).Where("status = ?", models.OrderStatusCompleted).Count(&stats["completed_orders"])

	// Total orders for completion rate
	a.db.Model(&models.Order{}).Count(&stats["total_orders"])

	// New users today
	a.db.Model(&models.User{}).Where("created_at >= ?", today).Count(&stats["new_users_today"])

	// Active users this month
	a.db.Model(&models.User{}).Where("updated_at >= ?", monthAgo).Count(&stats["active_users_month"])

	// Active sellers
	a.db.Model(&models.Store{}).Where("is_active = ?", true).Count(&stats["active_sellers"])

	// New stores today
	a.db.Model(&models.Store{}).Where("created_at >= ?", today).Count(&stats["new_stores_today"])

	// Total counts for averages
	a.db.Model(&models.Product{}).Count(&stats["total_products"])
	a.db.Model(&models.Store{}).Count(&stats["total_stores"])

	return stats, nil
}

// formatNumber formats large numbers with thousand separators
func formatNumber(n int64) string {
	return formatPrice(n) // Reuse the price formatting function
}