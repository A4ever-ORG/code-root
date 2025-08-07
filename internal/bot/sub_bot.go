package bot

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"telegram-store-hub/internal/models"
	"telegram-store-hub/internal/services"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

type SubBot struct {
	bot           *tgbotapi.BotAPI
	db            *gorm.DB
	store         *models.Store
	storeManager  *services.StoreManagerService
	subscription  *services.SubscriptionService
}

func NewSubBot(token string, db *gorm.DB, store *models.Store) (*SubBot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	bot.Debug = false
	log.Printf("🤖 Sub-bot for store %s (%s) is ready", store.StoreName, bot.Self.UserName)

	return &SubBot{
		bot:           bot,
		db:            db,
		store:         store,
		storeManager:  services.NewStoreManagerService(db),
		subscription:  services.NewSubscriptionService(db),
	}, nil
}

func (sb *SubBot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := sb.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			sb.handleMessage(update.Message)
		} else if update.CallbackQuery != nil {
			sb.handleCallback(update.CallbackQuery)
		}
	}
}

func (sb *SubBot) handleMessage(message *tgbotapi.Message) {
	chatID := message.Chat.ID
	text := message.Text

	switch {
	case text == "/start":
		sb.sendWelcome(chatID)
	case text == "/products" || text == "🛍 محصولات":
		sb.showProducts(chatID)
	case text == "/cart" || text == "🛒 سبد خرید":
		sb.showCart(chatID)
	case text == "/orders" || text == "📋 سفارش‌های من":
		sb.showUserOrders(chatID)
	case text == "/contact" || text == "📞 تماس با ما":
		sb.showContact(chatID)
	default:
		sb.sendMainMenu(chatID)
	}
}

func (sb *SubBot) sendWelcome(chatID int64) {
	welcomeText := sb.store.WelcomeMessage
	if welcomeText == "" {
		welcomeText = fmt.Sprintf(`🌟 به فروشگاه %s خوش آمدید! 🌟

برای مشاهده محصولات و خرید از دکمه‌های زیر استفاده کنید:`, sb.store.StoreName)
	}

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🛍 محصولات"),
			tgbotapi.NewKeyboardButton("🛒 سبد خرید"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📋 سفارش‌های من"),
			tgbotapi.NewKeyboardButton("📞 تماس با ما"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, welcomeText)
	msg.ReplyMarkup = keyboard

	sb.bot.Send(msg)
}

func (sb *SubBot) showProducts(chatID int64) {
	products, err := sb.storeManager.GetProducts(sb.store.ID)
	if err != nil {
		sb.sendError(chatID, "خطا در دریافت محصولات")
		return
	}

	if len(products) == 0 {
		msg := tgbotapi.NewMessage(chatID, "❌ هیچ محصولی در حال حاضر موجود نیست.")
		sb.bot.Send(msg)
		return
	}

	text := fmt.Sprintf("🛍 محصولات فروشگاه %s:\n\n", sb.store.StoreName)

	var keyboard [][]tgbotapi.InlineKeyboardButton
	for i, product := range products {
		if product.IsActive {
			text += fmt.Sprintf("%d. %s\n💰 قیمت: %,.0f تومان\n📝 %s\n\n",
				i+1, product.Name, product.Price, product.Description)

			keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					fmt.Sprintf("🛒 خرید %s", product.Name),
					fmt.Sprintf("buy_%d", product.ID),
				),
			))
		}
	}

	msg := tgbotapi.NewMessage(chatID, text)
	if len(keyboard) > 0 {
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
	}

	sb.bot.Send(msg)
}

func (sb *SubBot) showCart(chatID int64) {
	// For now, we'll show a simple cart message
	// In a full implementation, you'd store cart items in a session or database
	text := `🛒 سبد خرید شما

برای افزودن محصول به سبد خرید، از بخش محصولات استفاده کنید.

💡 راهنما: پس از انتخاب محصولات، می‌توانید سفارش خود را نهایی کنید.`

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🛍 مشاهده محصولات", "show_products"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard

	sb.bot.Send(msg)
}

