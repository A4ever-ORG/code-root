package bot

import (
	"encoding/json"
	"fmt"
	"log"

	"telegram-store-hub/internal/messages"
	"telegram-store-hub/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (mb *MotherBot) handlePaymentProof(chatID int64, user *models.User, photos []tgbotapi.PhotoSize, session *models.UserSession) {
	var sessionData map[string]interface{}
	if err := json.Unmarshal([]byte(session.Data), &sessionData); err != nil {
		mb.sendMessage(chatID, messages.ErrorGeneral)
		return
	}

	storeIDFloat, ok := sessionData["store_id"].(float64)
	if !ok {
		mb.sendMessage(chatID, messages.ErrorGeneral)
		return
	}
	storeID := uint(storeIDFloat)

	planType, ok := sessionData["plan_type"].(string)
	if !ok {
		mb.sendMessage(chatID, messages.ErrorGeneral)
		return
	}

	// Get photo URL
	photoURL := mb.getPhotoURL(photos)
	if photoURL == "" {
		mb.sendMessage(chatID, "خطا در دریافت تصویر. لطفاً دوباره تلاش کنید")
		return
	}

	// Get store
	store, err := mb.storeService.GetStoreByID(storeID)
	if err != nil || store.OwnerID != user.ID {
		mb.sendMessage(chatID, messages.ErrorGeneral)
		return
	}

	// Calculate payment amount
	planPrice := mb.getPlanPrice(planType)
	
	// Create payment record
	payment, err := mb.paymentService.CreatePayment(
		storeID,
		int64(planPrice),
		"subscription",
		photoURL,
		fmt.Sprintf("Payment for %s plan - Store: %s", planType, store.Name),
	)
	if err != nil {
		log.Printf("Error creating payment: %v", err)
		mb.sendMessage(chatID, messages.ErrorGeneral)
		return
	}

	// Clear session
	mb.sessionService.ClearSession(user.TelegramID)

	// Send confirmation to user
	mb.sendMessage(chatID, messages.PaymentReceived)

	// Notify admin
	mb.notifyAdminNewPayment(payment, store, user)
}

func (mb *MotherBot) notifyAdminNewPayment(payment *models.Payment, store *models.Store, user *models.User) {
	if mb.config.AdminChatID == 0 {
		return
	}

	planName := mb.getPlanName(store.PlanType)
	priceFormatted := mb.formatPrice(int(payment.Amount))

	adminText := fmt.Sprintf(messages.AdminNewStoreNotification,
		user.FirstName,
		user.Username,
		store.Name,
		planName,
		priceFormatted,
	)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(messages.AdminApproveBtn, fmt.Sprintf("approve_payment_%d", payment.ID)),
			tgbotapi.NewInlineKeyboardButtonData(messages.AdminRejectBtn, fmt.Sprintf("reject_payment_%d", payment.ID)),
		),
	)

	// Send payment proof image with admin notification
	if payment.ProofImageURL != "" {
		photoMsg := tgbotapi.NewPhoto(mb.config.AdminChatID, tgbotapi.FileURL(payment.ProofImageURL))
		photoMsg.Caption = adminText
		photoMsg.ReplyMarkup = keyboard
		mb.bot.Send(photoMsg)
	} else {
		msg := tgbotapi.NewMessage(mb.config.AdminChatID, adminText)
		msg.ReplyMarkup = keyboard
		mb.bot.Send(msg)
	}
}

func (mb *MotherBot) processPaymentApproval(paymentID uint, adminUser *models.User, approved bool) error {
	payment, err := mb.paymentService.GetPaymentByID(paymentID)
	if err != nil {
		return err
	}

	if approved {
		// Approve payment
		err = mb.paymentService.ApprovePayment(paymentID, adminUser.ID)
		if err != nil {
			return err
		}

		// Activate store
		err = mb.activateStore(payment.StoreID, &payment.Store.Owner)
		if err != nil {
			return err
		}

		// Notify store owner
		store, _ := mb.storeService.GetStoreByID(payment.StoreID)
		successMessage := fmt.Sprintf(messages.PaymentApproved, store.BotUsername, store.BotToken)
		mb.sendMessage(payment.Store.Owner.TelegramID, successMessage)

	} else {
		// Reject payment
		err = mb.paymentService.RejectPayment(paymentID, adminUser.ID)
		if err != nil {
			return err
		}

		// Notify store owner
		rejectionMessage := fmt.Sprintf("❌ پرداخت شما برای فروشگاه '%s' رد شد.\n\n📞 برای اطلاعات بیشتر با پشتیبانی تماس بگیرید.", payment.Store.Name)
		mb.sendMessage(payment.Store.Owner.TelegramID, rejectionMessage)
	}

	return nil
}

