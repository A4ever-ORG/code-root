package bot

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"telegram-store-hub/internal/messages"
	"telegram-store-hub/internal/models"
	"telegram-store-hub/internal/services"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

// ComprehensiveMotherBot represents the main bot with all features
type ComprehensiveMotherBot struct {
	bot               *tgbotapi.BotAPI
	db                *gorm.DB
	channelVerify     *services.ChannelVerificationService
	userService       *services.UserService
	sessionService    *services.SessionService
	storeManager      *services.StoreManagerService
	productService    *services.ProductService
	orderService      *services.OrderService
	paymentService    *services.PaymentService
	subscriptionSrv   *services.SubscriptionService
	botManager        *services.BotManagerService
	adminChatID       int64
	paymentCardNumber string
	paymentCardHolder string
}

// StoreRegistrationData holds store registration session data
type StoreRegistrationData struct {
	StoreName   string `json:"store_name"`
	Description string `json:"description"`
	PlanType    string `json:"plan_type"`
	Step        int    `json:"step"`
}

// ProductData holds product creation session data
type ProductData struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int64  `json:"price"`
	ImageURL    string `json:"image_url"`
	Category    string `json:"category"`
	Step        int    `json:"step"`
	StoreID     uint   `json:"store_id"`
}

// NewComprehensiveMotherBot creates a new comprehensive mother bot
func NewComprehensiveMotherBot(
	bot *tgbotapi.BotAPI,
	db *gorm.DB,
	channelVerify *services.ChannelVerificationService,
	userService *services.UserService,
	sessionService *services.SessionService,
	storeManager *services.StoreManagerService,
	productService *services.ProductService,
	orderService *services.OrderService,
	paymentService *services.PaymentService,
	subscriptionSrv *services.SubscriptionService,
	botManager *services.BotManagerService,
	adminChatID int64,
	paymentCardNumber string,
	paymentCardHolder string,
) *ComprehensiveMotherBot {
	return &ComprehensiveMotherBot{
		bot:               bot,
		db:                db,
		channelVerify:     channelVerify,
		userService:       userService,
		sessionService:    sessionService,
		storeManager:      storeManager,
		productService:    productService,
		orderService:      orderService,
		paymentService:    paymentService,
		subscriptionSrv:   subscriptionSrv,
		botManager:        botManager,
		adminChatID:       adminChatID,
		paymentCardNumber: paymentCardNumber,
		paymentCardHolder: paymentCardHolder,
	}
}

// Start starts the bot and listens for updates
func (mb *ComprehensiveMotherBot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := mb.bot.GetUpdatesChan(u)
	
	log.Println("👂 Comprehensive Mother Bot is listening for messages...")

	for update := range updates {
		if update.Message != nil {
			go mb.handleMessage(update.Message)
		} else if update.CallbackQuery != nil {
			go mb.handleCallback(update.CallbackQuery)
		}
	}
}

// handleMessage processes incoming messages
func (mb *ComprehensiveMotherBot) handleMessage(message *tgbotapi.Message) {
	chatID := message.Chat.ID
	messageText := message.Text

	// Create or update user
	user := &models.User{
		TelegramID: chatID,
		Username:   message.From.UserName,
		FirstName:  message.From.FirstName,
		LastName:   message.From.LastName,
	}

	if err := mb.userService.CreateOrUpdateUser(user); err != nil {
		log.Printf("Error creating/updating user: %v", err)
	}

	// Check channel membership first
	isJoined, err := mb.channelVerify.CheckAndHandleMembership(chatID)
	if err != nil {
		log.Printf("Error checking membership: %v", err)
	}
	if !isJoined {
		return // User needs to join channel first
	}

	// Handle user session state
	userState, err := mb.sessionService.GetUserState(chatID)
	if err == nil && userState != nil {
		mb.handleSessionState(message, userState)
		return
	}

	// Handle commands and regular messages
	switch {
	case strings.HasPrefix(messageText, "/start"):
		mb.sendWelcome(chatID)
	case strings.HasPrefix(messageText, "/help"):
		mb.sendHelp(chatID)
	case strings.HasPrefix(messageText, "/admin") && mb.isAdmin(chatID):
		mb.showAdminPanel(chatID)
	case strings.HasPrefix(messageText, "/stats") && mb.isAdmin(chatID):
		mb.showSystemStats(chatID)
	default:
		// If user has an active session, handle it
		if userState != nil {
			mb.handleSessionState(message, userState)
		} else {
			mb.sendMainMenu(chatID)
		}
	}
}

