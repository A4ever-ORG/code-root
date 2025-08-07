package bot

import (
        "fmt"
        "log"
        "strings"
        "telegram-store-hub/internal/messages"
        "telegram-store-hub/internal/models"
        "telegram-store-hub/internal/services"
        "time"

        tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
        "gorm.io/gorm"
)

type MotherBot struct {
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
}

func NewMotherBot(
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
) *MotherBot {
        return &MotherBot{
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
        }
}

func (mb *MotherBot) Start() {
        u := tgbotapi.NewUpdate(0)
        u.Timeout = 60

        updates := mb.bot.GetUpdatesChan(u)
        
        log.Println("👂 Mother Bot is listening for messages...")

        for update := range updates {
                if update.Message != nil {
                        go mb.handleMessage(update.Message) // Handle in goroutine for better performance
                } else if update.CallbackQuery != nil {
                        go mb.handleCallback(update.CallbackQuery)
                }
        }
}

func (mb *MotherBot) handleMessage(message *tgbotapi.Message) {
        chatID := message.Chat.ID
        text := message.Text
        
        // Get or create user
        username := ""
        if message.From.UserName != "" {
                username = message.From.UserName
        }
        
        user, err := mb.userService.GetOrCreateUser(chatID, username, message.From.FirstName, message.From.LastName)
        if err != nil {
                log.Printf("Error getting user: %v", err)
                return
        }

        // Check channel membership first
        isJoined, err := mb.channelVerify.CheckAndHandleMembership(chatID)
        if err != nil {
                log.Printf("Error checking membership: %v", err)
        }
        if !isJoined {
                return // User will receive join message
        }
        
        // Check if user is in conversation state
        session, err := mb.sessionService.GetUserState(chatID)
        if err == nil && session.State != "" {
                mb.handleConversationState(message, session)
                return
        }

        switch {
        case text == "/start":
                mb.sendWelcome(chatID)
        case text == "/register" || text == "🏪 ثبت فروشگاه جدید":
                mb.startStoreRegistration(chatID)
        case text == "/panel" || text == "📊 پنل مدیریت":
                mb.showUserPanel(chatID)
        case text == "/admin" && user.IsAdmin:
                mb.showAdminPanel(chatID)
        case text == "🆘 پشتیبانی":
                mb.showSupportMenu(chatID)
        case strings.HasPrefix(text, "/store"):
                mb.handleStoreCommand(chatID, text)
        default:
                // Handle conversation states or show main menu
                mb.sendMainMenu(chatID)
        }
}

func (mb *MotherBot) sendWelcome(chatID int64) {
        msg := tgbotapi.NewMessage(chatID, messages.WelcomeMessage)
        
        keyboard := tgbotapi.NewInlineKeyboardMarkup(
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData(messages.RegisterStoreBtn, "register_store"),
                ),
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData(messages.MyStoresBtn, "my_stores"),
                        tgbotapi.NewInlineKeyboardButtonData("💎 پلن‌ها", "view_plans"),
                ),
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData(messages.SupportBtn, "support"),
                        tgbotapi.NewInlineKeyboardButtonData(messages.AboutBtn, "about"),
                ),
        )
        
        msg.ReplyMarkup = keyboard
        msg.ParseMode = "HTML"
        mb.bot.Send(msg)
}

func (mb *MotherBot) sendMainMenu(chatID int64) {
        mb.sendWelcome(chatID)
}

// Remove duplicate methods that are already implemented in comprehensive_bot.go
func (mb *MotherBot) handleStoreCommand(chatID int64, command string) {
        // Handle store-specific commands
        mb.sendMainMenu(chatID)
}

