package bot

import (
        "encoding/json"
        "fmt"
        "log"
        "strconv"
        "strings"

        "telegram-store-hub/internal/messages"
        "telegram-store-hub/internal/models"

        tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (mb *MotherBot) handleCallbackQuery(callback *tgbotapi.CallbackQuery) {
        chatID := callback.Message.Chat.ID
        userID := callback.From.ID
        data := callback.Data

        // Answer callback query
        mb.bot.Request(tgbotapi.NewCallback(callback.ID, ""))

        // Get user
        user, err := mb.userService.GetUserByTelegramID(userID)
        if err != nil {
                log.Printf("Error getting user: %v", err)
                return
        }

        // Handle different callback types
        switch {
        case strings.HasPrefix(data, "plan_"):
                mb.handlePlanSelection(chatID, user, data)
        case strings.HasPrefix(data, "manage_store_"):
                mb.handleStoreManagement(chatID, user, data)
        case strings.HasPrefix(data, "add_product_"):
                mb.handleAddProduct(chatID, user, data)
        case strings.HasPrefix(data, "product_list_"):
                mb.handleProductList(chatID, user, data)
        case strings.HasPrefix(data, "orders_"):
                mb.handleOrdersList(chatID, user, data)
        case strings.HasPrefix(data, "sales_"):
                mb.handleSalesReport(chatID, user, data)
        case strings.HasPrefix(data, "renew_"):
                mb.handleRenewPlan(chatID, user, data)
        case strings.HasPrefix(data, "settings_"):
                mb.handleStoreSettings(chatID, user, data)
        case strings.HasPrefix(data, "edit_product_"):
                mb.handleProductEdit(chatID, user, data)
        case strings.HasPrefix(data, "delete_product_"):
                mb.handleProductDelete(chatID, user, data)
        case strings.HasPrefix(data, "confirm_delete_product_"):
                mb.handleConfirmProductDelete(chatID, user, data)
        case strings.HasPrefix(data, "toggle_product_"):
                mb.handleToggleProduct(chatID, user, data)
        case strings.HasPrefix(data, "admin_"):
                if user.IsAdmin {
                        mb.handleAdminCallback(chatID, user, data)
                }
        case strings.HasPrefix(data, "approve_payment_"):
                if user.IsAdmin {
                        mb.handlePaymentApproval(chatID, user, data, true)
                }
        case strings.HasPrefix(data, "reject_payment_"):
                if user.IsAdmin {
                        mb.handlePaymentApproval(chatID, user, data, false)
                }
        case strings.HasPrefix(data, "approve_renewal_"):
                if user.IsAdmin {
                        mb.handleRenewalApproval(chatID, user, data)
                }
        case data == "faq":
                mb.sendMessage(chatID, messages.FAQMessage)
        case data == "contact":
                mb.sendMessage(chatID, "📞 تماس با ما: @support")
        case data == "telegram_support":
                mb.sendMessage(chatID, "💬 پشتیبانی تلگرام: @telegram_support")
        case data == "check_membership":
                if mb.checkChannelMembership(userID) {
                        mb.sendWelcomeMessage(chatID)
                } else {
                        mb.sendForceJoinMessage(chatID)
                }
        default:
                mb.sendMessage(chatID, "❌ دستور نامعتبر")
        }
}

func (mb *MotherBot) handlePlanSelection(chatID int64, user *models.User, data string) {
        planType := strings.TrimPrefix(data, "plan_")
        
        // Get session data
        session, err := mb.sessionService.GetSession(user.TelegramID)
        if err != nil || session == nil || session.State != "plan_selection" {
                mb.sendMessage(chatID, messages.ErrorGeneral)
                return
        }

        var sessionData map[string]interface{}
        json.Unmarshal([]byte(session.Data), &sessionData)

        // Add plan to session data
        sessionData["plan_type"] = planType

        // Create store
        storeName := sessionData["store_name"].(string)
        storeDescription := sessionData["store_description"].(string)

        store, err := mb.storeService.CreateStore(user.ID, storeName, storeDescription, planType)
        if err != nil {
                log.Printf("Error creating store: %v", err)
                mb.sendMessage(chatID, messages.ErrorGeneral)
                return
        }

        // Add store ID to session
        sessionData["store_id"] = store.ID
        data, _ := json.Marshal(sessionData)
        mb.sessionService.SetSession(user.TelegramID, "payment_proof", string(data))

        // Send payment instructions
        planPrice := mb.getPlanPrice(planType)
        planName := mb.getPlanName(planType)
        
        if planType == "free" {
                // Free plan - activate immediately
                mb.activateStore(store.ID, user)
                mb.sessionService.ClearSession(user.TelegramID)
                mb.sendMessage(chatID, "🎉 فروشگاه رایگان شما فعال شد! از منوی اصلی 'فروشگاه‌های من' را انتخاب کنید.")
        } else {
                // Paid plan - require payment
                paymentText := fmt.Sprintf(messages.PaymentInstructions,
                        planName,
                        mb.formatPrice(planPrice),
                        mb.config.PaymentCardNumber,
                        mb.config.PaymentCardHolder,
                )
                mb.sendMessage(chatID, paymentText)
        }
}

func (mb *MotherBot) handleStoreManagement(chatID int64, user *models.User, data string) {
        storeIDStr := strings.TrimPrefix(data, "manage_store_")
        storeID, err := strconv.Atoi(storeIDStr)
        if err != nil {
                mb.sendMessage(chatID, messages.ErrorGeneral)
                return
        }

        store, err := mb.storeService.GetStoreByID(uint(storeID))
        if err != nil || store.OwnerID != user.ID {
                mb.sendMessage(chatID, messages.ErrorGeneral)
                return
        }

        mb.showStorePanelForStore(chatID, store)
}

func (mb *MotherBot) handleAddProduct(chatID int64, user *models.User, data string) {
        storeIDStr := strings.TrimPrefix(data, "add_product_")
        storeID, err := strconv.Atoi(storeIDStr)
        if err != nil {
                mb.sendMessage(chatID, messages.ErrorGeneral)
                return
        }

        store, err := mb.storeService.GetStoreByID(uint(storeID))
        if err != nil || store.OwnerID != user.ID {
                mb.sendMessage(chatID, messages.ErrorGeneral)
                return
        }

        // Check product limit
        productsCount, _ := mb.productService.GetProductsCount(uint(storeID))
        if store.ProductLimit != -1 && int(productsCount) >= store.ProductLimit {
                mb.sendMessage(chatID, messages.ErrorStoreLimit)
                return
        }

        // Start product creation flow
        sessionData := map[string]interface{}{
                "store_id": storeID,
        }
        data, _ := json.Marshal(sessionData)
        mb.sessionService.SetSession(user.TelegramID, "product_name", string(data))
        mb.sendMessage(chatID, messages.AddProductStart)
}

func (mb *MotherBot) handleProductList(chatID int64, user *models.User, data string) {
        storeIDStr := strings.TrimPrefix(data, "product_list_")
        storeID, err := strconv.Atoi(storeIDStr)
        if err != nil {
                mb.sendMessage(chatID, messages.ErrorGeneral)
                return
        }

        store, err := mb.storeService.GetStoreByID(uint(storeID))
        if err != nil || store.OwnerID != user.ID {
                mb.sendMessage(chatID, messages.ErrorGeneral)
                return
        }

        products, err := mb.productService.GetStoreProducts(uint(storeID))
        if err != nil {
                mb.sendMessage(chatID, messages.ErrorGeneral)
                return
        }

        if len(products) == 0 {
                mb.sendMessage(chatID, "📦 هیچ محصولی ثبت نشده است")
                return
        }

        for _, product := range products {
                productText := fmt.Sprintf(`📦 %s
💰 قیمت: %s تومان
📝 توضیحات: %s
✅ وضعیت: %s`,
                        product.Name,
                        mb.formatPrice(int(product.Price)),
                        product.Description,
                        func() string {
                                if product.IsAvailable {
                                        return "فعال"
                                }
                                return "غیرفعال"
                        }(),
                )

                keyboard := tgbotapi.NewInlineKeyboardMarkup(
                        tgbotapi.NewInlineKeyboardRow(
                                tgbotapi.NewInlineKeyboardButtonData("✏️ ویرایش", fmt.Sprintf("edit_product_%d", product.ID)),
                                tgbotapi.NewInlineKeyboardButtonData("🗑 حذف", fmt.Sprintf("delete_product_%d", product.ID)),
                        ),
                )

                msg := tgbotapi.NewMessage(chatID, productText)
                msg.ReplyMarkup = keyboard
                mb.bot.Send(msg)
        }
}

func (mb *MotherBot) handleOrdersList(chatID int64, user *models.User, data string) {
        storeIDStr := strings.TrimPrefix(data, "orders_")
        storeID, err := strconv.Atoi(storeIDStr)
        if err != nil {
                mb.sendMessage(chatID, messages.ErrorGeneral)
                return
        }

        store, err := mb.storeService.GetStoreByID(uint(storeID))
        if err != nil || store.OwnerID != user.ID {
                mb.sendMessage(chatID, messages.ErrorGeneral)
                return
        }

        // This would be implemented with OrderService
        mb.sendMessage(chatID, "🛒 بخش سفارش‌ها در حال توسعه است")
}

func (mb *MotherBot) handleSalesReport(chatID int64, user *models.User, data string) {
        storeIDStr := strings.TrimPrefix(data, "sales_")
        storeID, err := strconv.Atoi(storeIDStr)
        if err != nil {
                mb.sendMessage(chatID, messages.ErrorGeneral)
                return
        }

        store, err := mb.storeService.GetStoreByID(uint(storeID))
        if err != nil || store.OwnerID != user.ID {
                mb.sendMessage(chatID, messages.ErrorGeneral)
                return
        }

        // This would be implemented with OrderService
        mb.sendMessage(chatID, "📈 بخش گزارش فروش در حال توسعه است")
}

func (mb *MotherBot) handleRenewPlan(chatID int64, user *models.User, data string) {
        storeIDStr := strings.TrimPrefix(data, "renew_")
        storeID, err := strconv.Atoi(storeIDStr)
        if err != nil {
                mb.sendMessage(chatID, messages.ErrorGeneral)
                return
        }

        store, err := mb.storeService.GetStoreByID(uint(storeID))
        if err != nil || store.OwnerID != user.ID {
                mb.sendMessage(chatID, messages.ErrorGeneral)
                return
        }

        // Show renewal options
        keyboard := tgbotapi.NewInlineKeyboardMarkup(
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("📅 ۱ ماه", fmt.Sprintf("renew_months_%d_1", storeID)),
                        tgbotapi.NewInlineKeyboardButtonData("📅 ۳ ماه", fmt.Sprintf("renew_months_%d_3", storeID)),
                ),
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("📅 ۶ ماه", fmt.Sprintf("renew_months_%d_6", storeID)),
                        tgbotapi.NewInlineKeyboardButtonData("📅 ۱ سال", fmt.Sprintf("renew_months_%d_12", storeID)),
                ),
        )

        renewText := fmt.Sprintf("🔄 تمدید پلن %s\n\nمدت زمان تمدید را انتخاب کنید:", strings.ToUpper(store.PlanType))
        msg := tgbotapi.NewMessage(chatID, renewText)
        msg.ReplyMarkup = keyboard
        mb.bot.Send(msg)
}

