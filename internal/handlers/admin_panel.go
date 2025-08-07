package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"telegram-store-hub/internal/models"
	"telegram-store-hub/internal/services"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

type AdminPanelHandler struct {
	bot          *tgbotapi.BotAPI
	db           *gorm.DB
	storeManager *services.StoreManagerService
	botManager   *services.BotManagerService
	subscription *services.SubscriptionService
}

func NewAdminPanelHandler(bot *tgbotapi.BotAPI, db *gorm.DB) *AdminPanelHandler {
	return &AdminPanelHandler{
		bot:          bot,
		db:           db,
		storeManager: services.NewStoreManagerService(db),
		botManager:   services.NewBotManagerService(db),
		subscription: services.NewSubscriptionService(db),
	}
}

func (aph *AdminPanelHandler) HandleStoreManagement(chatID int64) {
	stores, err := aph.storeManager.GetAllStores()
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ خطا در دریافت فروشگاه‌ها.")
		aph.bot.Send(msg)
		return
	}

	if len(stores) == 0 {
		msg := tgbotapi.NewMessage(chatID, "🏪 هیچ فروشگاهی ثبت نشده است.")
		aph.bot.Send(msg)
		return
	}

	text := "🏪 مدیریت فروشگاه‌ها:\n\n"
	var keyboard [][]tgbotapi.InlineKeyboardButton

	for i, store := range stores {
		status := "✅ فعال"
		if !store.IsActive {
			status = "❌ غیرفعال"
		}

		// Get store stats
		var productCount, orderCount int64
		aph.db.Model(&models.Product{}).Where("store_id = ?", store.ID).Count(&productCount)
		aph.db.Model(&models.Order{}).Where("store_id = ?", store.ID).Count(&orderCount)

		text += fmt.Sprintf("%d. 🏪 %s\n"+
			"👤 مالک: %d\n"+
			"💎 پلن: %s\n"+
			"📊 وضعیت: %s\n"+
			"📦 محصولات: %d\n"+
			"🛒 سفارش‌ها: %d\n"+
			"📅 انقضا: %s\n\n",
			i+1, store.StoreName,
			store.OwnerChatID,
			store.PlanType,
			status,
			productCount,
			orderCount,
			store.ExpiresAt.Format("2006/01/02"),
		)

		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("⚙️ %s", store.StoreName),
				fmt.Sprintf("admin_store_%d", store.ID),
			),
		))
	}

	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🔙 بازگشت", "admin_panel"),
	))

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
	aph.bot.Send(msg)
}

func (aph *AdminPanelHandler) HandlePaymentManagement(chatID int64) {
	var payments []models.Payment
	err := aph.db.Preload("Store").Order("created_at desc").Limit(20).Find(&payments).Error
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ خطا در دریافت پرداخت‌ها.")
		aph.bot.Send(msg)
		return
	}

	if len(payments) == 0 {
		msg := tgbotapi.NewMessage(chatID, "💰 هیچ پرداختی ثبت نشده است.")
		aph.bot.Send(msg)
		return
	}

	text := "💰 آخرین پرداخت‌ها:\n\n"
	var keyboard [][]tgbotapi.InlineKeyboardButton

	for i, payment := range payments {
		statusEmoji := "⏳"
		if payment.Status == "approved" {
			statusEmoji = "✅"
		} else if payment.Status == "rejected" {
			statusEmoji = "❌"
		}

		text += fmt.Sprintf("%d. %s پرداخت #%d\n"+
			"🏪 فروشگاه: %s\n"+
			"💰 مبلغ: %,.0f تومان\n"+
			"💎 پلن: %s\n"+
			"📅 تاریخ: %s\n\n",
			i+1, statusEmoji, payment.ID,
			payment.Store.StoreName,
			payment.Amount,
			payment.PlanType,
			payment.CreatedAt.Format("2006/01/02 15:04"),
		)

		if payment.Status == "pending" {
			keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("✅ تایید", fmt.Sprintf("approve_payment_%d", payment.ID)),
				tgbotapi.NewInlineKeyboardButtonData("❌ رد", fmt.Sprintf("reject_payment_%d", payment.ID)),
			))
		}
	}

	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🔙 بازگشت", "admin_panel"),
	))

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
	aph.bot.Send(msg)
}