func (mb *MotherBot) showRegistrationMenu(chatID int64) {
        text := `🏪 ثبت فروشگاه جدید

لطفاً پلن مورد نظر خود را انتخاب کنید:

💚 پلن رایگان - 0 تومان
• حداکثر 10 محصول
• دکمه‌های ثابت
• کارمزد 5٪

💎 پلن حرفه‌ای - 50,000 تومان/ماه  
• تا 200 محصول
• گزارش‌های پیشرفته
• پیام خوش‌آمدگویی
• تبلیغات دلخواه
• کارمزد 5٪

👑 پلن VIP - 150,000 تومان/ماه
• محصولات نامحدود
• درگاه پرداخت اختصاصی
• بدون کارمزد
• تبلیغات ویژه
• شخصی‌سازی کامل`

        keyboard := tgbotapi.NewInlineKeyboardMarkup(
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("💚 پلن رایگان", "plan_free"),
                ),
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("💎 پلن حرفه‌ای", "plan_pro"),
                ),
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("👑 پلن VIP", "plan_vip"),
                ),
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("🔙 بازگشت", "back_main"),
                ),
        )

        msg := tgbotapi.NewMessage(chatID, text)
        msg.ReplyMarkup = keyboard

        mb.bot.Send(msg)
}

func (mb *MotherBot) handleCallback(callback *tgbotapi.CallbackQuery) {
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
                mb.handlePlanSelection(chatID, models.PlanType(planType))
        case data == "manage_store":
                mb.showStoreManagement(chatID)
        case data == "view_plans":
                mb.showRegistrationMenu(chatID)
        case data == "back_main":
                mb.sendWelcome(chatID)
        case data == "check_membership":
                // Re-check membership when user clicks the button
                mb.sendWelcome(chatID)
        // Seller panel callbacks
        case data == "add_product" || data == "list_products" || data == "view_orders" || 
                 data == "sales_report" || data == "store_settings" || data == "renew_plan" ||
                 strings.Contains(data, "product_") || strings.Contains(data, "order_") || strings.Contains(data, "upgrade_"):
                mb.handleSellerPanel(callback)
        // Admin panel callbacks  
        case strings.HasPrefix(data, "admin_") || strings.Contains(data, "payment_") || 
                 strings.Contains(data, "store_") && mb.isAdmin(chatID):
                mb.handleAdminPanel(callback)
        }
}

func (mb *MotherBot) handlePlanSelection(chatID int64, planType models.PlanType) {
        if planType == models.PlanFree {
                mb.createFreeStore(chatID)
                return
        }

        // For paid plans, show payment instructions
        var price string
        switch planType {
        case models.PlanPro:
                price = "50,000"
        case models.PlanVIP:
                price = "150,000"
        }

        paymentText := fmt.Sprintf(`💳 پرداخت پلن %s

مبلغ: %s تومان

🏦 اطلاعات پرداخت:
شماره کارت: 1234-5678-9012-3456
به نام: فروشگاه CodeRoot

پس از پرداخت، اسکرین‌شات رسید را ارسال کنید.

⚠️ توجه: پس از تایید پرداخت، ربات شما در کمتر از 10 دقیقه آماده خواهد شد.`, planType, price)

        keyboard := tgbotapi.NewInlineKeyboardMarkup(
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("✅ پرداخت کردم", fmt.Sprintf("paid_%s", planType)),
                ),
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("🔙 بازگشت", "back_main"),
                ),
        )

        msg := tgbotapi.NewMessage(chatID, paymentText)
        msg.ReplyMarkup = keyboard

        mb.bot.Send(msg)
}