func (mb *MotherBot) handleStoreSettings(chatID int64, user *models.User, data string) {
        storeIDStr := strings.TrimPrefix(data, "settings_")
        storeID, err := strconv.Atoi(storeIDStr)
        if err != nil {
                mb.sendMessage(chatID, messages.ErrorGeneral)
                return
        }

        store, err := mb.storeService.GetStoreByID(uint(storeID))
        if err != nil || store.OwnerID != user.ID {
                mb.sendMessage(chatID, messages.ErrorGeneral)
                return
        }

        settingsText := fmt.Sprintf(`⚙️ تنظیمات فروشگاه %s

🤖 نام کاربری ربات: @%s
📝 پیام خوش‌آمدگویی: %s
📞 اطلاعات تماس: %s`,
                store.Name,
                store.BotUsername,
                func() string {
                        if store.WelcomeMessage != "" {
                                return store.WelcomeMessage
                        }
                        return "تنظیم نشده"
                }(),
                func() string {
                        if store.SupportContact != "" {
                                return store.SupportContact
                        }
                        return "تنظیم نشده"
                }(),
        )

        keyboard := tgbotapi.NewInlineKeyboardMarkup(
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("✏️ پیام خوش‌آمدگویی", fmt.Sprintf("edit_welcome_%d", storeID)),
                ),
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("📞 اطلاعات تماس", fmt.Sprintf("edit_contact_%d", storeID)),
                ),
        )

        msg := tgbotapi.NewMessage(chatID, settingsText)
        msg.ReplyMarkup = keyboard
        mb.bot.Send(msg)
}