// handleCallback processes callback queries
func (mb *ComprehensiveMotherBot) handleCallback(callback *tgbotapi.CallbackQuery) {
	chatID := callback.Message.Chat.ID
	data := callback.Data

	// Answer callback query
	mb.bot.Request(tgbotapi.NewCallback(callback.ID, ""))

	// Check channel membership for callbacks too
	isJoined, err := mb.channelVerify.CheckAndHandleMembership(chatID)
	if err != nil {
		log.Printf("Error checking membership: %v", err)
	}
	if !isJoined {
		return
	}

	switch {
	case data == "register_store":
		mb.showRegistrationMenu(chatID)
	case strings.HasPrefix(data, "plan_"):
		planType := strings.TrimPrefix(data, "plan_")
		mb.handlePlanSelection(chatID, planType)
	case data == "manage_store":
		mb.showStoreManagement(chatID)
	case data == "view_plans":
		mb.showRegistrationMenu(chatID)
	case data == "support":
		mb.sendSupport(chatID)
	case data == "back_main":
		mb.sendWelcome(chatID)
	case data == "check_membership":
		// Re-check membership when user clicks the button
		mb.sendWelcome(chatID)
		
	// Store management callbacks
	case data == "add_product":
		mb.startProductAddition(chatID)
	case data == "list_products":
		mb.showProductList(chatID)
	case data == "view_orders":
		mb.showOrderList(chatID)
	case data == "sales_report":
		mb.showSalesReport(chatID)
	case data == "store_settings":
		mb.showStoreSettings(chatID)
	case data == "renew_plan":
		mb.showPlanRenewal(chatID)
		
	// Product management callbacks
	case strings.HasPrefix(data, "product_"):
		mb.handleProductCallback(callback, data)
		
	// Order management callbacks
	case strings.HasPrefix(data, "order_"):
		mb.handleOrderCallback(callback, data)
		
	// Payment callbacks
	case strings.HasPrefix(data, "paid_"):
		mb.handlePaymentCallback(callback, data)
		
	// Admin panel callbacks
	case strings.HasPrefix(data, "admin_") && mb.isAdmin(chatID):
		mb.handleAdminCallback(callback, data)
	}
}