func (mb *MotherBot) createFreeStore(chatID int64) {
        // Check if user already has a store
        var existingStore models.Store
        // First get the user ID from telegram ID
        var user models.User
        userResult := mb.db.Where("telegram_id = ?", chatID).First(&user)
        if userResult.Error != nil {
                msg := tgbotapi.NewMessage(chatID, "❌ لطفاً ابتدا با /start شروع کنید")
                mb.bot.Send(msg)
                return
        }
        
        result := mb.db.Where("owner_id = ?", user.ID).First(&existingStore)
        if result.Error == nil {
                msg := tgbotapi.NewMessage(chatID, "⚠️ شما قبلاً یک فروشگاه ثبت کرده‌اید!")
                mb.bot.Send(msg)
                return
        }

        // Create new store with proper user reference
        store := models.Store{
                OwnerID: user.ID,
                Name:    fmt.Sprintf("Store_%d", chatID),
                PlanType:    models.PlanFree,
                ExpiresAt:   time.Now().AddDate(0, 1, 0), // 1 month for free
                IsActive:    true,
        }

        if err := mb.db.Create(&store).Error; err != nil {
                msg := tgbotapi.NewMessage(chatID, "❌ خطا در ثبت فروشگاه. لطفاً دوباره تلاش کنید.")
                mb.bot.Send(msg)
                return
        }

        successText := `🎉 تبریک! فروشگاه شما با موفقیت ثبت شد!

📋 اطلاعات فروشگاه:
• نوع پلن: رایگان
• تعداد محصولات مجاز: 10
• مدت اعتبار: 1 ماه

🤖 ربات فروشگاهی شما تا 24 ساعت آینده آماده خواهد شد.

برای مدیریت فروشگاه از دکمه "پنل مدیریت" استفاده کنید.`

        keyboard := tgbotapi.NewInlineKeyboardMarkup(
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("📊 پنل مدیریت", "manage_store"),
                ),
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("🏠 منوی اصلی", "back_main"),
                ),
        )

        msg := tgbotapi.NewMessage(chatID, successText)
        msg.ReplyMarkup = keyboard

        mb.bot.Send(msg)
}

func (mb *MotherBot) showStoreManagement(chatID int64) {
        var store models.Store
        // Get user first, then find their stores
        var user models.User
        userResult := mb.db.Where("telegram_id = ?", chatID).First(&user)
        if userResult.Error != nil {
                msg := tgbotapi.NewMessage(chatID, "❌ لطفاً ابتدا با /start شروع کنید")
                mb.bot.Send(msg)
                return
        }
        
        result := mb.db.Where("owner_id = ? AND is_active = ?", user.ID, true).First(&store)
        if result.Error != nil {
                msg := tgbotapi.NewMessage(chatID, "❌ شما هنوز فروشگاهی ثبت نکرده‌اید!")
                mb.bot.Send(msg)
                return
        }

        // Get products count
        var productCount int64
        mb.db.Model(&models.Product{}).Where("store_id = ?", store.ID).Count(&productCount)

        // Get orders count
        var orderCount int64
        mb.db.Model(&models.Order{}).Where("store_id = ?", store.ID).Count(&orderCount)

        managementText := fmt.Sprintf(`📊 پنل مدیریت فروشگاه

🏪 نام فروشگاه: %s
📦 تعداد محصولات: %d
🛒 تعداد سفارش‌ها: %d
💎 پلن فعلی: %s
⏰ انقضا: %s

عملیات مورد نظر را انتخاب کنید:`, 
                store.Name, 
                productCount, 
                orderCount, 
                store.PlanType,
                store.ExpiresAt.Format("2006/01/02"),
        )

        keyboard := tgbotapi.NewInlineKeyboardMarkup(
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("➕ افزودن محصول", "add_product"),
                        tgbotapi.NewInlineKeyboardButtonData("📦 لیست محصولات", "list_products"),
                ),
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("🛒 سفارش‌ها", "view_orders"),
                        tgbotapi.NewInlineKeyboardButtonData("📈 گزارش فروش", "sales_report"),
                ),
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("⚙️ تنظیمات", "store_settings"),
                        tgbotapi.NewInlineKeyboardButtonData("🔄 تمدید پلن", "renew_plan"),
                ),
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("🔙 بازگشت", "back_main"),
                ),
        )

        msg := tgbotapi.NewMessage(chatID, managementText)
        msg.ReplyMarkup = keyboard

        mb.bot.Send(msg)
}

func (mb *MotherBot) sendMainMenu(chatID int64) {
        mb.sendWelcome(chatID)
}