func (aph *AdminPanelHandler) HandleFinancialReport(chatID int64) {
	// Calculate financial statistics
	var totalStores, activeStores int64
	var totalRevenue, pendingPayments, approvedPayments float64
	var totalOrders int64

	aph.db.Model(&models.Store{}).Count(&totalStores)
	aph.db.Model(&models.Store{}).Where("is_active = ?", true).Count(&activeStores)
	aph.db.Model(&models.Order{}).Count(&totalOrders)
	aph.db.Model(&models.Order{}).Select("COALESCE(SUM(total_amount), 0)").Scan(&totalRevenue)
	aph.db.Model(&models.Payment{}).Where("status = ?", "pending").Select("COALESCE(SUM(amount), 0)").Scan(&pendingPayments)
	aph.db.Model(&models.Payment{}).Where("status = ?", "approved").Select("COALESCE(SUM(amount), 0)").Scan(&approvedPayments)

	// Monthly statistics
	currentMonth := time.Now().Format("2006-01")
	var monthlyStores, monthlyOrders int64
	var monthlyRevenue float64

	aph.db.Model(&models.Store{}).Where("DATE_TRUNC('month', created_at) = ?", currentMonth).Count(&monthlyStores)
	aph.db.Model(&models.Order{}).Where("DATE_TRUNC('month', created_at) = ?", currentMonth).Count(&monthlyOrders)
	aph.db.Model(&models.Order{}).Where("DATE_TRUNC('month', created_at) = ?", currentMonth).Select("COALESCE(SUM(total_amount), 0)").Scan(&monthlyRevenue)

	reportText := fmt.Sprintf(`📊 گزارش مالی سیستم

🏪 آمار فروشگاه‌ها:
• کل فروشگاه‌ها: %d
• فروشگاه‌های فعال: %d
• فروشگاه‌های جدید این ماه: %d

💰 آمار مالی:
• کل درآمد سیستم: %,.0f تومان
• پرداخت‌های تایید شده: %,.0f تومان
• پرداخت‌های در انتظار: %,.0f تومان

🛒 آمار سفارش‌ها:
• کل سفارش‌ها: %d
• سفارش‌های این ماه: %d
• درآمد این ماه: %,.0f تومان

📈 نرخ رشد:
• فروشگاه‌های فعال: %.1f%%
• متوسط درآمد هر فروشگاه: %,.0f تومان

📅 تاریخ گزارش: %s`,
		totalStores,
		activeStores,
		monthlyStores,
		totalRevenue,
		approvedPayments,
		pendingPayments,
		totalOrders,
		monthlyOrders,
		monthlyRevenue,
		func() float64 {
			if totalStores > 0 {
				return float64(activeStores) / float64(totalStores) * 100
			}
			return 0
		}(),
		func() float64 {
			if activeStores > 0 {
				return totalRevenue / float64(activeStores)
			}
			return 0
		}(),
		time.Now().Format("2006/01/02 15:04"),
	)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📥 صادرات Excel", "export_financial"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 بازگشت", "admin_panel"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, reportText)
	msg.ReplyMarkup = keyboard
	aph.bot.Send(msg)
}

func (aph *AdminPanelHandler) HandleBroadcastMessage(chatID int64) {
	text := `📢 ارسال پیام همگانی

لطفاً پیام مورد نظر خود را برای ارسال به تمامی کاربران تایپ کنید:

⚠️ توجه: 
• پیام به تمامی مالکان فروشگاه ارسال خواهد شد
• از ارسال پیام‌های تبلیغاتی مکرر خودداری کنید
• پیام باید مربوط به سیستم یا اطلاعیه‌های مهم باشد

💡 برای لغو عملیات /cancel بفرستید.`

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ لغو", "admin_panel"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	aph.bot.Send(msg)
}

