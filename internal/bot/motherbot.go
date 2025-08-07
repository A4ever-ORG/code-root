package bot

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"telegram-store-hub/internal/config"
	"telegram-store-hub/internal/messages"
	"telegram-store-hub/internal/models"
	"telegram-store-hub/internal/services"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

type MotherBot struct {
	bot        *tgbotapi.BotAPI
	config     *config.Config
	db         *gorm.DB
	userService *services.UserService
	storeService *services.StoreService
	productService *services.ProductService
	sessionService *services.SessionService
	paymentService *services.PaymentService
	stopChannel chan bool
}

func NewMotherBot(token string, db *gorm.DB) (*MotherBot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	cfg := config.Load()
	
	mb := &MotherBot{
		bot:        bot,
		config:     cfg,
		db:         db,
		userService: services.NewUserService(db),
		storeService: services.NewStoreService(db),
		productService: services.NewProductService(db),
		sessionService: services.NewSessionService(db),
		paymentService: services.NewPaymentService(db),
		stopChannel: make(chan bool),
	}

	log.Printf("✅ Bot @%s initialized successfully", bot.Self.UserName)
	return mb, nil
}

func (mb *MotherBot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := mb.bot.GetUpdatesChan(u)

	for {
		select {
		case update := <-updates:
			go mb.handleUpdate(update)
		case <-mb.stopChannel:
			log.Println("🛑 Mother bot stopped")
			return
		}
	}
}

func (mb *MotherBot) Stop() {
	mb.stopChannel <- true
	mb.bot.StopReceivingUpdates()
}

func (mb *MotherBot) handleUpdate(update tgbotapi.Update) {
	if update.Message != nil {
		mb.handleMessage(update.Message)
	} else if update.CallbackQuery != nil {
		mb.handleCallbackQuery(update.CallbackQuery)
	}
}

func (mb *MotherBot) handleMessage(message *tgbotapi.Message) {
	userID := message.From.ID
	chatID := message.Chat.ID

	// First ensure user exists in database
	user, err := mb.userService.GetOrCreateUser(userID, message.From.UserName, message.From.FirstName, message.From.LastName)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		mb.sendMessage(chatID, messages.ErrorGeneral)
		return
	}

	// Check force join if configured
	if mb.config.ForceJoinChannelID != 0 && !user.IsAdmin {
		if !mb.checkChannelMembership(userID) {
			mb.sendForceJoinMessage(chatID)
			return
		}
	}

	// Get user session
	session, _ := mb.sessionService.GetSession(userID)

	// Handle commands
	if message.IsCommand() {
		mb.handleCommand(message, user)
		return
	}

	// Handle text based on session state
	if session != nil && session.State != "" {
		mb.handleSessionState(message, user, session)
		return
	}

	// Handle regular text/buttons
	mb.handleRegularMessage(message, user)
}

func (mb *MotherBot) handleCommand(message *tgbotapi.Message, user *models.User) {
	chatID := message.Chat.ID
	command := message.Command()

	switch command {
	case "start":
		mb.sendWelcomeMessage(chatID)
	case "panel":
		mb.showStorePanel(chatID, user)
	case "admin":
		if user.IsAdmin {
			mb.showAdminPanel(chatID)
		}
	default:
		mb.sendMessage(chatID, "❌ دستور نامعتبر")
	}
}

func (mb *MotherBot) handleRegularMessage(message *tgbotapi.Message, user *models.User) {
	chatID := message.Chat.ID
	text := message.Text

	switch text {
	case messages.RegisterStoreBtn:
		mb.startStoreRegistration(chatID, user)
	case messages.MyStoresBtn:
		mb.showUserStores(chatID, user)
	case messages.SupportBtn:
		mb.showSupportMenu(chatID)
	case messages.AboutBtn:
		mb.showAboutMessage(chatID)
	case messages.FAQBtn:
		mb.sendMessage(chatID, messages.FAQMessage)
	case messages.ContactUsBtn:
		mb.sendMessage(chatID, "📞 تماس با ما: @support")
	case messages.TelegramSupportBtn:
		mb.sendMessage(chatID, "💬 پشتیبانی تلگرام: @telegram_support")
	case messages.CheckMembershipBtn:
		if mb.checkChannelMembership(user.TelegramID) {
			mb.sendWelcomeMessage(chatID)
		} else {
			mb.sendForceJoinMessage(chatID)
		}
	default:
		mb.sendWelcomeMessage(chatID)
	}
}