// sendWelcome sends welcome message with main menu
func (mb *ComprehensiveMotherBot) sendWelcome(chatID int64) {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(messages.ButtonRegisterStore, "register_store"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(messages.ButtonManageStore, "manage_store"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(messages.ButtonViewPlans, "view_plans"),
			tgbotapi.NewInlineKeyboardButtonData(messages.ButtonSupport, "support"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, messages.WelcomeMessage)
	msg.ReplyMarkup = keyboard

	mb.bot.Send(msg)
}

// sendMainMenu sends main menu
func (mb *ComprehensiveMotherBot) sendMainMenu(chatID int64) {
	mb.sendWelcome(chatID)
}

// showRegistrationMenu shows store registration menu with plans
func (mb *ComprehensiveMotherBot) showRegistrationMenu(chatID int64) {
	text := fmt.Sprintf(`%s

%s

%s

%s`,
		messages.StoreRegistrationStart,
		messages.FreePlanDescription,
		messages.ProPlanDescription,
		messages.VIPPlanDescription)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(messages.ButtonFreePlan, "plan_free"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(messages.ButtonProPlan, "plan_pro"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(messages.ButtonVIPPlan, "plan_vip"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(messages.ButtonBack, "back_main"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard

	mb.bot.Send(msg)
}

// handlePlanSelection handles plan selection
func (mb *ComprehensiveMotherBot) handlePlanSelection(chatID int64, planType string) {
	if planType == "free" {
		mb.startStoreRegistration(chatID, planType)
		return
	}

	// For paid plans, show payment instructions
	var price string
	switch planType {
	case "pro":
		price = "50,000"
	case "vip":
		price = "150,000"
	default:
		price = "0"
	}

	paymentText := fmt.Sprintf(`💳 پرداخت پلن %s

مبلغ: %s تومان

%s

%s`,
		planType,
		price,
		fmt.Sprintf(messages.PaymentCardInfo, mb.paymentCardNumber, mb.paymentCardHolder, price),
		messages.PaymentInstructions)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(messages.ButtonPaymentComplete, fmt.Sprintf("paid_%s", planType)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(messages.ButtonBack, "view_plans"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, paymentText)
	msg.ReplyMarkup = keyboard

	mb.bot.Send(msg)
}

// startStoreRegistration starts the store registration process
func (mb *ComprehensiveMotherBot) startStoreRegistration(chatID int64, planType string) {
	// Check if user already has a store
	user, err := mb.userService.GetUserByTelegramID(chatID)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, messages.ErrorDatabaseError)
		mb.bot.Send(msg)
		return
	}

	stores, err := mb.userService.GetUserStores(chatID)
	if err == nil && len(stores) > 0 {
		msg := tgbotapi.NewMessage(chatID, messages.ErrorStoreExists)
		mb.bot.Send(msg)
		return
	}

	// Start registration process
	registrationData := StoreRegistrationData{
		PlanType: planType,
		Step:     1,
	}

	mb.sessionService.SetUserState(chatID, messages.StateWaitingStoreName, registrationData)

	msg := tgbotapi.NewMessage(chatID, "📝 لطفاً نام فروشگاه خود را وارد کنید:")
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboardRemove{RemoveKeyboard: true}
	
	mb.bot.Send(msg)
}

// handleSessionState handles user session states
func (mb *ComprehensiveMotherBot) handleSessionState(message *tgbotapi.Message, userState *models.UserSession) {
	chatID := message.Chat.ID
	messageText := message.Text

	switch userState.State {
	case messages.StateWaitingStoreName:
		mb.handleStoreNameInput(chatID, messageText, userState)
	case messages.StateWaitingStoreDescription:
		mb.handleStoreDescriptionInput(chatID, messageText, userState)
	case messages.StateWaitingPaymentProof:
		mb.handlePaymentProofInput(message, userState)
	case messages.StateWaitingProductName:
		mb.handleProductNameInput(chatID, messageText, userState)
	case messages.StateWaitingProductPrice:
		mb.handleProductPriceInput(chatID, messageText, userState)
	case messages.StateWaitingProductImage:
		mb.handleProductImageInput(message, userState)
	default:
		// Unknown state, clear it
		mb.sessionService.ClearUserState(chatID)
		mb.sendMainMenu(chatID)
	}
}

// handleStoreNameInput handles store name input
func (mb *ComprehensiveMotherBot) handleStoreNameInput(chatID int64, storeName string, userState *models.UserSession) {
	if len(storeName) < 3 || len(storeName) > 50 {
		msg := tgbotapi.NewMessage(chatID, "❌ نام فروشگاه باید بین 3 تا 50 کاراکتر باشد. لطفاً دوباره وارد کنید:")
		mb.bot.Send(msg)
		return
	}

	// Update registration data
	var registrationData StoreRegistrationData
	if err := json.Unmarshal([]byte(userState.Data), &registrationData); err != nil {
		log.Printf("Error unmarshaling registration data: %v", err)
		mb.sessionService.ClearUserState(chatID)
		mb.sendMainMenu(chatID)
		return
	}

	registrationData.StoreName = storeName
	registrationData.Step = 2

	mb.sessionService.SetUserState(chatID, messages.StateWaitingStoreDescription, registrationData)

	msg := tgbotapi.NewMessage(chatID, "📝 حالا توضیح کوتاهی از فروشگاه خود بنویسید:")
	mb.bot.Send(msg)
}

// handleStoreDescriptionInput handles store description input
func (mb *ComprehensiveMotherBot) handleStoreDescriptionInput(chatID int64, description string, userState *models.UserSession) {
	if len(description) > 500 {
		msg := tgbotapi.NewMessage(chatID, "❌ توضیحات نباید بیشتر از 500 کاراکتر باشد. لطفاً دوباره وارد کنید:")
		mb.bot.Send(msg)
		return
	}

	// Update registration data
	var registrationData StoreRegistrationData
	if err := json.Unmarshal([]byte(userState.Data), &registrationData); err != nil {
		log.Printf("Error unmarshaling registration data: %v", err)
		mb.sessionService.ClearUserState(chatID)
		mb.sendMainMenu(chatID)
		return
	}

	registrationData.Description = description

	// Create the store
	mb.createStore(chatID, registrationData)
}

// createStore creates a new store
func (mb *ComprehensiveMotherBot) createStore(chatID int64, registrationData StoreRegistrationData) {
	// Get user
	user, err := mb.userService.GetUserByTelegramID(chatID)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, messages.ErrorDatabaseError)
		mb.bot.Send(msg)
		return
	}

	// Determine plan limits
	var productLimit int
	var commissionRate int
	var expiresAt time.Time

	switch registrationData.PlanType {
	case "free":
		productLimit = 10
		commissionRate = 5
		expiresAt = time.Now().AddDate(0, 1, 0) // 1 month
	case "pro":
		productLimit = 200
		commissionRate = 5
		expiresAt = time.Now().AddDate(0, 1, 0) // 1 month
	case "vip":
		productLimit = -1 // unlimited
		commissionRate = 0
		expiresAt = time.Now().AddDate(0, 1, 0) // 1 month
	}

	// Create store
	store := &models.Store{
		OwnerID:        user.ID,
		Name:           registrationData.StoreName,
		Description:    registrationData.Description,
		PlanType:       models.PlanType(registrationData.PlanType),
		ExpiresAt:      expiresAt,
		IsActive:       true,
		ProductLimit:   productLimit,
		CommissionRate: commissionRate,
	}

	if err := mb.storeManager.CreateStore(store); err != nil {
		log.Printf("Error creating store: %v", err)
		msg := tgbotapi.NewMessage(chatID, messages.ErrorCreateStore)
		mb.bot.Send(msg)
		return
	}

	// Clear session
	mb.sessionService.ClearUserState(chatID)

	// Send success message
	successText := fmt.Sprintf(`🎉 تبریک! فروشگاه "%s" با موفقیت ثبت شد!

📋 اطلاعات فروشگاه:
• نوع پلن: %s
• تعداد محصولات مجاز: %s
• کارمزد: %d%%
• مدت اعتبار: 1 ماه

🤖 ربات فروشگاهی شما تا 24 ساعت آینده آماده خواهد شد.

برای مدیریت فروشگاه از دکمه "پنل مدیریت" استفاده کنید.`,
		store.Name,
		string(store.PlanType),
		func() string {
			if productLimit == -1 {
				return "نامحدود"
			}
			return fmt.Sprintf("%d", productLimit)
		}(),
		commissionRate)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(messages.ButtonManageStore, "manage_store"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(messages.ButtonMainMenu, "back_main"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, successText)
	msg.ReplyMarkup = keyboard

	mb.bot.Send(msg)

	// Trigger sub-bot creation (asynchronous)
	go mb.createSubBot(store)
}

// createSubBot creates a sub-bot for the store (placeholder for now)
func (mb *ComprehensiveMotherBot) createSubBot(store *models.Store) {
	log.Printf("Creating sub-bot for store: %s (ID: %d)", store.Name, store.ID)
	// TODO: Implement sub-bot creation logic
	// This would involve:
	// 1. Creating a new bot token via BotFather API (if available)
	// 2. Setting up the sub-bot with store-specific configuration
	// 3. Updating store record with bot token and username
	// 4. Starting the sub-bot instance
}

// isAdmin checks if user is admin
func (mb *ComprehensiveMotherBot) isAdmin(chatID int64) bool {
	return chatID == mb.adminChatID
}

// sendHelp sends help message
func (mb *ComprehensiveMotherBot) sendHelp(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, messages.SupportMessage)
	mb.bot.Send(msg)
}

// sendSupport sends support message
func (mb *ComprehensiveMotherBot) sendSupport(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, messages.SupportMessage)
	mb.bot.Send(msg)
}