func (mb *MotherBot) showAdminPanel(chatID int64) {
        // Get statistics
        var storeCount, activeStoreCount, totalOrders int64
        mb.db.Model(&models.Store{}).Count(&storeCount)
        mb.db.Model(&models.Store{}).Where("is_active = ?", true).Count(&activeStoreCount)
        mb.db.Model(&models.Order{}).Count(&totalOrders)

        adminText := fmt.Sprintf(`🔧 پنل مدیریت سیستم

📊 آمار کلی:
• تعداد کل فروشگاه‌ها: %d
• فروشگاه‌های فعال: %d  
• کل سفارش‌ها: %d

عملیات مورد نظر را انتخاب کنید:`, storeCount, activeStoreCount, totalOrders)

        keyboard := tgbotapi.NewInlineKeyboardMarkup(
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("🏪 مدیریت فروشگاه‌ها", "admin_stores"),
                        tgbotapi.NewInlineKeyboardButtonData("💰 پرداخت‌ها", "admin_payments"),
                ),
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("📊 گزارش مالی", "admin_financial"),
                        tgbotapi.NewInlineKeyboardButtonData("📢 ارسال پیام", "admin_broadcast"),
                ),
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("🔙 بازگشت", "back_main"),
                ),
        )

        msg := tgbotapi.NewMessage(chatID, adminText)
        msg.ReplyMarkup = keyboard

        mb.bot.Send(msg)
}

func (mb *MotherBot) isAdmin(chatID int64) bool {
        var user models.User
        result := mb.db.Where("chat_id = ? AND is_admin = ?", chatID, true).First(&user)
        return result.Error == nil
}

// Stub methods for missing handlers - in production these would be full implementations
func (mb *MotherBot) handleStoreRegistration(callback *tgbotapi.CallbackQuery) {
        // Simplified store registration handler
        mb.bot.Send(tgbotapi.NewMessage(callback.Message.Chat.ID, "🏪 Store registration functionality coming soon!"))
}

func (mb *MotherBot) handleAdminPanel(callback *tgbotapi.CallbackQuery) {
        // Simplified admin panel handler
        mb.bot.Send(tgbotapi.NewMessage(callback.Message.Chat.ID, "👨‍💼 Admin panel functionality coming soon!"))
}

func (mb *MotherBot) handleSellerPanel(callback *tgbotapi.CallbackQuery) {
        // Simplified seller panel handler
        mb.bot.Send(tgbotapi.NewMessage(callback.Message.Chat.ID, "📊 Seller panel functionality coming soon!"))
}

func (mb *MotherBot) handleStoreCommand(chatID int64, command string) {
        // Handle store-specific commands
        parts := strings.Split(command, " ")
        if len(parts) < 2 {
                return
        }

        switch parts[1] {
        case "stats":
                mb.showStoreStats(chatID)
        }
}

func (mb *MotherBot) showStoreStats(chatID int64) {
        var store models.Store
        result := mb.db.Where("owner_chat_id = ?", chatID).First(&store)
        if result.Error != nil {
                return
        }

        var productCount, orderCount int64
        var totalRevenue float64

        mb.db.Model(&models.Product{}).Where("store_id = ?", store.ID).Count(&productCount)
        mb.db.Model(&models.Order{}).Where("store_id = ?", store.ID).Count(&orderCount)
        mb.db.Model(&models.Order{}).Where("store_id = ?", store.ID).Select("COALESCE(SUM(total_amount), 0)").Scan(&totalRevenue)

        statsText := fmt.Sprintf(`📈 آمار فروشگاه %s

📦 تعداد محصولات: %d
🛒 تعداد سفارش‌ها: %d
💰 کل فروش: %.0f تومان
💎 پلن: %s`, store.StoreName, productCount, orderCount, totalRevenue, store.PlanType)

        msg := tgbotapi.NewMessage(chatID, statsText)
        mb.bot.Send(msg)
}