func (mb *MotherBot) handleSessionState(message *tgbotapi.Message, user *models.User, session *models.UserSession) {
	chatID := message.Chat.ID

	switch session.State {
	case "store_name":
		mb.handleStoreName(chatID, user, message.Text, session)
	case "store_description":
		mb.handleStoreDescription(chatID, user, message.Text, session)
	case "payment_proof":
		if message.Photo != nil {
			mb.handlePaymentProof(chatID, user, message.Photo, session)
		} else {
			mb.sendMessage(chatID, "لطفاً عکس رسید پرداخت را ارسال کنید")
		}
	case "product_name":
		mb.handleProductName(chatID, user, message.Text, session)
	case "product_description":
		mb.handleProductDescription(chatID, user, message.Text, session)
	case "product_price":
		mb.handleProductPrice(chatID, user, message.Text, session)
	case "product_image":
		if message.Text == "/skip" {
			mb.finalizeProduct(chatID, user, "", session)
		} else if message.Photo != nil {
			photoURL := mb.getPhotoURL(message.Photo)
			mb.finalizeProduct(chatID, user, photoURL, session)
		} else {
			mb.sendMessage(chatID, "لطفاً عکس محصول را ارسال کنید یا /skip بزنید")
		}
	default:
		mb.sessionService.ClearSession(user.TelegramID)
		mb.sendWelcomeMessage(chatID)
	}
}

func (mb *MotherBot) sendWelcomeMessage(chatID int64) {
	keyboard := mb.createMainMenuKeyboard()
	msg := tgbotapi.NewMessage(chatID, messages.WelcomeMessage)
	msg.ReplyMarkup = keyboard
	mb.bot.Send(msg)
}

func (mb *MotherBot) createMainMenuKeyboard() tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(messages.RegisterStoreBtn),
			tgbotapi.NewKeyboardButton(messages.MyStoresBtn),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(messages.SupportBtn),
			tgbotapi.NewKeyboardButton(messages.AboutBtn),
		),
	)
	keyboard.ResizeKeyboard = true
	return keyboard
}

func (mb *MotherBot) startStoreRegistration(chatID int64, user *models.User) {
	// Check if user already has a pending store
	stores, _ := mb.storeService.GetUserStores(user.ID)
	for _, store := range stores {
		if !store.IsActive {
			mb.sendMessage(chatID, "شما یک فروشگاه در انتظار تایید دارید")
			return
		}
	}

	mb.sessionService.SetSession(user.TelegramID, "store_name", "{}")
	mb.sendMessage(chatID, messages.StoreRegistrationStart)
}

func (mb *MotherBot) handleStoreName(chatID int64, user *models.User, storeName string, session *models.UserSession) {
	if len(storeName) < 3 {
		mb.sendMessage(chatID, "نام فروشگاه باید حداقل ۳ کاراکتر باشد")
		return
	}

	sessionData := map[string]interface{}{
		"store_name": storeName,
	}
	data, _ := json.Marshal(sessionData)
	
	mb.sessionService.SetSession(user.TelegramID, "store_description", string(data))
	mb.sendMessage(chatID, fmt.Sprintf(messages.StoreNameReceived, storeName))
}

func (mb *MotherBot) handleStoreDescription(chatID int64, user *models.User, description string, session *models.UserSession) {
	var sessionData map[string]interface{}
	json.Unmarshal([]byte(session.Data), &sessionData)
	
	sessionData["store_description"] = description
	data, _ := json.Marshal(sessionData)
	
	mb.sessionService.SetSession(user.TelegramID, "plan_selection", string(data))
	mb.showPlanSelection(chatID)
}

func (mb *MotherBot) showPlanSelection(chatID int64) {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(messages.FreePlanBtn, "plan_free"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(messages.ProPlanBtn, "plan_pro"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(messages.VIPPlanBtn, "plan_vip"),
		),
	)

	planDetails := fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s", 
		messages.StoreDescriptionReceived,
		messages.FreePlanDetails,
		messages.ProPlanDetails,
		messages.VIPPlanDetails)

	msg := tgbotapi.NewMessage(chatID, planDetails)
	msg.ReplyMarkup = keyboard
	mb.bot.Send(msg)
}

func (mb *MotherBot) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	mb.bot.Send(msg)
}

func (mb *MotherBot) checkChannelMembership(userID int64) bool {
	if mb.config.ForceJoinChannelID == 0 {
		return true
	}

	member, err := mb.bot.GetChatMember(tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID: mb.config.ForceJoinChannelID,
			UserID: userID,
		},
	})

	if err != nil {
		log.Printf("Error checking membership: %v", err)
		return false
	}

	return member.Status == "member" || member.Status == "administrator" || member.Status == "creator"
}

func (mb *MotherBot) sendForceJoinMessage(chatID int64) {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("عضویت در کانال", "https://t.me/"+strings.TrimPrefix(mb.config.ForceJoinChannelUsername, "@")),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(messages.CheckMembershipBtn, "check_membership"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf(messages.ForceJoinMessage, mb.config.ForceJoinChannelUsername))
	msg.ReplyMarkup = keyboard
	mb.bot.Send(msg)
}

func (mb *MotherBot) showUserStores(chatID int64, user *models.User) {
	stores, err := mb.storeService.GetUserStores(user.ID)
	if err != nil || len(stores) == 0 {
		mb.sendMessage(chatID, messages.ErrorNoStore)
		return
	}

	for _, store := range stores {
		storeInfo := fmt.Sprintf(`🏪 %s
📊 پلن: %s
📅 انقضا: %s
🤖 ربات: @%s
✅ وضعیت: %s`,
			store.Name,
			strings.ToUpper(store.PlanType),
			store.ExpiresAt.Format("2006/01/02"),
			store.BotUsername,
			func() string {
				if store.IsActive {
					return "فعال"
				}
				return "در انتظار تایید"
			}(),
		)

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📊 مدیریت", fmt.Sprintf("manage_store_%d", store.ID)),
			),
		)

		msg := tgbotapi.NewMessage(chatID, storeInfo)
		msg.ReplyMarkup = keyboard
		mb.bot.Send(msg)
	}
}

