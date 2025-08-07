package main

import (
	"log"
	"net/http"
	"telegram-store-hub/internal/config"
	"telegram-store-hub/internal/database"
	"telegram-store-hub/server"

	"github.com/joho/godotenv"
)

func main() {
	// This is the web dashboard (optional)
	// The main bot runs from cmd/bot/main.go
	
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Run migrations
	if err := database.Migrate(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Start web dashboard (optional)
	log.Println("🌐 Starting web dashboard at http://localhost:5000")
	log.Println("⚠️  Note: The main bot should be started separately with: go run cmd/bot/main.go")
	
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
    <title>Telegram Store Hub Dashboard</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background: #f5f5f5; }
        .container { max-width: 800px; margin: 0 auto; background: white; padding: 40px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #2c3e50; text-align: center; }
        .status { padding: 20px; background: #e8f5e8; border-radius: 5px; margin: 20px 0; }
        .command { background: #f8f9fa; padding: 15px; border-radius: 5px; font-family: monospace; margin: 10px 0; }
        .note { background: #fff3cd; padding: 15px; border-radius: 5px; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🤖 Telegram Store Hub</h1>
        <div class="status">
            <h3>✅ Web Dashboard Active</h3>
            <p>This dashboard is running, but the main bot needs to be started separately.</p>
        </div>
        
        <div class="note">
            <h3>⚠️ Important</h3>
            <p>This web interface is optional. The main functionality is the Telegram bot.</p>
        </div>
        
        <h3>🚀 To Start the Bot:</h3>
        <div class="command">go run cmd/bot/main.go</div>
        
        <h3>📋 Or use the compiled binary:</h3>
        <div class="command">./telegram-store-hub</div>
        
        <h3>📁 Configuration:</h3>
        <p>Make sure your .env file is configured with:</p>
        <ul>
            <li>BOT_TOKEN (from @BotFather)</li>
            <li>DATABASE_URL (PostgreSQL connection)</li>
            <li>ADMIN_CHAT_ID (your Telegram chat ID)</li>
        </ul>
        
        <h3>📊 System Status:</h3>
        <p>Database: <span style="color: green;">Connected</span></p>
        <p>Bot Status: Check terminal output</p>
    </div>
</body>
</html>
		`))
	})

	http.ListenAndServe(":5000", mux)
}