package bot

import (
        "fmt"
        "log"
        "strconv"
        "strings"
        "telegram-store-hub/internal/messages"
        "telegram-store-hub/internal/models"
        "time"

        tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Store registration data structure for sessions
type StoreRegistrationData struct {
        StoreName   string `json:"store_name"`
        Description string `json:"description"`
        PlanType    string `json:"plan_type"`
        Step        int    `json:"step"`
}

// Product addition data structure for sessions
type ProductData struct {
        Name        string `json:"name"`
        Description string `json:"description"`
        Price       int64  `json:"price"`
        Step        int    `json:"step"`
        StoreID     uint   `json:"store_id"`
}

// startStoreRegistration begins store registration process
func (mb *MotherBot) startStoreRegistration(chatID int64) {
        // Clear any existing session
        mb.sessionService.ClearUserState(chatID)
        
        // Set initial session state
        data := StoreRegistrationData{Step: 1}
        mb.sessionService.SetUserState(chatID, "registering_store", data)
        
        msg := tgbotapi.NewMessage(chatID, messages.StoreRegistrationStart)
        msg.ReplyMarkup = tgbotapi.NewReplyKeyboardRemove{RemoveKeyboard: true}
        mb.bot.Send(msg)
}

// showUserPanel shows user management panel
func (mb *MotherBot) showUserPanel(chatID int64) {
        stores, err := mb.userService.GetUserStores(chatID)
        if err != nil || len(stores) == 0 {
                msg := tgbotapi.NewMessage(chatID, messages.ErrorNoStore)
                mb.bot.Send(msg)
                return
        }

        // Show stores list
        text := "📊 فروشگاه‌های شما:\n\n"
        for i, store := range stores {
                status := "✅ فعال"
                if !store.IsActive {
                        status = "❌ غیرفعال"
                }
                
                daysLeft := int(time.Until(store.ExpiresAt).Hours() / 24)
                if daysLeft < 0 {
                        status = "⏰ منقضی شده"
                }
                
                text += fmt.Sprintf("🏪 %s\n", store.Name)
                text += fmt.Sprintf("📋 پلن: %s\n", string(store.PlanType))
                text += fmt.Sprintf("📊 وضعیت: %s\n", status)
                text += fmt.Sprintf("🗓 روزهای باقی‌مانده: %d\n", daysLeft)
                text += fmt.Sprintf("📦 محصولات: %d/%d\n", len(store.Products), store.ProductLimit)
                
                if i < len(stores)-1 {
                        text += "\n────────────\n\n"
                }
        }
        
        keyboard := tgbotapi.NewInlineKeyboardMarkup()
        for _, store := range stores {
                row := tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData(
                                fmt.Sprintf("🏪 پنل %s", store.Name), 
                                fmt.Sprintf("manage_store_%d", store.ID)),
                )
                keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
        }
        
        // Add main menu button
        keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, 
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("🔙 منوی اصلی", "main_menu"),
                ))

        msg := tgbotapi.NewMessage(chatID, text)
        msg.ReplyMarkup = keyboard
        mb.bot.Send(msg)
}

// showSupportMenu shows support options
func (mb *MotherBot) showSupportMenu(chatID int64) {
        msg := tgbotapi.NewMessage(chatID, messages.SupportMessage)
        
        keyboard := tgbotapi.NewInlineKeyboardMarkup(
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData(messages.FAQBtn, "faq"),
                ),
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData(messages.ContactUsBtn, "contact_us"),
                ),
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData(messages.TelegramSupportBtn, "telegram_support"),
                ),
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("🔙 بازگشت", "main_menu"),
                ),
        )
        
        msg.ReplyMarkup = keyboard
        mb.bot.Send(msg)
}