func (mb *MotherBot) handleAdminCallback(chatID int64, user *models.User, data string) {
        switch data {
        case "admin_stats":
                mb.showAdminStats(chatID)
        case "admin_stores":
                mb.showAdminStores(chatID)
        case "admin_payments":
                mb.showAdminPayments(chatID)
        case "admin_broadcast":
                mb.sendMessage(chatID, "📢 بخش ارسال پیام همگانی در حال توسعه است")
        }
}

// Additional callback handlers
func (mb *MotherBot) handleProductEdit(chatID int64, user *models.User, data string) {
        productIDStr := strings.TrimPrefix(data, "edit_product_")
        productID, err := strconv.Atoi(productIDStr)
        if err != nil {
                mb.sendMessage(chatID, messages.ErrorGeneral)
                return
        }

        mb.handleProductEdit(chatID, user, uint(productID))
}

func (mb *MotherBot) handleProductDelete(chatID int64, user *models.User, data string) {
        productIDStr := strings.TrimPrefix(data, "delete_product_")
        productID, err := strconv.Atoi(productIDStr)
        if err != nil {
                mb.sendMessage(chatID, messages.ErrorGeneral)
                return
        }

        mb.handleProductDelete(chatID, user, uint(productID))
}

func (mb *MotherBot) handleConfirmProductDelete(chatID int64, user *models.User, data string) {
        productIDStr := strings.TrimPrefix(data, "confirm_delete_product_")
        productID, err := strconv.Atoi(productIDStr)
        if err != nil {
                mb.sendMessage(chatID, messages.ErrorGeneral)
                return
        }

        mb.confirmProductDelete(chatID, user, uint(productID))
}

