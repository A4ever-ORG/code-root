package services

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"telegram-store-hub/internal/messages"
	"telegram-store-hub/internal/models"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

// SellerPanelService manages seller panel operations
type SellerPanelService struct {
	bot               *tgbotapi.BotAPI
	db                *gorm.DB
	storeManager      *StoreManagerService
	productService    *ProductService
	orderService      *OrderService
	sessionService    *SessionService
}

// NewSellerPanelService creates a new seller panel service
func NewSellerPanelService(
	bot *tgbotapi.BotAPI,
	db *gorm.DB,
	storeManager *StoreManagerService,
	productService *ProductService,
	orderService *OrderService,
	sessionService *SessionService,
) *SellerPanelService {
	return &SellerPanelService{
		bot:            bot,
		db:             db,
		storeManager:   storeManager,
		productService: productService,
		orderService:   orderService,
		sessionService: sessionService,
	}
}

// ShowStoreManagement displays the main store management panel
func (s *SellerPanelService) ShowStoreManagement(chatID int64) {
	// Get user's store
	store, err := s.getUserStore(chatID)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, messages.ErrorNoStore)
		s.bot.Send(msg)
		return
	}

	// Get store statistics
	stats, err := s.getStoreStats(store.ID)
	if err != nil {
		log.Printf("Error getting store stats: %v", err)
		stats = map[string]int64{
			"products": 0,
			"orders":   0,
			"sales":    0,
		}
	}

	// Calculate remaining days
	daysRemaining := int(time.Until(store.ExpiresAt).Hours() / 24)
	if daysRemaining < 0 {
		daysRemaining = 0
	}

	text := fmt.Sprintf(`🏪 پنل مدیریت فروشگاه "%s"

📊 آمار فروشگاه:
• محصولات: %d
• سفارشات: %d
• فروش کل: %s تومان

📋 اطلاعات پلن:
• نوع: %s
• محصولات مجاز: %s
• کارمزد: %d%%
• باقیمانده: %d روز

⚠️ %s`,
		store.Name,
		stats["products"],
		stats["orders"],
		formatPrice(stats["sales"]),
		string(store.PlanType),
		func() string {
			if store.ProductLimit == -1 {
				return "نامحدود"
			}
			return fmt.Sprintf("%d", store.ProductLimit)
		}(),
		store.CommissionRate,
		daysRemaining,
		func() string {
			if daysRemaining <= 7 {
				return "پلن شما به زودی منقضی می‌شود!"
			}
			if !store.IsActive {
				return "فروشگاه غیرفعال است!"
			}
			return "فروشگاه فعال است"
		}())

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ افزودن محصول", "add_product"),
			tgbotapi.NewInlineKeyboardButtonData("📦 محصولات", "list_products"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 سفارشات", "view_orders"),
			tgbotapi.NewInlineKeyboardButtonData("📊 گزارش فروش", "sales_report"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⚙️ تنظیمات", "store_settings"),
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

// StartProductAddition starts the product addition process
func (s *SellerPanelService) StartProductAddition(chatID int64) {
	// Get user's store
	store, err := s.getUserStore(chatID)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, messages.ErrorNoStore)
		s.bot.Send(msg)
		return
	}

	// Check if store is active
	if !store.IsActive {
		msg := tgbotapi.NewMessage(chatID, "❌ فروشگاه شما غیرفعال است. لطفاً پلن خود را تمدید کنید.")
		s.bot.Send(msg)
		return
	}

	// Check product limit
	currentProducts, err := s.productService.GetStoreProductCount(store.ID)
	if err != nil {
		log.Printf("Error getting product count: %v", err)
		msg := tgbotapi.NewMessage(chatID, messages.ErrorDatabaseError)
		s.bot.Send(msg)
		return
	}

	if store.ProductLimit > 0 && currentProducts >= store.ProductLimit {
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("❌ شما به حد مجاز محصولات (%d) رسیده‌اید. برای افزودن محصول بیشتر، پلن خود را ارتقا دهید.", store.ProductLimit))
		s.bot.Send(msg)
		return
	}

	// Start product addition process
	productData := ProductData{
		Step:    1,
		StoreID: store.ID,
	}

	s.sessionService.SetUserState(chatID, messages.StateWaitingProductName, productData)

	msg := tgbotapi.NewMessage(chatID, "📝 لطفاً نام محصول را وارد کنید:")
	s.bot.Send(msg)
}