// showAdminPanel shows admin control panel
func (mb *MotherBot) showAdminPanel(chatID int64) {
        text := `👨‍💼 پنل مدیریت سیستم

📊 خلاصه سیستم:`
        
        // Get system stats
        var userCount, storeCount, activeStoreCount, pendingPaymentCount int64
        mb.db.Model(&models.User{}).Count(&userCount)
        mb.db.Model(&models.Store{}).Count(&storeCount)
        mb.db.Model(&models.Store{}).Where("is_active = ?", true).Count(&activeStoreCount)
        mb.db.Model(&models.Payment{}).Where("status = ?", "pending").Count(&pendingPaymentCount)
        
        text += fmt.Sprintf("\n👥 تعداد کاربران: %d", userCount)
        text += fmt.Sprintf("\n🏪 تعداد فروشگاه‌ها: %d", storeCount)
        text += fmt.Sprintf("\n✅ فروشگاه‌های فعال: %d", activeStoreCount)
        text += fmt.Sprintf("\n⏳ پرداخت‌های در انتظار: %d", pendingPaymentCount)
        
        keyboard := tgbotapi.NewInlineKeyboardMarkup(
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("💳 بررسی پرداخت‌ها", "admin_payments"),
                        tgbotapi.NewInlineKeyboardButtonData("🏪 مدیریت فروشگاه‌ها", "admin_stores"),
                ),
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("👥 مدیریت کاربران", "admin_users"),
                        tgbotapi.NewInlineKeyboardButtonData("📊 گزارشات", "admin_reports"),
                ),
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("📢 پیام همگانی", "admin_broadcast"),
                        tgbotapi.NewInlineKeyboardButtonData("⚙️ تنظیمات", "admin_settings"),
                ),
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("🔙 منوی اصلی", "main_menu"),
                ),
        )
        
        msg := tgbotapi.NewMessage(chatID, text)
        msg.ReplyMarkup = keyboard
        mb.bot.Send(msg)
}

// handleConversationState processes user conversation states
func (mb *MotherBot) handleConversationState(message *tgbotapi.Message, session *models.UserSession) {
        chatID := message.Chat.ID
        
        switch session.State {
        case "registering_store":
                mb.handleStoreRegistrationFlow(message, session)
        case "adding_product":
                mb.handleProductAdditionFlow(message, session)
        case "editing_store":
                mb.handleStoreEditingFlow(message, session)
        case "contacting_support":
                mb.handleSupportContactFlow(message, session)
        case "admin_broadcast":
                mb.handleBroadcastFlow(message, session)
        default:
                // Unknown state, clear session
                mb.sessionService.ClearUserState(chatID)
                mb.sendMainMenu(chatID)
        }
}

func (mb *MotherBot) handleStoreRegistrationFlowOld(message *tgbotapi.Message, session *models.UserSession) {
        // This method is now in the main handleConversationState
}

// handleStoreRegistrationFlow handles store registration conversation
func (mb *MotherBot) handleStoreRegistrationFlow(message *tgbotapi.Message, session *models.UserSession) {
        chatID := message.Chat.ID
        text := message.Text
        
        var data StoreRegistrationData
        if err := mb.sessionService.GetSessionData(chatID, &data); err != nil {
                log.Printf("Error getting session data: %v", err)
                mb.sessionService.ClearUserState(chatID)
                mb.sendMainMenu(chatID)
                return
        }
        
        switch data.Step {
        case 1: // Store name
                if len(text) < 2 || len(text) > 50 {
                        msg := tgbotapi.NewMessage(chatID, "❌ نام فروشگاه باید بین 2 تا 50 کاراکتر باشد")
                        mb.bot.Send(msg)
                        return
                }
                
                data.StoreName = text
                data.Step = 2
                mb.sessionService.SetUserState(chatID, "registering_store", data)
                
                responseText := fmt.Sprintf(messages.StoreNameReceived, text)
                msg := tgbotapi.NewMessage(chatID, responseText)
                mb.bot.Send(msg)
                
        case 2: // Store description
                if len(text) < 10 || len(text) > 500 {
                        msg := tgbotapi.NewMessage(chatID, "❌ توضیحات فروشگاه باید بین 10 تا 500 کاراکتر باشد")
                        mb.bot.Send(msg)
                        return
                }
                
                data.Description = text
                data.Step = 3
                mb.sessionService.SetUserState(chatID, "registering_store", data)
                
                // Show plan selection
                mb.showPlanSelection(chatID)
        }
}

