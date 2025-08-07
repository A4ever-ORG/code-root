package main

import (
        "log"
        "telegram-store-hub/internal/bot"
        "telegram-store-hub/internal/config"
        "telegram-store-hub/internal/database"
        "telegram-store-hub/internal/services"

        tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
        log.Println("🚀 Starting Telegram Store Hub...")

        // Load configuration
        cfg, err := config.LoadConfig()
        if err != nil {
                log.Fatalf("❌ Failed to load config: %v", err)
        }

        // Initialize database
        db, err := database.Connect(cfg.DatabaseURL)
        if err != nil {
                log.Fatalf("❌ Database connection failed: %v", err)
        }
        log.Println("✅ Database connected successfully")

        // Run migrations
        if err := database.Migrate(db); err != nil {
                log.Fatalf("❌ Database migration failed: %v", err)
        }
        log.Println("✅ Database migrated successfully")

        // Seed default data
        if err := database.SeedDefaultData(db, 1); err != nil {
                log.Printf("⚠️ Seeding warning: %v", err)
        }
        log.Println("✅ Database seeded successfully")

        // Initialize services
        log.Println("🔧 Initializing services...")
        
        userService := services.NewUserService(db)
        sessionService := services.NewSessionService(db)
        storeManager := services.NewStoreManagerService(db)
        productService := services.NewProductService(db)
        orderService := services.NewOrderService(db)
        paymentService := services.NewPaymentService(db)
        subscriptionService := services.NewSubscriptionService(db)
        botManager := services.NewBotManagerService(db)
        
        log.Println("✅ Services initialized")

        // Initialize mother bot
        motherBot, err := tgbotapi.NewBotAPI(cfg.MotherBotToken)
        if err != nil {
                log.Fatalf("❌ Failed to create mother bot: %v", err)
        }
        
        motherBot.Debug = cfg.Debug
        log.Printf("✅ Mother bot authorized as @%s", motherBot.Self.UserName)

        // Initialize channel verification with bot
        channelVerify := services.NewChannelVerificationService(motherBot, cfg.RequiredChannelID)

        // Set bot for subscription service notifications
        subscriptionService.SetBot(motherBot)

        // Initialize bot manager
        mb := bot.NewMotherBot(
                motherBot,
                db,
                channelVerify,
                userService,
                sessionService,
                storeManager,
                productService,
                orderService,
                paymentService,
                subscriptionService,
                botManager,
        )

        // Start all store bots
        log.Println("🤖 Starting store bots...")
        if err := botManager.StartAllBots(); err != nil {
                log.Printf("⚠️ Error starting store bots: %v", err)
        }

        // Start subscription checker
        log.Println("⏰ Starting subscription checker...")
        subscriptionService.StartSubscriptionChecker()

        // Start mother bot
        log.Println("🤖 Starting mother bot...")
        mb.Start()
}