// ShowProductList displays all products for the store
func (s *SellerPanelService) ShowProductList(chatID int64) {
	// Get user's store
	store, err := s.getUserStore(chatID)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, messages.ErrorNoStore)
		s.bot.Send(msg)
		return
	}

	products, err := s.productService.GetStoreProducts(store.ID)
	if err != nil {
		log.Printf("Error getting products: %v", err)
		msg := tgbotapi.NewMessage(chatID, messages.ErrorDatabaseError)
		s.bot.Send(msg)
		return
	}

	if len(products) == 0 {
		text := "📦 هیچ محصولی در فروشگاه شما موجود نیست.\n\nبرای افزودن محصول جدید روی دکمه زیر کلیک کنید:"
		
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("➕ افزودن محصول", "add_product"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🔙 بازگشت", "manage_store"),
			),
		)

		msg := tgbotapi.NewMessage(chatID, text)
		msg.ReplyMarkup = keyboard
		s.bot.Send(msg)
		return
	}

	text := fmt.Sprintf("📦 محصولات فروشگاه (%d محصول):\n\n", len(products))

	var keyboard [][]tgbotapi.InlineKeyboardButton
	
	for i, product := range products {
		status := "🟢 فعال"
		if !product.IsAvailable {
			status = "🔴 غیرفعال"
		}

		text += fmt.Sprintf("%d. %s\n💰 %s تومان %s\n\n",
			i+1,
			product.Name,
			formatPrice(product.Price),
			status)

		// Add product management buttons (2 per row)
		if i%2 == 0 {
			if i+1 < len(products) {
				keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("✏️ %s", truncateString(product.Name, 10)), fmt.Sprintf("product_%d", product.ID)),
					tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("✏️ %s", truncateString(products[i+1].Name, 10)), fmt.Sprintf("product_%d", products[i+1].ID)),
				))
			} else {
				keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("✏️ %s", truncateString(product.Name, 10)), fmt.Sprintf("product_%d", product.ID)),
				))
			}
		}
	}

	// Add navigation buttons
	keyboard = append(keyboard,
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ افزودن محصول", "add_product"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 بازگشت", "manage_store"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)

	s.bot.Send(msg)
}

// ShowOrderList displays orders for the store
func (s *SellerPanelService) ShowOrderList(chatID int64) {
	// Get user's store
	store, err := s.getUserStore(chatID)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, messages.ErrorNoStore)
		s.bot.Send(msg)
		return
	}

	orders, err := s.orderService.GetStoreOrders(store.ID, 10) // Get last 10 orders
	if err != nil {
		log.Printf("Error getting orders: %v", err)
		msg := tgbotapi.NewMessage(chatID, messages.ErrorDatabaseError)
		s.bot.Send(msg)
		return
	}

	if len(orders) == 0 {
		text := "📋 هیچ سفارشی برای فروشگاه شما ثبت نشده است."
		
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🔙 بازگشت", "manage_store"),
			),
		)

		msg := tgbotapi.NewMessage(chatID, text)
		msg.ReplyMarkup = keyboard
		s.bot.Send(msg)
		return
	}

	text := fmt.Sprintf("📋 آخرین سفارشات فروشگاه (%d سفارش):\n\n", len(orders))

	var keyboard [][]tgbotapi.InlineKeyboardButton

	for i, order := range orders {
		statusEmoji := "⏳"
		switch order.Status {
		case models.OrderStatusCompleted:
			statusEmoji = "✅"
		case models.OrderStatusCancelled:
			statusEmoji = "❌"
		case models.OrderStatusShipped:
			statusEmoji = "🚚"
		}

		paymentStatus := "❌ پرداخت نشده"
		if order.PaymentStatus == models.PaymentStatusPaid {
			paymentStatus = "✅ پرداخت شده"
		}

		text += fmt.Sprintf("%d. سفارش #%d\n%s %s\n💰 %s تومان\n%s\n📅 %s\n\n",
			i+1,
			order.ID,
			statusEmoji,
			string(order.Status),
			formatPrice(order.TotalAmount),
			paymentStatus,
			order.CreatedAt.Format("2006/01/02 15:04"))

		// Add order management button
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("📋 سفارش #%d", order.ID), fmt.Sprintf("order_%d", order.ID)),
		))
	}

	// Add navigation buttons
	keyboard = append(keyboard,
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 بازگشت", "manage_store"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)

	s.bot.Send(msg)
}