func (mb *MotherBot) handleToggleProduct(chatID int64, user *models.User, data string) {
        productIDStr := strings.TrimPrefix(data, "toggle_product_")
        productID, err := strconv.Atoi(productIDStr)
        if err != nil {
                mb.sendMessage(chatID, messages.ErrorGeneral)
                return
        }

        mb.toggleProductAvailability(chatID, user, uint(productID))
}

func (mb *MotherBot) handleRenewalApproval(chatID int64, user *models.User, data string) {
        // Parse renewal approval data: approve_renewal_PAYMENTID_MONTHS
        parts := strings.Split(data, "_")
        if len(parts) < 4 {
                mb.sendMessage(chatID, messages.ErrorGeneral)
                return
        }

        paymentID, err1 := strconv.Atoi(parts[2])
        months, err2 := strconv.Atoi(parts[3])
        if err1 != nil || err2 != nil {
                mb.sendMessage(chatID, messages.ErrorGeneral)
                return
        }

        err := mb.processRenewalApproval(uint(paymentID), months, user)
        if err != nil {
                mb.sendMessage(chatID, messages.ErrorGeneral)
        } else {
                mb.sendMessage(chatID, "✅ تمدید تایید شد و فروشگاه تمدید شد")
        }
}

func (mb *MotherBot) handlePaymentApproval(chatID int64, user *models.User, data string, approve bool) {
        var paymentIDStr string
        if approve {
                paymentIDStr = strings.TrimPrefix(data, "approve_payment_")
        } else {
                paymentIDStr = strings.TrimPrefix(data, "reject_payment_")
        }

        paymentID, err := strconv.Atoi(paymentIDStr)
        if err != nil {
                mb.sendMessage(chatID, messages.ErrorGeneral)
                return
        }

        payment, err := mb.paymentService.GetPaymentByID(uint(paymentID))
        if err != nil {
                mb.sendMessage(chatID, messages.ErrorGeneral)
                return
        }

        if approve {
                err = mb.paymentService.ApprovePayment(uint(paymentID), user.ID)
                if err == nil {
                        // Activate store
                        mb.activateStore(payment.StoreID, &payment.Store.Owner)
                        mb.sendMessage(chatID, "✅ پرداخت تایید شد و فروشگاه فعال شد")
                        
                        // Notify store owner
                        mb.sendMessage(payment.Store.Owner.TelegramID, fmt.Sprintf(messages.PaymentApproved, payment.Store.BotUsername, payment.Store.BotToken))
                }
        } else {
                err = mb.paymentService.RejectPayment(uint(paymentID), user.ID)
                if err == nil {
                        mb.sendMessage(chatID, "❌ پرداخت رد شد")
                        // Notify store owner
                        mb.sendMessage(payment.Store.Owner.TelegramID, "❌ پرداخت شما رد شد. لطفاً با پشتیبانی تماس بگیرید.")
                }
        }

        if err != nil {
                mb.sendMessage(chatID, messages.ErrorGeneral)
        }
}

// Helper methods
func (mb *MotherBot) getPlanPrice(planType string) int {
        switch planType {
        case "free":
                return mb.config.FreePlanPrice
        case "pro":
                return mb.config.ProPlanPrice
        case "vip":
                return mb.config.VIPPlanPrice
        default:
                return 0
        }
}

func (mb *MotherBot) getPlanName(planType string) string {
        switch planType {
        case "free":
                return "رایگان"
        case "pro":
                return "حرفه‌ای"
        case "vip":
                return "VIP"
        default:
                return "نامشخص"
        }
}