// showPlanSelection shows available subscription plans
func (mb *MotherBot) showPlanSelection(chatID int64) {
        text := messages.StoreDescriptionReceived + "\n\n"
        text += "📋 پلن‌های موجود:\n\n"
        text += messages.FreePlanDetails + "\n\n"
        text += messages.ProPlanDetails + "\n\n"
        text += messages.VIPPlanDetails
        
        keyboard := tgbotapi.NewInlineKeyboardMarkup(
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData(messages.FreePlanBtn, "select_plan_free"),
                ),
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData(messages.ProPlanBtn, "select_plan_pro"),
                ),
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData(messages.VIPPlanBtn, "select_plan_vip"),
                ),
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("🔙 لغو", "cancel_registration"),
                ),
        )
        
        msg := tgbotapi.NewMessage(chatID, text)
        msg.ReplyMarkup = keyboard
        mb.bot.Send(msg)
}

// handleCallback handles all callback queries
func (mb *MotherBot) handleCallback(callback *tgbotapi.CallbackQuery) {
        chatID := callback.Message.Chat.ID
        data := callback.Data
        userID := callback.From.ID
        
        // Answer callback query
        mb.bot.Request(tgbotapi.NewCallback(callback.ID, ""))
        
        // Get user
        user, err := mb.userService.GetOrCreateUser(userID, callback.From.UserName, callback.From.FirstName, callback.From.LastName)
        if err != nil {
                log.Printf("Error getting user: %v", err)
                return
        }
        
        // Check channel membership
        isJoined, err := mb.channelVerify.CheckAndHandleMembership(chatID)
        if err != nil {
                log.Printf("Error checking membership: %v", err)
        }
        if !isJoined {
                return
        }
        
        switch {
        case data == "register_store":
                mb.startStoreRegistration(chatID)
                
        case data == "my_stores":
                mb.showUserPanel(chatID)
                
        case data == "support":
                mb.showSupportMenu(chatID)
                
        case data == "main_menu":
                mb.sendMainMenu(chatID)
                
        case data == "faq":
                msg := tgbotapi.NewMessage(chatID, messages.FAQMessage)
                keyboard := tgbotapi.NewInlineKeyboardMarkup(
                        tgbotapi.NewInlineKeyboardRow(
                                tgbotapi.NewInlineKeyboardButtonData("🔙 بازگشت", "support"),
                        ),
                )
                msg.ReplyMarkup = keyboard
                mb.bot.Send(msg)
                
        case data == "contact_us":
                // Start contact flow
                mb.sessionService.SetUserState(chatID, "contacting_support", map[string]interface{}{"step": 1})
                msg := tgbotapi.NewMessage(chatID, "📝 لطفاً پیام خود را بنویسید:")
                mb.bot.Send(msg)
                
        case data == "telegram_support":
                msg := tgbotapi.NewMessage(chatID, "💬 برای پشتیبانی سریع با ما در تلگرام تماس بگیرید:\n\n@CodeRootSupport")
                keyboard := tgbotapi.NewInlineKeyboardMarkup(
                        tgbotapi.NewInlineKeyboardRow(
                                tgbotapi.NewInlineKeyboardButtonURL("💬 تماس با پشتیبانی", "https://t.me/CodeRootSupport"),
                        ),
                        tgbotapi.NewInlineKeyboardRow(
                                tgbotapi.NewInlineKeyboardButtonData("🔙 بازگشت", "support"),
                        ),
                )
                msg.ReplyMarkup = keyboard
                mb.bot.Send(msg)
                
        case strings.HasPrefix(data, "select_plan_"):
                planType := strings.TrimPrefix(data, "select_plan_")
                mb.handlePlanSelection(chatID, planType)
                
        case strings.HasPrefix(data, "manage_store_"):
                storeIDStr := strings.TrimPrefix(data, "manage_store_")
                storeID, err := strconv.ParseUint(storeIDStr, 10, 32)
                if err == nil {
                        mb.showStoreManagement(chatID, uint(storeID))
                }
                
        case strings.HasPrefix(data, "admin_") && user.IsAdmin:
                mb.handleAdminCallback(chatID, data)
                
        case data == "check_membership":
                // Re-check membership and show main menu if joined
                mb.sendMainMenu(chatID)
        }
}

