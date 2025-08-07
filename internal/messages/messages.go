package messages

// Persian messages for the Telegram bot interface

const (
	// Welcome and main menu messages
	WelcomeMessage = `🎉 به ربات CodeRoot خوش آمدید!

🤖 این ربات به شما کمک می‌کند تا فروشگاه تلگرامی خود را راه‌اندازی کنید.

📋 امکانات:
• ایجاد ربات فروشگاهی اختصاصی
• مدیریت محصولات و سفارش‌ها
• سیستم پرداخت
• پشتیبانی ۲۴ ساعته

برای شروع روی دکمه "ثبت فروشگاه" کلیک کنید.`

	MainMenuKeyboard = "منوی اصلی"

	// Store registration messages
	StoreRegistrationStart = "📝 برای ثبت فروشگاه، ابتدا یکی از پلن‌های زیر را انتخاب کنید:"

	// Error messages
	ErrorNoStore        = "❌ شما هنوز فروشگاهی ثبت نکرده‌اید!"
	ErrorStoreExists    = "⚠️ شما قبلاً یک فروشگاه ثبت کرده‌اید!"
	ErrorCreateStore    = "❌ خطا در ثبت فروشگاه. لطفاً دوباره تلاش کنید."
	ErrorNoPermission   = "❌ شما مجوز دسترسی به این بخش را ندارید."
	ErrorInvalidCommand = "❌ دستور نامعتبر. لطفاً از منوی اصلی استفاده کنید."
	ErrorDatabaseError  = "❌ خطا در اتصال به پایگاه داده. لطفاً بعداً تلاش کنید."

	// Success messages
	SuccessStoreCreated = `🎉 تبریک! فروشگاه شما با موفقیت ثبت شد!

📋 اطلاعات فروشگاه:
• نوع پلن: رایگان
• تعداد محصولات مجاز: 10
• مدت اعتبار: 1 ماه

🤖 ربات فروشگاهی شما تا 24 ساعت آینده آماده خواهد شد.

برای مدیریت فروشگاه از دکمه "پنل مدیریت" استفاده کنید.`

	// Channel membership messages
	ChannelJoinRequired = `⚠️ برای استفاده از ربات باید عضو کانال ما باشید:

👆 ابتدا در کانال عضو شوید، سپس روی دکمه "تایید عضویت" کلیک کنید.`

	ChannelJoinSuccess = "✅ عضویت شما تایید شد! حالا می‌توانید از ربات استفاده کنید."

	// Admin messages
	AdminPanelWelcome = `👑 پنل مدیریت

به پنل مدیریت خوش آمدید. از اینجا می‌توانید سیستم را مدیریت کنید.`

	// Plan descriptions
	FreePlanDescription = `💚 پلن رایگان
• تا 10 محصول
• کارمزد 5٪
• پشتیبانی عمومی
• مدت اعتبار: 1 ماه`

	ProPlanDescription = `💎 پلن حرفه‌ای - 50,000 تومان/ماه
• تا 200 محصول
• کارمزد 5٪
• پشتیبانی اولویت‌دار
• آمار فروش تفصیلی
• درگاه پرداخت`

	VIPPlanDescription = `👑 پلن VIP - 150,000 تومان/ماه
• محصولات نامحدود
• درگاه پرداخت اختصاصی
• بدون کارمزد
• تبلیغات ویژه
• شخصی‌سازی کامل`

	// Store management messages
	StoreManagementMenu = "📊 پنل مدیریت فروشگاه"
	AddProductMenu      = "➕ افزودن محصول"
	ListProductsMenu    = "📦 لیست محصولات"
	ViewOrdersMenu      = "🛒 مشاهده سفارش‌ها"
	SalesReportMenu     = "📈 گزارش فروش"
	StoreSettingsMenu   = "⚙️ تنظیمات فروشگاه"
	RenewPlanMenu       = "🔄 تمدید/ارتقاء پلن"

	// Payment messages
	PaymentInstructions = `💳 راهنمای پرداخت

پس از انتقال وجه، اسکرین‌شات رسید را ارسال کنید.

⚠️ توجه: پس از تایید پرداخت توسط ادمین، ربات شما فعال خواهد شد.`

	PaymentCardInfo = `🏦 اطلاعات پرداخت:
شماره کارت: %s
به نام: %s

مبلغ قابل پرداخت: %s تومان`

	// User states for conversation flow
	StateWaitingStoreName        = "waiting_store_name"
	StateWaitingStoreDescription = "waiting_store_description"
	StateWaitingPaymentProof     = "waiting_payment_proof"
	StateWaitingProductName      = "waiting_product_name"
	StateWaitingProductPrice     = "waiting_product_price"
	StateWaitingProductImage     = "waiting_product_image"

	// Button texts
	ButtonRegisterStore    = "🏪 ثبت فروشگاه"
	ButtonManageStore      = "📊 پنل مدیریت"
	ButtonViewPlans        = "💎 مشاهده پلن‌ها"
	ButtonSupport          = "🆘 پشتیبانی"
	ButtonBack             = "🔙 بازگشت"
	ButtonMainMenu         = "🏠 منوی اصلی"
	ButtonCheckMembership  = "✅ تایید عضویت"
	ButtonJoinChannel      = "📢 عضویت در کانال"
	ButtonFreePlan         = "💚 پلن رایگان"
	ButtonProPlan          = "💎 پلن حرفه‌ای"
	ButtonVIPPlan          = "👑 پلن VIP"
	ButtonPaymentComplete  = "✅ پرداخت کردم"
	ButtonCancel           = "❌ انصراف"

	// Help and support messages
	SupportMessage = `🆘 پشتیبانی

برای دریافت کمک می‌توانید:

📧 با ما تماس بگیرید: @support
📞 شماره پشتیبانی: +98912345678
🕐 ساعات کاری: شنبه تا پنج‌شنبه 9 تا 18

❓ سوالات متداول:
• چگونه محصول اضافه کنم؟
• نحوه تغییر تنظیمات ربات
• راهنمای پرداخت و تسویه`

	// Statistics messages
	StatsMessage = `📊 آمار سیستم

👥 تعداد کاربران: %d
🏪 تعداد فروشگاه‌ها: %d
📦 تعداد محصولات: %d
🛒 تعداد سفارش‌ها: %d
💰 مجموع فروش: %s تومان`

	// Common formatting
	Separator     = "\n────────────\n"
	BoldStart     = "*"
	BoldEnd       = "*"
	ItalicStart   = "_"
	ItalicEnd     = "_"
	CodeStart     = "`"
	CodeEnd       = "`"
	PreStart      = "```"
	PreEnd        = "```"

	// Time formats
	DateFormat     = "2006/01/02"
	TimeFormat     = "15:04"
	DateTimeFormat = "2006/01/02 15:04"

	// Currency
	CurrencySymbol = "تومان"
	PriceFormat    = "%s %s" // price + currency
)