package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"telegram-store-hub/internal/models"
	"telegram-store-hub/internal/services"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

type SellerPanelHandler struct {
	bot          *tgbotapi.BotAPI
	db           *gorm.DB
	storeManager *services.StoreManagerService
	botManager   *services.BotManagerService
	subscription *services.SubscriptionService
}

func NewSellerPanelHandler(bot *tgbotapi.BotAPI, db *gorm.DB) *SellerPanelHandler {
	return &SellerPanelHandler{
		bot:          bot,
		db:           db,
		storeManager: services.NewStoreManagerService(db),
		botManager:   services.NewBotManagerService(db),
		subscription: services.NewSubscriptionService(db),
	}
}

func (sph *SellerPanelHandler) HandleAddProduct(chatID int64) {
	store, err := sph.storeManager.GetStoreByOwner(chatID)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ ابتدا باید فروشگاه خود را ثبت کنید.")
		sph.bot.Send(msg)
		return
	}

	// Check if can add more products
	canAdd, err := sph.subscription.CanAddProduct(store.ID)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ خطا در بررسی محدودیت محصولات.")
		sph.bot.Send(msg)
		return
	}

	if !canAdd {
		limits := sph.subscription.GetPlanLimits(store.PlanType)
		msg := tgbotapi.NewMessage(chatID, 
			fmt.Sprintf("❌ شما به حداکثر تعداد محصولات مجاز (%d) رسیده‌اید. برای افزودن محصول بیشتر، پلن خود را ارتقا دهید.", limits.MaxProducts))
		sph.bot.Send(msg)
		return
	}

	text := `➕ افزودن محصول جدید

لطفاً اطلاعات محصول را به ترتیب زیر ارسال کنید:

📝 فرمت: نام محصول | قیمت | توضیحات

مثال:
گوشی سامسونگ A54 | 15000000 | گوشی هوشمند با کیفیت عالی، 8GB RAM

⚠️ توجه: قیمت را به تومان وارد کنید.`

	msg := tgbotapi.NewMessage(chatID, text)
	sph.bot.Send(msg)

	// Set user state to adding product (in a real implementation, you'd store this in a session)
}

func (sph *SellerPanelHandler) HandleProductList(chatID int64) {
	store, err := sph.storeManager.GetStoreByOwner(chatID)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ فروشگاهی یافت نشد.")
		sph.bot.Send(msg)
		return
	}

	products, err := sph.storeManager.GetProducts(store.ID)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ خطا در دریافت محصولات.")
		sph.bot.Send(msg)
		return
	}

	if len(products) == 0 {
		text := `📦 لیست محصولات

هیچ محصولی یافت نشد.

برای افزودن محصول جدید از دکمه "➕ افزودن محصول" استفاده کنید.`

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("➕ افزودن محصول", "add_product"),
			),
		)

		msg := tgbotapi.NewMessage(chatID, text)
		msg.ReplyMarkup = keyboard
		sph.bot.Send(msg)
		return
	}

	text := fmt.Sprintf("📦 لیست محصولات فروشگاه %s:\n\n", store.StoreName)
	
	var keyboard [][]tgbotapi.InlineKeyboardButton
	for i, product := range products {
		status := "✅ فعال"
		if !product.IsActive {
			status = "❌ غیرفعال"
		}

		text += fmt.Sprintf("%d. %s\n💰 قیمت: %,.0f تومان\n📊 وضعیت: %s\n📅 تاریخ: %s\n\n",
			i+1, product.Name, product.Price, status, product.CreatedAt.Format("2006/01/02"))

		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✏️ ویرایش", fmt.Sprintf("edit_product_%d", product.ID)),
			tgbotapi.NewInlineKeyboardButtonData("🗑 حذف", fmt.Sprintf("delete_product_%d", product.ID)),
		))
	}

	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("➕ افزودن محصول", "add_product"),
	))

	msg := tgbotapi.NewMessage(chatID, text)
	if len(keyboard) > 0 {
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
	}
	sph.bot.Send(msg)
}

func (sph *SellerPanelHandler) HandleOrdersList(chatID int64) {
	store, err := sph.storeManager.GetStoreByOwner(chatID)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ فروشگاهی یافت نشد.")
		sph.bot.Send(msg)
		return
	}

	orders, err := sph.storeManager.GetOrders(store.ID)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ خطا در دریافت سفارش‌ها.")
		sph.bot.Send(msg)
		return
	}

	if len(orders) == 0 {
		msg := tgbotapi.NewMessage(chatID, "🛒 هیچ سفارشی یافت نشد.")
		sph.bot.Send(msg)
		return
	}

	text := fmt.Sprintf("🛒 سفارش‌های فروشگاه %s:\n\n", store.StoreName)
	
	var keyboard [][]tgbotapi.InlineKeyboardButton
	for i, order := range orders {
		statusEmoji := sph.getStatusEmoji(order.Status)
		
		text += fmt.Sprintf("%d. سفارش #%d\n%s وضعیت: %s\n👤 مشتری: %s\n💰 مبلغ: %,.0f تومان\n📅 تاریخ: %s\n\n",
			i+1, order.ID, statusEmoji, order.Status, order.CustomerName, order.TotalAmount, order.CreatedAt.Format("2006/01/02"))

		if order.Status == "pending" {
			keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("✅ تایید", fmt.Sprintf("confirm_order_%d", order.ID)),
				tgbotapi.NewInlineKeyboardButtonData("❌ لغو", fmt.Sprintf("cancel_order_%d", order.ID)),
			))
		}
	}

	msg := tgbotapi.NewMessage(chatID, text)
	if len(keyboard) > 0 {
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
	}
	sph.bot.Send(msg)
}