func (aph *AdminPanelHandler) HandleStoreDetails(chatID int64, storeID uint) {
	var store models.Store
	err := aph.db.Preload("Products").Preload("Orders").First(&store, storeID).Error
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ فروشگاه یافت نشد.")
		aph.bot.Send(msg)
		return
	}

	// Calculate store statistics
	var totalRevenue, todayRevenue float64
	var todayOrders int64

	aph.db.Model(&models.Order{}).Where("store_id = ?", storeID).Select("COALESCE(SUM(total_amount), 0)").Scan(&totalRevenue)
	
	today := time.Now().Format("2006-01-02")
	aph.db.Model(&models.Order{}).Where("store_id = ? AND DATE(created_at) = ?", storeID, today).Count(&todayOrders)
	aph.db.Model(&models.Order{}).Where("store_id = ? AND DATE(created_at) = ?", storeID, today).Select("COALESCE(SUM(total_amount), 0)").Scan(&todayRevenue)

	limits := aph.subscription.GetPlanLimits(store.PlanType)
	commission := totalRevenue * limits.CommissionRate

	detailsText := fmt.Sprintf(`🏪 جزئیات فروشگاه %s

👤 اطلاعات مالک:
• شناسه: %d
• نام فروشگاه: %s

💎 اطلاعات پلن:
• نوع پلن: %s
• تاریخ انقضا: %s
• وضعیت: %s
• کارمزد: %.1f%%

📊 آمار فروش:
• کل محصولات: %d
• کل سفارش‌ها: %d
• کل فروش: %,.0f تومان
• کارمزد سیستم: %,.0f تومان
• سفارش‌های امروز: %d
• فروش امروز: %,.0f تومان

🤖 وضعیت ربات:
• توکن: %s
• نام کاربری: %s
• وضعیت: %s

📅 تاریخ ثبت: %s`,
		store.StoreName,
		store.OwnerChatID,
		store.StoreName,
		store.PlanType,
		store.ExpiresAt.Format("2006/01/02"),
		func() string {
			if store.IsActive {
				return "✅ فعال"
			}
			return "❌ غیرفعال"
		}(),
		limits.CommissionRate*100,
		len(store.Products),
		len(store.Orders),
		totalRevenue,
		commission,
		todayOrders,
		todayRevenue,
		func() string {
			if store.BotToken != "" {
				return "تنظیم شده"
			}
			return "تنظیم نشده"
		}(),
		store.BotUsername,
		func() string {
			isRunning, _ := aph.botManager.GetBotStatus(store.ID)
			if isRunning {
				return "✅ در حال اجرا"
			}
			return "❌ متوقف"
		}(),
		store.CreatedAt.Format("2006/01/02"),
	)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ فعال‌سازی", fmt.Sprintf("activate_store_%d", storeID)),
			tgbotapi.NewInlineKeyboardButtonData("❌ غیرفعال‌سازی", fmt.Sprintf("deactivate_store_%d", storeID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 راه‌اندازی ربات", fmt.Sprintf("restart_store_bot_%d", storeID)),
			tgbotapi.NewInlineKeyboardButtonData("⏹ توقف ربات", fmt.Sprintf("stop_store_bot_%d", storeID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📧 پیام به مالک", fmt.Sprintf("message_owner_%d", storeID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 بازگشت", "admin_stores"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, detailsText)
	msg.ReplyMarkup = keyboard
	aph.bot.Send(msg)
}

func (aph *AdminPanelHandler) HandleCallback(callback *tgbotapi.CallbackQuery) {
	chatID := callback.Message.Chat.ID
	data := callback.Data

	// Answer callback query
	aph.bot.Request(tgbotapi.NewCallback(callback.ID, ""))

	switch {
	case data == "admin_stores":
		aph.HandleStoreManagement(chatID)
	case data == "admin_payments":
		aph.HandlePaymentManagement(chatID)
	case data == "admin_financial":
		aph.HandleFinancialReport(chatID)
	case data == "admin_broadcast":
		aph.HandleBroadcastMessage(chatID)
	case strings.HasPrefix(data, "admin_store_"):
		storeIDStr := strings.TrimPrefix(data, "admin_store_")
		storeID, err := strconv.ParseUint(storeIDStr, 10, 32)
		if err == nil {
			aph.HandleStoreDetails(chatID, uint(storeID))
		}
	case strings.HasPrefix(data, "approve_payment_"):
		paymentIDStr := strings.TrimPrefix(data, "approve_payment_")
		aph.handlePaymentApproval(chatID, paymentIDStr, "approved")
	case strings.HasPrefix(data, "reject_payment_"):
		paymentIDStr := strings.TrimPrefix(data, "reject_payment_")
		aph.handlePaymentApproval(chatID, paymentIDStr, "rejected")
	case strings.HasPrefix(data, "activate_store_"):
		storeIDStr := strings.TrimPrefix(data, "activate_store_")
		aph.handleStoreActivation(chatID, storeIDStr, true)
	case strings.HasPrefix(data, "deactivate_store_"):
		storeIDStr := strings.TrimPrefix(data, "deactivate_store_")
		aph.handleStoreActivation(chatID, storeIDStr, false)
	}
}

func (aph *AdminPanelHandler) handlePaymentApproval(chatID int64, paymentIDStr, status string) {
	paymentID, err := strconv.ParseUint(paymentIDStr, 10, 32)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ خطا در شناسایی پرداخت.")
		aph.bot.Send(msg)
		return
	}

	// Update payment status
	var payment models.Payment
	err = aph.db.Preload("Store").First(&payment, paymentID).Error
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ پرداخت یافت نشد.")
		aph.bot.Send(msg)
		return
	}

	err = aph.db.Model(&payment).Update("status", status).Error
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ خطا در بروزرسانی وضعیت پرداخت.")
		aph.bot.Send(msg)
		return
	}

	// If approved, extend subscription
	if status == "approved" {
		err = aph.subscription.ExtendSubscription(payment.StoreID, payment.PlanType, 1)
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "⚠️ پرداخت تایید شد اما خطا در تمدید اشتراک رخ داد.")
			aph.bot.Send(msg)
			return
		}

		// Notify store owner
		notificationText := fmt.Sprintf(`✅ پرداخت شما تایید شد!

💎 پلن: %s
⏰ مدت اعتبار: 1 ماه
📅 انقضا: %s

فروشگاه شما فعال است و می‌توانید از تمامی امکانات استفاده کنید.`, payment.PlanType, time.Now().AddDate(0, 1, 0).Format("2006/01/02"))

		ownerMsg := tgbotapi.NewMessage(payment.Store.OwnerChatID, notificationText)
		aph.bot.Send(ownerMsg)
	} else {
		// Notify store owner of rejection
		notificationText := `❌ پرداخت شما تایید نشد.

لطفاً با پشتیبانی تماس بگیرید تا مشکل بررسی شود.`

		ownerMsg := tgbotapi.NewMessage(payment.Store.OwnerChatID, notificationText)
		aph.bot.Send(ownerMsg)
	}

	statusText := "تایید"
	if status == "rejected" {
		statusText = "رد"
	}

	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("✅ پرداخت با موفقیت %s شد.", statusText))
	aph.bot.Send(msg)
}