func (mb *MotherBot) formatPrice(price int) string {
        // Simple number formatting with commas
        str := strconv.Itoa(price)
        result := ""
        for i, char := range str {
                if i > 0 && (len(str)-i)%3 == 0 {
                        result += ","
                }
                result += string(char)
        }
        return result
}

func (mb *MotherBot) activateStore(storeID uint, owner *models.User) error {
        // Generate bot username and token (mock implementation)
        store, err := mb.storeService.GetStoreByID(storeID)
        if err != nil {
                return err
        }

        botUsername := mb.storeService.GenerateBotUsername(store.Name, storeID)
        botToken := fmt.Sprintf("mock_token_%d", storeID) // In real implementation, create actual bot

        return mb.storeService.ActivateStore(storeID, botToken, botUsername)
}

func (mb *MotherBot) showAdminStats(chatID int64) {
        usersCount, _ := mb.userService.GetUsersCount()
        storesCount, _ := mb.storeService.GetStoresCount()
        activeStoresCount, _ := mb.storeService.GetActiveStoresCount()
        totalRevenue, _ := mb.paymentService.GetTotalRevenue()

        statsText := fmt.Sprintf(`📊 آمار کلی سیستم

👥 تعداد کاربران: %d
🏪 تعداد فروشگاه‌ها: %d
✅ فروشگاه‌های فعال: %d
💰 درآمد کل: %s تومان`,
                usersCount,
                storesCount,
                activeStoresCount,
                mb.formatPrice(int(totalRevenue)),
        )

        mb.sendMessage(chatID, statsText)
}

func (mb *MotherBot) showAdminStores(chatID int64) {
        pendingStores, _ := mb.storeService.GetPendingStores()
        
        if len(pendingStores) == 0 {
                mb.sendMessage(chatID, "هیچ فروشگاهی در انتظار تایید نیست")
                return
        }

        for _, store := range pendingStores {
                storeText := fmt.Sprintf(`🏪 %s
👤 مالک: %s (@%s)
💎 پلن: %s
📅 تاریخ ثبت: %s`,
                        store.Name,
                        store.Owner.FirstName,
                        store.Owner.Username,
                        strings.ToUpper(store.PlanType),
                        store.CreatedAt.Format("2006/01/02"),
                )

                keyboard := tgbotapi.NewInlineKeyboardMarkup(
                        tgbotapi.NewInlineKeyboardRow(
                                tgbotapi.NewInlineKeyboardButtonData("✅ تایید", fmt.Sprintf("approve_store_%d", store.ID)),
                                tgbotapi.NewInlineKeyboardButtonData("❌ رد", fmt.Sprintf("reject_store_%d", store.ID)),
                        ),
                )

                msg := tgbotapi.NewMessage(chatID, storeText)
                msg.ReplyMarkup = keyboard
                mb.bot.Send(msg)
        }
}

func (mb *MotherBot) showAdminPayments(chatID int64) {
        pendingPayments, _ := mb.paymentService.GetPendingPayments()
        
        if len(pendingPayments) == 0 {
                mb.sendMessage(chatID, "هیچ پرداختی در انتظار تایید نیست")
                return
        }

        for _, payment := range pendingPayments {
                paymentText := fmt.Sprintf(`💳 پرداخت جدید

🏪 فروشگاه: %s
👤 مالک: %s (@%s)
💰 مبلغ: %s تومان
📅 تاریخ: %s`,
                        payment.Store.Name,
                        payment.Store.Owner.FirstName,
                        payment.Store.Owner.Username,
                        mb.formatPrice(int(payment.Amount)),
                        payment.CreatedAt.Format("2006/01/02 15:04"),
                )

                keyboard := tgbotapi.NewInlineKeyboardMarkup(
                        tgbotapi.NewInlineKeyboardRow(
                                tgbotapi.NewInlineKeyboardButtonData("✅ تایید", fmt.Sprintf("approve_payment_%d", payment.ID)),
                                tgbotapi.NewInlineKeyboardButtonData("❌ رد", fmt.Sprintf("reject_payment_%d", payment.ID)),
                        ),
                )

                // Send payment proof image if available
                if payment.ProofImageURL != "" {
                        photoMsg := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(payment.ProofImageURL))
                        photoMsg.Caption = paymentText
                        photoMsg.ReplyMarkup = keyboard
                        mb.bot.Send(photoMsg)
                } else {
                        msg := tgbotapi.NewMessage(chatID, paymentText)
                        msg.ReplyMarkup = keyboard
                        mb.bot.Send(msg)
                }
        }
}