func (sph *SellerPanelHandler) HandleSalesReport(chatID int64) {
	store, err := sph.storeManager.GetStoreByOwner(chatID)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ فروشگاهی یافت نشد.")
		sph.bot.Send(msg)
		return
	}

	stats, err := sph.storeManager.GetStoreStats(store.ID)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ خطا در دریافت گزارش.")
		sph.bot.Send(msg)
		return
	}

	limits := sph.subscription.GetPlanLimits(store.PlanType)
	commission := stats["total_revenue"].(float64) * limits.CommissionRate

	reportText := fmt.Sprintf(`📈 گزارش فروش فروشگاه %s

📊 آمار کلی:
📦 تعداد محصولات: %d
🛒 کل سفارش‌ها: %d
💰 کل فروش: %,.0f تومان
🏆 سفارش‌های امروز: %d

💳 محاسبات مالی:
💎 نرخ کارمزد: %.1f%%
🔸 کارمزد: %,.0f تومان
💵 درآمد خالص: %,.0f تومان

📅 بازه زمانی: از ابتدای ثبت فروشگاه تا اکنون`,
		store.StoreName,
		stats["product_count"].(int64),
		stats["order_count"].(int64),
		stats["total_revenue"].(float64),
		stats["today_orders"].(int64),
		limits.CommissionRate*100,
		commission,
		stats["total_revenue"].(float64)-commission,
	)

	msg := tgbotapi.NewMessage(chatID, reportText)
	sph.bot.Send(msg)
}

func (sph *SellerPanelHandler) HandleStoreSettings(chatID int64) {
	store, err := sph.storeManager.GetStoreByOwner(chatID)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ فروشگاهی یافت نشد.")
		sph.bot.Send(msg)
		return
	}

	limits := sph.subscription.GetPlanLimits(store.PlanType)
	botStatus := "❌ غیرفعال"
	if store.BotToken != "" {
		isRunning, _ := sph.botManager.GetBotStatus(store.ID)
		if isRunning {
			botStatus = "✅ فعال"
		} else {
			botStatus = "⏸ متوقف"
		}
	}

	settingsText := fmt.Sprintf(`⚙️ تنظیمات فروشگاه %s

🏪 اطلاعات فروشگاه:
• نام: %s
• پلن: %s
• وضعیت ربات: %s
• انقضا: %s

📊 محدودیت‌های پلن:
• حداکثر محصولات: %s
• کارمزد: %.1f%%
• پیام خوش‌آمدگویی: %s

🤖 اطلاعات ربات:
• نام کاربری: %s
• توکن: %s`,
		store.StoreName,
		store.StoreName,
		store.PlanType,
		botStatus,
		store.ExpiresAt.Format("2006/01/02"),
		func() string {
			if limits.MaxProducts == -1 {
				return "نامحدود"
			}
			return fmt.Sprintf("%d", limits.MaxProducts)
		}(),
		limits.CommissionRate*100,
		func() string {
			if limits.HasWelcomeMsg {
				return "✅ دارد"
			}
			return "❌ ندارد"
		}(),
		store.BotUsername,
		func() string {
			if store.BotToken != "" {
				return "تنظیم شده"
			}
			return "تنظیم نشده"
		}(),
	)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✏️ ویرایش نام", "edit_store_name"),
			tgbotapi.NewInlineKeyboardButtonData("💬 پیام خوش‌آمد", "edit_welcome_msg"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🤖 تنظیم ربات", "setup_bot"),
			tgbotapi.NewInlineKeyboardButtonData("🔄 راه‌اندازی مجدد", "restart_bot"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 بازگشت", "manage_store"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, settingsText)
	msg.ReplyMarkup = keyboard
	sph.bot.Send(msg)
}

func (sph *SellerPanelHandler) HandleRenewPlan(chatID int64) {
	store, err := sph.storeManager.GetStoreByOwner(chatID)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ فروشگاهی یافت نشد.")
		sph.bot.Send(msg)
		return
	}

	renewText := fmt.Sprintf(`🔄 تمدید پلن فروشگاه

📊 وضعیت فعلی:
• پلن: %s
• انقضا: %s
• باقی‌مانده: %d روز

💎 پلن‌های موجود:`, 
		store.PlanType,
		store.ExpiresAt.Format("2006/01/02"),
		int(store.ExpiresAt.Sub(store.ExpiresAt).Hours()/24),
	)

	plans := []models.PlanType{models.PlanFree, models.PlanPro, models.PlanVIP}
	var keyboard [][]tgbotapi.InlineKeyboardButton

	for _, planType := range plans {
		limits := sph.subscription.GetPlanLimits(planType)
		planName := string(planType)
		price := "رایگان"
		if limits.PricePerMonth > 0 {
			price = fmt.Sprintf("%,.0f تومان/ماه", limits.PricePerMonth)
		}

		renewText += fmt.Sprintf("\n\n%s - %s", planName, price)
		for _, feature := range limits.Features {
			renewText += fmt.Sprintf("\n• %s", feature)
		}

		if planType != store.PlanType {
			keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					fmt.Sprintf("📈 ارتقا به %s", planName),
					fmt.Sprintf("upgrade_%s", planType),
				),
			))
		}
	}

	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🔙 بازگشت", "manage_store"),
	))

	msg := tgbotapi.NewMessage(chatID, renewText)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
	sph.bot.Send(msg)
}

