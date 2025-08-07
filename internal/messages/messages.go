package messages

// Persian messages for the bot
const (
        // Welcome and main menu
        WelcomeMessage = `🤖 سلام! به ربات مادر CodeRoot خوش آمدید
        
🏪 این ربات به شما کمک می‌کند تا فروشگاه تلگرامی خودتان را راه‌اندازی کنید

📋 منوی اصلی:`

        MainMenuKeyboard = `منوی اصلی`
        
        // Main menu buttons
        RegisterStoreBtn    = "🏪 ثبت فروشگاه جدید"
        MyStoresBtn        = "📊 فروشگاه‌های من"
        SupportBtn         = "🆘 پشتیبانی"
        AboutBtn           = "ℹ️ درباره ما"
        
        // Store registration
        StoreRegistrationStart = `🏪 ثبت فروشگاه جدید

لطفاً نام فروشگاه خود را وارد کنید:`

        StoreNameReceived = `✅ نام فروشگاه: %s

حالا توضیح کوتاهی از فروشگاه خود بنویسید:`

        StoreDescriptionReceived = `✅ توضیحات فروشگاه ثبت شد

حالا یکی از پلن‌های زیر را انتخاب کنید:`

        // Plans
        FreePlanBtn  = "🆓 رایگان (۰ تومان)"
        ProPlanBtn   = "💎 حرفه‌ای (۵۰,۰۰۰ تومان)"
        VIPPlanBtn   = "👑 VIP (۱۵۰,۰۰۰ تومان)"
        
        FreePlanDetails = `🆓 پلن رایگان:
• حداکثر ۱۰ محصول
• دکمه‌های ثابت
• ۵٪ کارمزد از فروش
• پشتیبانی پایه`

        ProPlanDetails = `💎 پلن حرفه‌ای:
• تا ۲۰۰ محصول
• گزارش‌های پیشرفته
• پیام خوش‌آمدگویی
• تبلیغات دلخواه
• ۵٪ کارمزد از فروش
• پشتیبانی اولویت‌دار`

        VIPPlanDetails = `👑 پلن VIP:
• محصولات نامحدود
• درگاه پرداخت اختصاصی
• بدون کارمزد
• تبلیغات ویژه
• دکمه‌های شخصی‌سازی شده
• پشتیبانی ۲۴/۷`

        // Payment
        PaymentInstructions = `💳 دستورالعمل پرداخت

پلن انتخابی: %s
مبلغ: %s تومان

📋 شماره کارت:
%s

👤 به نام: %s

پس از پرداخت، عکس رسید را ارسال کنید.`

        PaymentReceived = `✅ رسید پرداخت دریافت شد

🔄 در حال بررسی توسط ادمین...
📞 در صورت تایید، فروشگاه شما فعال خواهد شد`

        PaymentApproved = `🎉 تبریک! پرداخت شما تایید شد

🤖 ربات فروشگاه شما آماده است:
@%s

🔑 توکن ربات: %s
📊 پنل مدیریت: /panel`

        // Force join
        ForceJoinMessage = `⚠️ برای استفاده از ربات باید عضو کانال ما باشید

🔗 کانال: %s

پس از عضویت دوباره امتحان کنید.`

        CheckMembershipBtn = "✅ عضو شدم"

        // Store panel
        StorePanelMessage = `📊 پنل مدیریت فروشگاه

فروشگاه: %s
پلن: %s
تاریخ انقضا: %s
محصولات: %d/%d`

        AddProductBtn     = "➕ افزودن محصول"
        ProductListBtn    = "📦 لیست محصولات"
        OrdersBtn         = "🛒 سفارش‌ها"
        SalesReportBtn    = "📈 گزارش فروش"
        RenewPlanBtn      = "🔄 تمدید پلن"
        StoreSettingsBtn  = "⚙️ تنظیمات فروشگاه"

        // Product management
        AddProductStart = `➕ افزودن محصول جدید

نام محصول را وارد کنید:`

        ProductNameReceived = `✅ نام محصول: %s

توضیحات محصول را وارد کنید:`

        ProductDescReceived = `✅ توضیحات ثبت شد

قیمت محصول را به تومان وارد کنید:`

        ProductPriceReceived = `✅ قیمت: %s تومان

عکس محصول را ارسال کنید (اختیاری):
یا /skip را بزنید`

        ProductAdded = `✅ محصول با موفقیت اضافه شد

نام: %s
قیمت: %s تومان
وضعیت: فعال`

        // Support
        SupportMessage = `🆘 بخش پشتیبانی

چگونه می‌توانیم کمکتان کنیم؟`

        FAQBtn           = "❓ سوالات متداول"
        ContactUsBtn     = "📞 تماس با ما"
        TelegramSupportBtn = "💬 پشتیبانی تلگرام"

        // FAQ
        FAQMessage = `❓ سوالات متداول

🔹 چگونه فروشگاه بسازم؟
از دکمه "ثبت فروشگاه جدید" استفاده کنید

🔹 تفاوت پلن‌ها چیست؟
پلن رایگان: ۱۰ محصول، ۵٪ کارمزد
پلن حرفه‌ای: ۲۰۰ محصول، گزارش‌های پیشرفته
پلن VIP: نامحدود، بدون کارمزد

🔹 چگونه محصول اضافه کنم؟
از پنل فروشگاه > افزودن محصول

🔹 کارمزد چگونه محاسبه می‌شود؟
از هر فروش درصدی کسر می‌شود`

        // Contact and About
        ContactMessage = `📞 تماس با ما
        
📧 ایمیل: support@coderoot.ir
📱 تلگرام: @CodeRootSupport
📞 تلفن: ۰۲۱-۱۲۳۴۵۶۷۸

🕐 ساعات کاری: شنبه تا پنج‌شنبه ۹ تا ۱۸`
        
        AboutMessage = `ℹ️ درباره CodeRoot

🤖 سیستم هوشمند ساخت ربات‌های فروشگاهی

🌟 ویژگی‌ها:
✅ ساخت ربات در کمتر از ۱ دقیقه
✅ پنل مدیریت پیشرفته
✅ سیستم پرداخت امن
✅ پشتیبانی ۲۴ ساعته
✅ بروزرسانی رایگان

📊 آمار:
👥 بیش از ۱۰۰۰ کاربر راضی
🏪 بیش از ۵۰۰ فروشگاه فعال
💰 بیش از ۱ میلیارد تومان گردش مالی

💎 نسخه ۱.۰.۰`

        // Errors
        ErrorGeneral     = "❌ خطایی رخ داد. لطفاً دوباره تلاش کنید"
        ErrorInvalidInput = "❌ ورودی نامعتبر. لطفاً دوباره امتحان کنید"
        ErrorNoStore     = "❌ شما هیچ فروشگاهی ندارید"
        ErrorStoreLimit  = "❌ محدودیت تعداد محصولات"
        ErrorExpiredPlan = "❌ پلن شما منقضی شده. لطفاً تمدید کنید"

        // Admin messages
        AdminNewStoreNotification = `🏪 فروشگاه جدید ثبت شد

👤 کاربر: %s (@%s)
🏪 فروشگاه: %s
💰 پلن: %s
💳 مبلغ: %s تومان

تایید یا رد کنید:`

        AdminApproveBtn = "✅ تایید"
        AdminRejectBtn  = "❌ رد"

        // Success messages
        SuccessStoreCreated = "🎉 فروشگاه با موفقیت ایجاد شد!"
        SuccessPaymentConfirmed = "✅ پرداخت تایید شد"
        SuccessProductAdded = "✅ محصول اضافه شد"
)