// ShowSalesReport displays sales report
func (s *SellerPanelService) ShowSalesReport(chatID int64) {
	// Get user's store
	store, err := s.getUserStore(chatID)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, messages.ErrorNoStore)
		s.bot.Send(msg)
		return
	}

	// Get sales statistics
	salesData, err := s.orderService.GetStoreSalesReport(store.ID)
	if err != nil {
		log.Printf("Error getting sales report: %v", err)
		msg := tgbotapi.NewMessage(chatID, messages.ErrorDatabaseError)
		s.bot.Send(msg)
		return
	}

	text := fmt.Sprintf(`📊 گزارش فروش فروشگاه "%s"

💰 فروش کل: %s تومان
💳 درآمد خالص: %s تومان
📋 تعداد سفارشات: %d
✅ سفارشات تکمیل شده: %d
⏳ سفارشات در حال انجام: %d

📈 فروش این ماه: %s تومان
📅 فروش امروز: %s تومان

💵 کارمزد پلن: %d%%
🏦 کارمزد پرداختی: %s تومان`,
		store.Name,
		formatPrice(salesData.TotalSales),
		formatPrice(salesData.NetIncome),
		salesData.TotalOrders,
		salesData.CompletedOrders,
		salesData.PendingOrders,
		formatPrice(salesData.MonthlySales),
		formatPrice(salesData.DailySales),
		store.CommissionRate,
		formatPrice(salesData.CommissionPaid))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 جزئیات سفارشات", "view_orders"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 بازگشت", "manage_store"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard

	s.bot.Send(msg)
}

// ShowStoreSettings displays store settings
func (s *SellerPanelService) ShowStoreSettings(chatID int64) {
	// Get user's store
	store, err := s.getUserStore(chatID)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, messages.ErrorNoStore)
		s.bot.Send(msg)
		return
	}

	text := fmt.Sprintf(`⚙️ تنظیمات فروشگاه "%s"

📝 نام: %s
📄 توضیحات: %s
🏷️ نوع پلن: %s
🤖 ربات فروشگاه: %s
📱 وضعیت: %s

📊 محدودیت‌ها:
• محصولات: %s
• کارمزد: %d%%

📅 تاریخ انقضا: %s`,
		store.Name,
		store.Name,
		func() string {
			if store.Description == "" {
				return "توضیحاتی ثبت نشده"
			}
			return store.Description
		}(),
		string(store.PlanType),
		func() string {
			if store.BotUsername == "" {
				return "در حال آماده‌سازی..."
			}
			return "@" + store.BotUsername
		}(),
		func() string {
			if store.IsActive {
				return "🟢 فعال"
			}
			return "🔴 غیرفعال"
		}(),
		func() string {
			if store.ProductLimit == -1 {
				return "نامحدود"
			}
			return fmt.Sprintf("%d", store.ProductLimit)
		}(),
		store.CommissionRate,
		store.ExpiresAt.Format("2006/01/02"))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✏️ ویرایش نام", "edit_store_name"),
			tgbotapi.NewInlineKeyboardButtonData("📝 ویرایش توضیحات", "edit_store_desc"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 تمدید پلن", "renew_plan"),
			tgbotapi.NewInlineKeyboardButtonData("⬆️ ارتقای پلن", "upgrade_plan"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 بازگشت", "manage_store"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard

	s.bot.Send(msg)
}

// Helper methods

func (s *SellerPanelService) getUserStore(chatID int64) (*models.Store, error) {
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

func (s *SellerPanelService) getStoreStats(storeID uint) (map[string]int64, error) {
	stats := make(map[string]int64)

	// Count products
	var productCount int64
	if err := s.db.Model(&models.Product{}).Where("store_id = ?", storeID).Count(&productCount).Error; err != nil {
		return nil, err
	}
	stats["products"] = productCount

	// Count orders
	var orderCount int64
	if err := s.db.Model(&models.Order{}).Where("store_id = ?", storeID).Count(&orderCount).Error; err != nil {
		return nil, err
	}
	stats["orders"] = orderCount

	// Calculate total sales
	var totalSales int64
	if err := s.db.Model(&models.Order{}).Where("store_id = ? AND payment_status = ?", storeID, models.PaymentStatusPaid).Select("SUM(total_amount)").Scan(&totalSales).Error; err != nil {
		return nil, err
	}
	stats["sales"] = totalSales

	return stats, nil
}

// formatPrice formats price with thousand separators
func formatPrice(price int64) string {
	priceStr := strconv.FormatInt(price, 10)
	
	if len(priceStr) <= 3 {
		return priceStr
	}

	// Add thousand separators
	var result strings.Builder
	for i, char := range priceStr {
		if i > 0 && (len(priceStr)-i)%3 == 0 {
			result.WriteRune(',')
		}
		result.WriteRune(char)
	}

	return result.String()
}

// truncateString truncates string to specified length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}