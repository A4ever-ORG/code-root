package main

import (
        "log"
        "os"
        "os/signal"
        "syscall"
        "telegram-store-hub/internal/bot"
        "telegram-store-hub/internal/config"
        "telegram-store-hub/internal/database"
        "telegram-store-hub/internal/services"

        "github.com/joho/godotenv"
)

func main() {
        // Load environment variables from .env file
        if err := godotenv.Load(); err != nil {
                log.Println("⚠️  No .env file found, using system environment variables")
        }

        // Initialize configuration
        cfg := config.Load()

        // Validate required configuration
        if cfg.BotToken == "" {
                log.Fatal("❌ BOT_TOKEN is required. Please set it in .env file or environment variables")
        }

        if cfg.DatabaseURL == "" {
                log.Fatal("❌ DATABASE_URL is required. Please set it in .env file or environment variables")
        }

        log.Println("🚀 Starting Telegram Store Hub System...")
        log.Printf("📋 Environment: %s", cfg.Environment)

        // Initialize database
        log.Println("🔗 Connecting to database...")
        db, err := database.Connect(cfg.DatabaseURL)
        if err != nil {
                log.Fatal("❌ Failed to connect to database:", err)
        }
        log.Println("✅ Database connected successfully")

        // Run migrations
        log.Println("🔄 Running database migrations...")
        if err := database.Migrate(db); err != nil {
                log.Fatal("❌ Failed to run migrations:", err)
        }
        log.Println("✅ Database migrations completed")

        // Seed default data (admin user)
        if err := database.SeedDefaultData(db, cfg.AdminChatID); err != nil {
                log.Printf("⚠️  Warning: Failed to seed default data: %v", err)
        }

        // Initialize services
        botManager := services.NewBotManagerService(db)

        // Initialize and start the mother bot (CodeRoot)
        log.Println("🤖 Starting CodeRoot Mother Bot...")
        motherBot, err := bot.NewMotherBot(cfg.BotToken, db)
        if err != nil {
                log.Fatal("❌ Failed to create mother bot:", err)
        }

        // Start all existing sub-bots
        log.Println("🔄 Starting existing store bots...")
        if err := botManager.StartAllBots(); err != nil {
                log.Printf("⚠️  Warning: Failed to start some store bots: %v", err)
        }

        log.Println("✅ CodeRoot Mother Bot started successfully!")
        log.Println("📞 Bot is ready to receive messages")
        
        if cfg.ChannelID != "" {
                log.Printf("📢 Forced join channel: %s", cfg.ChannelID)
        }

        if cfg.AdminChatID != 0 {
                log.Printf("👨‍💼 Admin chat ID: %d", cfg.AdminChatID)
        }

        // Set up graceful shutdown
        c := make(chan os.Signal, 1)
        signal.Notify(c, os.Interrupt, syscall.SIGTERM)

        go func() {
                <-c
                log.Println("\n🛑 Shutting down Telegram Store Hub...")
                
                // Stop all bots
                log.Println("⏹ Stopping all store bots...")
                botManager.StopAllBots()
                
                log.Println("👋 Telegram Store Hub stopped successfully")
                os.Exit(0)
        }()

        // Start the mother bot (blocking call)
        motherBot.Start()
}