func (mb *MotherBot) showStorePanel(chatID int64, user *models.User) {
	stores, err := mb.storeService.GetUserStores(user.ID)
	if err != nil || len(stores) == 0 {
		mb.sendMessage(chatID, messages.ErrorNoStore)
		return
	}

	// Show panel for first active store
	for _, store := range stores {
		if store.IsActive {
			mb.showStorePanelForStore(chatID, &store)
			return
		}
	}

	mb.sendMessage(chatID, "هیچ فروشگاه فعالی ندارید")
}

func (mb *MotherBot) showStorePanelForStore(chatID int64, store *models.Store) {
	products, _ := mb.productService.GetStoreProducts(store.ID)
	
	panelText := fmt.Sprintf(messages.StorePanelMessage,
		store.Name,
		strings.ToUpper(store.PlanType),
		store.ExpiresAt.Format("2006/01/02"),
		len(products),
		store.ProductLimit,
	)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(messages.AddProductBtn, fmt.Sprintf("add_product_%d", store.ID)),
			tgbotapi.NewInlineKeyboardButtonData(messages.ProductListBtn, fmt.Sprintf("product_list_%d", store.ID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(messages.OrdersBtn, fmt.Sprintf("orders_%d", store.ID)),
			tgbotapi.NewInlineKeyboardButtonData(messages.SalesReportBtn, fmt.Sprintf("sales_%d", store.ID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(messages.RenewPlanBtn, fmt.Sprintf("renew_%d", store.ID)),
			tgbotapi.NewInlineKeyboardButtonData(messages.StoreSettingsBtn, fmt.Sprintf("settings_%d", store.ID)),
		),
	)

	msg := tgbotapi.NewMessage(chatID, panelText)
	msg.ReplyMarkup = keyboard
	mb.bot.Send(msg)
}

func (mb *MotherBot) showSupportMenu(chatID int64) {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(messages.FAQBtn, "faq"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(messages.ContactUsBtn, "contact"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(messages.TelegramSupportBtn, "telegram_support"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, messages.SupportMessage)
	msg.ReplyMarkup = keyboard
	mb.bot.Send(msg)
}

func (mb *MotherBot) showAboutMessage(chatID int64) {
	aboutText := `ℹ️ درباره ربات CodeRoot

🤖 این ربات به شما کمک می‌کند تا فروشگاه تلگرامی خودتان را راه‌اندازی کنید

✨ امکانات:
• ساخت ربات فروشگاهی اختصاصی
• مدیریت محصولات و سفارش‌ها
• سیستم پرداخت کارت‌به‌کارت
• گزارش‌های فروش
• پشتیبانی ۲۴/۷

📞 پشتیبانی: @support
🌐 کانال: @channel`

	mb.sendMessage(chatID, aboutText)
}

func (mb *MotherBot) showAdminPanel(chatID int64) {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 آمار کلی", "admin_stats"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏪 مدیریت فروشگاه‌ها", "admin_stores"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💰 مدیریت پرداخت‌ها", "admin_payments"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📢 ارسال پیام همگانی", "admin_broadcast"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, "🔧 پنل مدیریت")
	msg.ReplyMarkup = keyboard
	mb.bot.Send(msg)
}

// Helper methods
func (mb *MotherBot) getPhotoURL(photos []tgbotapi.PhotoSize) string {
	if len(photos) == 0 {
		return ""
	}
	
	// Get the largest photo
	photo := photos[len(photos)-1]
	file, err := mb.bot.GetFile(tgbotapi.FileConfig{FileID: photo.FileID})
	if err != nil {
		return ""
	}
	
	return fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", mb.bot.Token, file.FilePath)
}

// Additional handler methods will be implemented in separate files
func (mb *MotherBot) handleCallbackQuery(callback *tgbotapi.CallbackQuery) {
	// This will be implemented in callback.go
}

func (mb *MotherBot) handlePaymentProof(chatID int64, user *models.User, photos []tgbotapi.PhotoSize, session *models.UserSession) {
	// This will be implemented in payment.go
}

func (mb *MotherBot) handleProductName(chatID int64, user *models.User, productName string, session *models.UserSession) {
	// This will be implemented in product.go
}

func (mb *MotherBot) handleProductDescription(chatID int64, user *models.User, description string, session *models.UserSession) {
	// This will be implemented in product.go
}

func (mb *MotherBot) handleProductPrice(chatID int64, user *models.User, priceStr string, session *models.UserSession) {
	// This will be implemented in product.go
}

func (mb *MotherBot) finalizeProduct(chatID int64, user *models.User, imageURL string, session *models.UserSession) {
	// This will be implemented in product.go
}