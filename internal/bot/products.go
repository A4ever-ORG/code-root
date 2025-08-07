package bot

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"telegram-store-hub/internal/messages"
	"telegram-store-hub/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (mb *MotherBot) handleProductName(chatID int64, user *models.User, productName string, session *models.UserSession) {
	if len(strings.TrimSpace(productName)) < 2 {
		mb.sendMessage(chatID, "نام محصول باید حداقل ۲ کاراکتر باشد")
		return
	}

	var sessionData map[string]interface{}
	json.Unmarshal([]byte(session.Data), &sessionData)
	
	sessionData["product_name"] = productName
	data, _ := json.Marshal(sessionData)
	
	mb.sessionService.SetSession(user.TelegramID, "product_description", string(data))
	mb.sendMessage(chatID, fmt.Sprintf(messages.ProductNameReceived, productName))
}

func (mb *MotherBot) handleProductDescription(chatID int64, user *models.User, description string, session *models.UserSession) {
	if len(strings.TrimSpace(description)) < 5 {
		mb.sendMessage(chatID, "توضیحات محصول باید حداقل ۵ کاراکتر باشد")
		return
	}

	var sessionData map[string]interface{}
	json.Unmarshal([]byte(session.Data), &sessionData)
	
	sessionData["product_description"] = description
	data, _ := json.Marshal(sessionData)
	
	mb.sessionService.SetSession(user.TelegramID, "product_price", string(data))
	mb.sendMessage(chatID, messages.ProductDescReceived)
}

func (mb *MotherBot) handleProductPrice(chatID int64, user *models.User, priceStr string, session *models.UserSession) {
	// Remove commas and parse price
	cleanPriceStr := strings.ReplaceAll(priceStr, ",", "")
	price, err := strconv.ParseInt(cleanPriceStr, 10, 64)
	if err != nil || price <= 0 {
		mb.sendMessage(chatID, "قیمت وارد شده نامعتبر است. لطفاً یک عدد معتبر وارد کنید")
		return
	}

	if price > 999999999 {
		mb.sendMessage(chatID, "قیمت وارد شده خیلی زیاد است")
		return
	}

	var sessionData map[string]interface{}
	json.Unmarshal([]byte(session.Data), &sessionData)
	
	sessionData["product_price"] = price
	data, _ := json.Marshal(sessionData)
	
	mb.sessionService.SetSession(user.TelegramID, "product_image", string(data))
	mb.sendMessage(chatID, fmt.Sprintf(messages.ProductPriceReceived, mb.formatPrice(int(price))))
}

func (mb *MotherBot) finalizeProduct(chatID int64, user *models.User, imageURL string, session *models.UserSession) {
	var sessionData map[string]interface{}
	json.Unmarshal([]byte(session.Data), &sessionData)

	storeIDFloat := sessionData["store_id"].(float64)
	storeID := uint(storeIDFloat)
	productName := sessionData["product_name"].(string)
	productDescription := sessionData["product_description"].(string)
	productPriceFloat := sessionData["product_price"].(float64)
	productPrice := int64(productPriceFloat)

	// Create product
	product, err := mb.productService.CreateProduct(
		storeID,
		productName,
		productDescription,
		productPrice,
		imageURL,
		"general", // default category
	)
	
	if err != nil {
		mb.sendMessage(chatID, messages.ErrorGeneral)
		return
	}

	// Clear session
	mb.sessionService.ClearSession(user.TelegramID)

	// Send success message
	successText := fmt.Sprintf(messages.ProductAdded, 
		product.Name, 
		mb.formatPrice(int(product.Price)))
	
	mb.sendMessage(chatID, successText)

	// Show product management options
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ افزودن محصول دیگر", fmt.Sprintf("add_product_%d", storeID)),
			tgbotapi.NewInlineKeyboardButtonData("📦 مشاهده لیست", fmt.Sprintf("product_list_%d", storeID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 بازگشت به پنل", fmt.Sprintf("manage_store_%d", storeID)),
		),
	)

	msg := tgbotapi.NewMessage(chatID, "محصول با موفقیت اضافه شد! چه کار می‌خواهید انجام دهید؟")
	msg.ReplyMarkup = keyboard
	mb.bot.Send(msg)
}