func (sb *SubBot) showUserOrders(chatID int64) {
	// Get user's orders
	var orders []models.Order
	err := sb.db.Preload("Products").Where("store_id = ? AND customer_id = ?", sb.store.ID, chatID).Find(&orders).Error
	if err != nil {
		sb.sendError(chatID, "خطا در دریافت سفارش‌ها")
		return
	}

	if len(orders) == 0 {
		msg := tgbotapi.NewMessage(chatID, "📋 شما هنوز سفارشی ثبت نکرده‌اید.")
		sb.bot.Send(msg)
		return
	}

	text := "📋 سفارش‌های شما:\n\n"
	for i, order := range orders {
		statusEmoji := sb.getStatusEmoji(order.Status)
		text += fmt.Sprintf("%d. سفارش #%d\n%s وضعیت: %s\n💰 مبلغ: %,.0f تومان\n📅 تاریخ: %s\n\n",
			i+1, order.ID, statusEmoji, order.Status, order.TotalAmount, order.CreatedAt.Format("2006/01/02"))
	}

	msg := tgbotapi.NewMessage(chatID, text)
	sb.bot.Send(msg)
}

func (sb *SubBot) showContact(chatID int64) {
	contactText := fmt.Sprintf(`📞 تماس با فروشگاه %s

برای تماس با ما از راه‌های زیر استفاده کنید:

📧 پشتیبانی: از طریق همین ربات پیام بفرستید
⏰ ساعت کاری: 9 صبح تا 21 شب
📱 پاسخگویی: حداکثر 2 ساعت

💬 برای ارسال پیام، کافی است متن خود را بنویسید.`, sb.store.StoreName)

	msg := tgbotapi.NewMessage(chatID, contactText)
	sb.bot.Send(msg)
}

func (sb *SubBot) handleCallback(callback *tgbotapi.CallbackQuery) {
	chatID := callback.Message.Chat.ID
	data := callback.Data

	// Answer callback query
	sb.bot.Request(tgbotapi.NewCallback(callback.ID, ""))

	switch {
	case strings.HasPrefix(data, "buy_"):
		productIDStr := strings.TrimPrefix(data, "buy_")
		productID, err := strconv.ParseUint(productIDStr, 10, 32)
		if err != nil {
			sb.sendError(chatID, "خطا در شناسایی محصول")
			return
		}
		sb.handleProductPurchase(chatID, uint(productID))
	case data == "show_products":
		sb.showProducts(chatID)
	case data == "confirm_order":
		sb.handleOrderConfirmation(chatID)
	}
}

func (sb *SubBot) handleProductPurchase(chatID int64, productID uint) {
	var product models.Product
	err := sb.db.First(&product, productID).Error
	if err != nil {
		sb.sendError(chatID, "محصول یافت نشد")
		return
	}

	purchaseText := fmt.Sprintf(`🛒 خرید محصول

📦 محصول: %s
💰 قیمت: %,.0f تومان
📝 توضیحات: %s

آیا مایل به خرید این محصول هستید؟`, product.Name, product.Price, product.Description)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ تایید خرید", fmt.Sprintf("confirm_buy_%d", productID)),
			tgbotapi.NewInlineKeyboardButtonData("❌ انصراف", "cancel_buy"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, purchaseText)
	msg.ReplyMarkup = keyboard

	sb.bot.Send(msg)
}

func (sb *SubBot) handleOrderConfirmation(chatID int64) {
	// This is a simplified order confirmation
	// In a real implementation, you'd handle payment processing
	
	confirmText := `✅ سفارش شما با موفقیت ثبت شد!

📋 شماره سفارش: #` + fmt.Sprintf("%d", chatID) + `
⏰ زمان تحویل: 1-2 روز کاری
📞 برای پیگیری با پشتیبانی تماس بگیرید.

🙏 از خرید شما متشکریم!`

	msg := tgbotapi.NewMessage(chatID, confirmText)
	sb.bot.Send(msg)

	// Notify store owner
	sb.notifyStoreOwner(chatID)
}

func (sb *SubBot) notifyStoreOwner(customerID int64) {
	notificationText := fmt.Sprintf(`🔔 سفارش جدید!

🏪 فروشگاه: %s
👤 مشتری: %d
📅 زمان: الان

برای مشاهده جزئیات به پنل مدیریت مراجعه کنید.`, sb.store.StoreName, customerID)

	msg := tgbotapi.NewMessage(sb.store.OwnerChatID, notificationText)
	sb.bot.Send(msg)
}

func (sb *SubBot) sendMainMenu(chatID int64) {
	sb.sendWelcome(chatID)
}

func (sb *SubBot) sendError(chatID int64, errorMsg string) {
	msg := tgbotapi.NewMessage(chatID, "❌ "+errorMsg)
	sb.bot.Send(msg)
}

func (sb *SubBot) getStatusEmoji(status string) string {
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