func (mb *MotherBot) handleSubscriptionRenewal(chatID int64, user *models.User, storeID uint, months int) {
	store, err := mb.storeService.GetStoreByID(storeID)
	if err != nil || store.OwnerID != user.ID {
		mb.sendMessage(chatID, messages.ErrorGeneral)
		return
	}

	// Calculate renewal price
	basePlanPrice := mb.getPlanPrice(store.PlanType)
	renewalPrice := basePlanPrice * months

	if renewalPrice == 0 {
		// Free plan renewal
		err = mb.storeService.RenewStore(storeID, months)
		if err != nil {
			mb.sendMessage(chatID, messages.ErrorGeneral)
			return
		}

		mb.sendMessage(chatID, fmt.Sprintf("✅ پلن رایگان شما برای %d ماه تمدید شد", months))
		return
	}

	// Paid plan renewal - show payment instructions
	planName := mb.getPlanName(store.PlanType)
	renewalText := fmt.Sprintf(`🔄 تمدید پلن %s

📅 مدت: %d ماه
💰 مبلغ: %s تومان

📋 شماره کارت:
%s

👤 به نام: %s

پس از پرداخت، عکس رسید را ارسال کنید.`,
		planName,
		months,
		mb.formatPrice(renewalPrice),
		mb.config.PaymentCardNumber,
		mb.config.PaymentCardHolder,
	)

	// Set session for renewal payment
	sessionData := map[string]interface{}{
		"store_id":       storeID,
		"renewal_months": months,
		"renewal_price":  renewalPrice,
		"plan_type":      store.PlanType,
	}
	data, _ := json.Marshal(sessionData)
	mb.sessionService.SetSession(user.TelegramID, "renewal_payment", string(data))

	mb.sendMessage(chatID, renewalText)
}

func (mb *MotherBot) handleRenewalPaymentProof(chatID int64, user *models.User, photos []tgbotapi.PhotoSize, session *models.UserSession) {
	var sessionData map[string]interface{}
	if err := json.Unmarshal([]byte(session.Data), &sessionData); err != nil {
		mb.sendMessage(chatID, messages.ErrorGeneral)
		return
	}

	storeIDFloat := sessionData["store_id"].(float64)
	storeID := uint(storeIDFloat)
	renewalMonthsFloat := sessionData["renewal_months"].(float64)
	renewalMonths := int(renewalMonthsFloat)
	renewalPriceFloat := sessionData["renewal_price"].(float64)
	renewalPrice := int64(renewalPriceFloat)

	// Get photo URL
	photoURL := mb.getPhotoURL(photos)
	if photoURL == "" {
		mb.sendMessage(chatID, "خطا در دریافت تصویر. لطفاً دوباره تلاش کنید")
		return
	}

	// Get store
	store, err := mb.storeService.GetStoreByID(storeID)
	if err != nil || store.OwnerID != user.ID {
		mb.sendMessage(chatID, messages.ErrorGeneral)
		return
	}

	// Create payment record
	payment, err := mb.paymentService.CreatePayment(
		storeID,
		renewalPrice,
		"renewal",
		photoURL,
		fmt.Sprintf("Renewal payment for %d months - Store: %s", renewalMonths, store.Name),
	)
	if err != nil {
		log.Printf("Error creating renewal payment: %v", err)
		mb.sendMessage(chatID, messages.ErrorGeneral)
		return
	}

	// Clear session
	mb.sessionService.ClearSession(user.TelegramID)

	// Send confirmation
	mb.sendMessage(chatID, "✅ رسید تمدید دریافت شد\n\n🔄 در حال بررسی توسط ادمین...\n📞 پس از تایید، پلن شما تمدید خواهد شد")

	// Notify admin about renewal payment
	mb.notifyAdminRenewalPayment(payment, store, user, renewalMonths)
}

func (mb *MotherBot) notifyAdminRenewalPayment(payment *models.Payment, store *models.Store, user *models.User, months int) {
	if mb.config.AdminChatID == 0 {
		return
	}

	planName := mb.getPlanName(store.PlanType)
	priceFormatted := mb.formatPrice(int(payment.Amount))

	adminText := fmt.Sprintf(`🔄 درخواست تمدید جدید

🏪 فروشگاه: %s
👤 مالک: %s (@%s)
💎 پلن: %s
📅 مدت: %d ماه
💰 مبلغ: %s تومان
📅 انقضای فعلی: %s`,
		store.Name,
		user.FirstName,
		user.Username,
		planName,
		months,
		priceFormatted,
		store.ExpiresAt.Format("2006/01/02"),
	)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ تایید تمدید", fmt.Sprintf("approve_renewal_%d_%d", payment.ID, months)),
			tgbotapi.NewInlineKeyboardButtonData("❌ رد", fmt.Sprintf("reject_payment_%d", payment.ID)),
		),
	)

	// Send payment proof with admin notification
	if payment.ProofImageURL != "" {
		photoMsg := tgbotapi.NewPhoto(mb.config.AdminChatID, tgbotapi.FileURL(payment.ProofImageURL))
		photoMsg.Caption = adminText
		photoMsg.ReplyMarkup = keyboard
		mb.bot.Send(photoMsg)
	} else {
		msg := tgbotapi.NewMessage(mb.config.AdminChatID, adminText)
		msg.ReplyMarkup = keyboard
		mb.bot.Send(msg)
	}
}

func (mb *MotherBot) processRenewalApproval(paymentID uint, months int, adminUser *models.User) error {
	payment, err := mb.paymentService.GetPaymentByID(paymentID)
	if err != nil {
		return err
	}

	// Approve payment
	err = mb.paymentService.ApprovePayment(paymentID, adminUser.ID)
	if err != nil {
		return err
	}

	// Renew store
	err = mb.storeService.RenewStore(payment.StoreID, months)
	if err != nil {
		return err
	}

	// Get updated store info
	store, _ := mb.storeService.GetStoreByID(payment.StoreID)

	// Notify store owner
	renewalMessage := fmt.Sprintf(`✅ پلن شما با موفقیت تمدید شد!

🏪 فروشگاه: %s
📅 تاریخ انقضای جدید: %s
💎 پلن: %s`,
		store.Name,
		store.ExpiresAt.Format("2006/01/02"),
		mb.getPlanName(store.PlanType),
	)

	mb.sendMessage(payment.Store.Owner.TelegramID, renewalMessage)
	return nil
}