// handlePlanSelection processes plan selection
func (mb *MotherBot) handlePlanSelection(chatID int64, planType string) {
        var data StoreRegistrationData
        if err := mb.sessionService.GetSessionData(chatID, &data); err != nil {
                log.Printf("Error getting session data: %v", err)
                mb.sendMainMenu(chatID)
                return
        }
        
        data.PlanType = planType
        
        switch planType {
        case "free":
                // Create free store immediately
                mb.createFreeStore(chatID, data)
                
        case "pro", "vip":
                // Show payment instructions
                mb.showPaymentInstructions(chatID, data)
        }
}

// createFreeStore creates a free store
func (mb *MotherBot) createFreeStore(chatID int64, data StoreRegistrationData) {
        user, err := mb.userService.GetOrCreateUser(chatID, "", "", "")
        if err != nil {
                log.Printf("Error getting user: %v", err)
                return
        }
        
        // Check if user already has a free store
        var existingStores []models.Store
        mb.db.Where("owner_id = ? AND plan_type = ?", user.ID, models.PlanFree).Find(&existingStores)
        if len(existingStores) > 0 {
                msg := tgbotapi.NewMessage(chatID, "❌ شما قبلاً یک فروشگاه رایگان دارید. برای ایجاد فروشگاه دوم، پلن حرفه‌ای یا VIP انتخاب کنید.")
                mb.bot.Send(msg)
                return
        }
        
        // Create store
        store := models.Store{
                OwnerID:        user.ID,
                Name:           data.StoreName,
                Description:    data.Description,
                PlanType:       models.PlanFree,
                ExpiresAt:      time.Now().AddDate(1, 0, 0), // 1 year
                IsActive:       true,
                ProductLimit:   10,
                CommissionRate: 5,
        }
        
        if err := mb.db.Create(&store).Error; err != nil {
                log.Printf("Error creating store: %v", err)
                msg := tgbotapi.NewMessage(chatID, messages.ErrorGeneral)
                mb.bot.Send(msg)
                return
        }
        
        // Clear session
        mb.sessionService.ClearUserState(chatID)
        
        // Send success message
        text := fmt.Sprintf(messages.SuccessStoreCreated+"\n\n🏪 نام فروشگاه: %s\n💰 پلن: رایگان\n📦 حداکثر محصول: 10\n📈 کارمزد: 5%%", store.Name)
        
        keyboard := tgbotapi.NewInlineKeyboardMarkup(
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("🏪 ورود به پنل فروشگاه", fmt.Sprintf("manage_store_%d", store.ID)),
                ),
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("🔙 منوی اصلی", "main_menu"),
                ),
        )
        
        msg := tgbotapi.NewMessage(chatID, text)
        msg.ReplyMarkup = keyboard
        mb.bot.Send(msg)
}

// showPaymentInstructions shows payment details for paid plans
func (mb *MotherBot) showPaymentInstructions(chatID int64, data StoreRegistrationData) {
        var price string
        var planName string
        
        switch data.PlanType {
        case "pro":
                price = "50,000"
                planName = "حرفه‌ای"
        case "vip":
                price = "150,000"
                planName = "VIP"
        }
        
        // Store registration data for after payment
        mb.sessionService.SetUserState(chatID, "awaiting_payment", data)
        
        text := fmt.Sprintf(messages.PaymentInstructions, planName, price, "6037-9981-2345-6789", "علی محمدی")
        
        keyboard := tgbotapi.NewInlineKeyboardMarkup(
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("💳 رسید پرداخت ارسال شد", "payment_sent"),
                ),
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("🔙 لغو", "cancel_registration"),
                ),
        )
        
        msg := tgbotapi.NewMessage(chatID, text)
        msg.ReplyMarkup = keyboard
        mb.bot.Send(msg)
}