func (mb *MotherBot) handleProductEdit(chatID int64, user *models.User, productID uint) {
	product, err := mb.productService.GetProductByID(productID)
	if err != nil || product.Store.OwnerID != user.ID {
		mb.sendMessage(chatID, messages.ErrorGeneral)
		return
	}

	productText := fmt.Sprintf(`✏️ ویرایش محصول

📦 نام: %s
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
			tgbotapi.NewInlineKeyboardButtonData("✏️ نام", fmt.Sprintf("edit_product_name_%d", productID)),
			tgbotapi.NewInlineKeyboardButtonData("💰 قیمت", fmt.Sprintf("edit_product_price_%d", productID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📝 توضیحات", fmt.Sprintf("edit_product_desc_%d", productID)),
			tgbotapi.NewInlineKeyboardButtonData("🖼 تصویر", fmt.Sprintf("edit_product_image_%d", productID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				func() string {
					if product.IsAvailable {
						return "❌ غیرفعال کردن"
					}
					return "✅ فعال کردن"
				}(),
				fmt.Sprintf("toggle_product_%d", productID),
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 بازگشت", fmt.Sprintf("product_list_%d", product.StoreID)),
		),
	)

	msg := tgbotapi.NewMessage(chatID, productText)
	msg.ReplyMarkup = keyboard
	mb.bot.Send(msg)
}

func (mb *MotherBot) handleProductDelete(chatID int64, user *models.User, productID uint) {
	product, err := mb.productService.GetProductByID(productID)
	if err != nil || product.Store.OwnerID != user.ID {
		mb.sendMessage(chatID, messages.ErrorGeneral)
		return
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ بله، حذف کن", fmt.Sprintf("confirm_delete_product_%d", productID)),
			tgbotapi.NewInlineKeyboardButtonData("❌ انصراف", fmt.Sprintf("product_list_%d", product.StoreID)),
		),
	)

	confirmText := fmt.Sprintf("⚠️ آیا مطمئن هستید که محصول '%s' را حذف کنید؟\n\n⚠️ این عمل قابل بازگشت نیست!", product.Name)
	msg := tgbotapi.NewMessage(chatID, confirmText)
	msg.ReplyMarkup = keyboard
	mb.bot.Send(msg)
}

func (mb *MotherBot) confirmProductDelete(chatID int64, user *models.User, productID uint) {
	product, err := mb.productService.GetProductByID(productID)
	if err != nil || product.Store.OwnerID != user.ID {
		mb.sendMessage(chatID, messages.ErrorGeneral)
		return
	}

	storeID := product.StoreID
	productName := product.Name

	err = mb.productService.DeleteProduct(productID)
	if err != nil {
		mb.sendMessage(chatID, messages.ErrorGeneral)
		return
	}

	successText := fmt.Sprintf("✅ محصول '%s' با موفقیت حذف شد", productName)
	mb.sendMessage(chatID, successText)

	// Show remaining products
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📦 مشاهده لیست", fmt.Sprintf("product_list_%d", storeID)),
			tgbotapi.NewInlineKeyboardButtonData("🏠 بازگشت به پنل", fmt.Sprintf("manage_store_%d", storeID)),
		),
	)

	msg := tgbotapi.NewMessage(chatID, "چه کار می‌خواهید انجام دهید؟")
	msg.ReplyMarkup = keyboard
	mb.bot.Send(msg)
}

func (mb *MotherBot) toggleProductAvailability(chatID int64, user *models.User, productID uint) {
	product, err := mb.productService.GetProductByID(productID)
	if err != nil || product.Store.OwnerID != user.ID {
		mb.sendMessage(chatID, messages.ErrorGeneral)
		return
	}

	err = mb.productService.ToggleAvailability(productID)
	if err != nil {
		mb.sendMessage(chatID, messages.ErrorGeneral)
		return
	}

	newStatus := "فعال"
	if product.IsAvailable {
		newStatus = "غیرفعال"
	}

	successText := fmt.Sprintf("✅ وضعیت محصول '%s' به %s تغییر یافت", product.Name, newStatus)
	mb.sendMessage(chatID, successText)

	// Show updated product info
	mb.handleProductEdit(chatID, user, productID)
}