func (sph *SellerPanelHandler) HandleCallback(callback *tgbotapi.CallbackQuery) {
	chatID := callback.Message.Chat.ID
	data := callback.Data

	// Answer callback query
	sph.bot.Request(tgbotapi.NewCallback(callback.ID, ""))

	switch {
	case data == "add_product":
		sph.HandleAddProduct(chatID)
	case data == "list_products":
		sph.HandleProductList(chatID)
	case data == "view_orders":
		sph.HandleOrdersList(chatID)
	case data == "sales_report":
		sph.HandleSalesReport(chatID)
	case data == "store_settings":
		sph.HandleStoreSettings(chatID)
	case data == "renew_plan":
		sph.HandleRenewPlan(chatID)
	case strings.HasPrefix(data, "edit_product_"):
		productIDStr := strings.TrimPrefix(data, "edit_product_")
		sph.handleEditProduct(chatID, productIDStr)
	case strings.HasPrefix(data, "delete_product_"):
		productIDStr := strings.TrimPrefix(data, "delete_product_")
		sph.handleDeleteProduct(chatID, productIDStr)
	case strings.HasPrefix(data, "confirm_order_"):
		orderIDStr := strings.TrimPrefix(data, "confirm_order_")
		sph.handleConfirmOrder(chatID, orderIDStr)
	case strings.HasPrefix(data, "upgrade_"):
		planType := strings.TrimPrefix(data, "upgrade_")
		sph.handlePlanUpgrade(chatID, models.PlanType(planType))
	}
}

func (sph *SellerPanelHandler) handleEditProduct(chatID int64, productIDStr string) {
	msg := tgbotapi.NewMessage(chatID, "✏️ ویرایش محصول - این قابلیت به زودی اضافه می‌شود.")
	sph.bot.Send(msg)
}

func (sph *SellerPanelHandler) handleDeleteProduct(chatID int64, productIDStr string) {
	productID, err := strconv.ParseUint(productIDStr, 10, 32)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ خطا در شناسایی محصول.")
		sph.bot.Send(msg)
		return
	}

	err = sph.storeManager.DeleteProduct(uint(productID))
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ خطا در حذف محصول.")
		sph.bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(chatID, "✅ محصول با موفقیت حذف شد.")
	sph.bot.Send(msg)
}

func (sph *SellerPanelHandler) handleConfirmOrder(chatID int64, orderIDStr string) {
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ خطا در شناسایی سفارش.")
		sph.bot.Send(msg)
		return
	}

	err = sph.storeManager.UpdateOrderStatus(uint(orderID), "confirmed")
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ خطا در تایید سفارش.")
		sph.bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(chatID, "✅ سفارش تایید شد و به مشتری اطلاع داده شد.")
	sph.bot.Send(msg)
}

func (sph *SellerPanelHandler) handlePlanUpgrade(chatID int64, planType models.PlanType) {
	limits := sph.subscription.GetPlanLimits(planType)
	
	upgradeText := fmt.Sprintf(`📈 ارتقا به پلن %s

💰 قیمت: %,.0f تومان/ماه

🎯 امکانات:`, planType, limits.PricePerMonth)

	for _, feature := range limits.Features {
		upgradeText += fmt.Sprintf("\n✅ %s", feature)
	}

	upgradeText += "\n\n💳 برای پرداخت به کارت زیر واریز کنید:\n1234-5678-9012-3456\n\nسپس اسکرین‌شات رسید را ارسال کنید."

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ پرداخت کردم", fmt.Sprintf("paid_upgrade_%s", planType)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 بازگشت", "renew_plan"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, upgradeText)
	msg.ReplyMarkup = keyboard
	sph.bot.Send(msg)
}

func (sph *SellerPanelHandler) getStatusEmoji(status string) string {
	switch status {
	case "pending":
		return "⏳"
	case "confirmed":
		return "✅"
	case "shipped":
		return "🚚"
	case "delivered":
		return "📋"
	case "cancelled":
		return "❌"
	default:
		return "❓"
	}
}