func (aph *AdminPanelHandler) handleStoreActivation(chatID int64, storeIDStr string, activate bool) {
	storeID, err := strconv.ParseUint(storeIDStr, 10, 32)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ خطا در شناسایی فروشگاه.")
		aph.bot.Send(msg)
		return
	}

	var store models.Store
	err = aph.db.First(&store, storeID).Error
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ فروشگاه یافت نشد.")
		aph.bot.Send(msg)
		return
	}

	err = aph.db.Model(&store).Update("is_active", activate).Error
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ خطا در بروزرسانی وضعیت فروشگاه.")
		aph.bot.Send(msg)
		return
	}

	statusText := "فعال"
	if !activate {
		statusText = "غیرفعال"
		// Stop the bot if deactivating
		aph.botManager.StopSubBot(uint(storeID))
	} else {
		// Start the bot if activating
		aph.botManager.StartSubBot(uint(storeID))
	}

	// Notify store owner
	notificationText := fmt.Sprintf(`🔔 وضعیت فروشگاه شما تغییر کرد

🏪 فروشگاه: %s
📊 وضعیت جدید: %s

برای اطلاعات بیشتر با پشتیبانی تماس بگیرید.`, store.StoreName, statusText)

	ownerMsg := tgbotapi.NewMessage(store.OwnerChatID, notificationText)
	aph.bot.Send(ownerMsg)

	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("✅ فروشگاه با موفقیت %s شد.", statusText))
	aph.bot.Send(msg)
}