// showStoreManagement shows individual store management panel
func (mb *MotherBot) showStoreManagement(chatID int64, storeID uint) {
        var store models.Store
        if err := mb.db.Preload("Products").Preload("Orders").First(&store, storeID).Error; err != nil {
                msg := tgbotapi.NewMessage(chatID, messages.ErrorGeneral)
                mb.bot.Send(msg)
                return
        }
        
        // Check ownership
        user, _ := mb.userService.GetOrCreateUser(chatID, "", "", "")
        if store.OwnerID != user.ID {
                msg := tgbotapi.NewMessage(chatID, "❌ شما مالک این فروشگاه نیستید")
                mb.bot.Send(msg)
                return
        }
        
        daysLeft := int(time.Until(store.ExpiresAt).Hours() / 24)
        status := "✅ فعال"
        if daysLeft < 0 {
                status = "❌ منقضی شده"
        } else if !store.IsActive {
                status = "⏸ متوقف شده"
        }
        
        text := fmt.Sprintf(messages.StorePanelMessage, 
                store.Name, 
                string(store.PlanType), 
                store.ExpiresAt.Format("2006/01/02"), 
                len(store.Products), 
                store.ProductLimit)
        
        text += fmt.Sprintf("\n📊 وضعیت: %s", status)
        text += fmt.Sprintf("\n🗓 روزهای باقی‌مانده: %d", daysLeft)
        
        keyboard := tgbotapi.NewInlineKeyboardMarkup(
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData(messages.AddProductBtn, fmt.Sprintf("add_product_%d", storeID)),
                        tgbotapi.NewInlineKeyboardButtonData(messages.ProductListBtn, fmt.Sprintf("product_list_%d", storeID)),
                ),
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData(messages.OrdersBtn, fmt.Sprintf("orders_%d", storeID)),
                        tgbotapi.NewInlineKeyboardButtonData(messages.SalesReportBtn, fmt.Sprintf("sales_report_%d", storeID)),
                ),
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData(messages.StoreSettingsBtn, fmt.Sprintf("store_settings_%d", storeID)),
                ),
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData(messages.RenewPlanBtn, fmt.Sprintf("renew_plan_%d", storeID)),
                ),
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("🔙 بازگشت", "my_stores"),
                ),
        )
        
        msg := tgbotapi.NewMessage(chatID, text)
        msg.ReplyMarkup = keyboard
        mb.bot.Send(msg)
}

// handleAdminCallback processes admin panel callbacks
func (mb *MotherBot) handleAdminCallback(chatID int64, data string) {
        switch {
        case data == "admin_payments":
                mb.showPendingPayments(chatID)
                
        case data == "admin_stores":
                mb.showAllStores(chatID)
                
        case data == "admin_users":
                mb.showAllUsers(chatID)
                
        case data == "admin_broadcast":
                mb.sessionService.SetUserState(chatID, "admin_broadcast", map[string]interface{}{"step": 1})
                msg := tgbotapi.NewMessage(chatID, "📢 پیام همگانی خود را بنویسید:")
                mb.bot.Send(msg)
        }
}

// showPendingPayments shows payments awaiting approval
func (mb *MotherBot) showPendingPayments(chatID int64) {
        var payments []models.Payment
        mb.db.Where("status = ?", "pending").Limit(5).Find(&payments)
        
        if len(payments) == 0 {
                msg := tgbotapi.NewMessage(chatID, "✅ هیچ پرداخت در انتظاری وجود ندارد")
                mb.bot.Send(msg)
                return
        }
        
        text := "💳 پرداخت‌های در انتظار تایید:\n\n"
        keyboard := tgbotapi.NewInlineKeyboardMarkup()
        
        for _, payment := range payments {
                text += fmt.Sprintf("💰 مبلغ: %d تومان\n", payment.Amount)
                text += fmt.Sprintf("📅 تاریخ: %s\n", payment.CreatedAt.Format("2006/01/02 15:04"))
                text += "────────────\n"
                
                row := tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("✅ تایید", fmt.Sprintf("approve_payment_%d", payment.ID)),
                        tgbotapi.NewInlineKeyboardButtonData("❌ رد", fmt.Sprintf("reject_payment_%d", payment.ID)),
                )
                keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
        }
        
        // Add back button
        keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, 
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("🔙 بازگشت", "admin_panel"),
                ))
        
        msg := tgbotapi.NewMessage(chatID, text)
        msg.ReplyMarkup = keyboard
        mb.bot.Send(msg)
}