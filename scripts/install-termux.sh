#!/bin/bash

# Telegram Store Hub - Termux (Android) Installation Script
# Specialized installation for Android Termux environment

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}📱 Telegram Store Hub - Termux Installation${NC}"
echo "============================================="

# Check if running in Termux
if [[ -z "$PREFIX" || "$PREFIX" != "/data/data/com.termux"* ]]; then
    echo -e "${RED}❌ This script is only for Termux (Android)${NC}"
    echo -e "${YELLOW}Please use the main setup script: ./scripts/setup.sh${NC}"
    exit 1
fi

# Update Termux packages
echo -e "${BLUE}📦 Updating Termux packages...${NC}"
pkg update -y

# Install basic tools
echo -e "${BLUE}🔧 Installing basic tools...${NC}"
pkg install -y git curl wget make clang

# Install Go
echo -e "${BLUE}🔧 Installing Go...${NC}"
pkg install -y golang

# Install PostgreSQL
echo -e "${BLUE}🗄️ Installing PostgreSQL...${NC}"
pkg install -y postgresql

# Initialize PostgreSQL
echo -e "${BLUE}🔧 Initializing PostgreSQL...${NC}"
if [[ ! -d "$PREFIX/var/lib/postgresql" ]]; then
    initdb $PREFIX/var/lib/postgresql
fi

# Start PostgreSQL
echo -e "${BLUE}🚀 Starting PostgreSQL...${NC}"
pg_ctl -D $PREFIX/var/lib/postgresql -l $PREFIX/var/lib/postgresql/logfile start

# Wait for PostgreSQL to start
sleep 3

# Create database
echo -e "${BLUE}🗄️ Setting up database...${NC}"
DB_PASSWORD="tsh_$(openssl rand -hex 16)"
DATABASE_NAME="telegram_store_hub"
DATABASE_USER="telegram_hub"

createdb $DATABASE_NAME
psql -d $DATABASE_NAME -c "CREATE USER $DATABASE_USER WITH ENCRYPTED PASSWORD '$DB_PASSWORD';"
psql -d $DATABASE_NAME -c "GRANT ALL PRIVILEGES ON DATABASE $DATABASE_NAME TO $DATABASE_USER;"

# Create tables (simplified for Termux)
psql -d $DATABASE_NAME << EOF
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    chat_id BIGINT UNIQUE NOT NULL,
    username VARCHAR(255),
    first_name VARCHAR(255),
    is_admin BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS stores (
    id SERIAL PRIMARY KEY,
    owner_chat_id BIGINT NOT NULL,
    store_name VARCHAR(255) NOT NULL,
    bot_token VARCHAR(255) UNIQUE,
    plan_type VARCHAR(50) DEFAULT 'free',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    store_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    price DECIMAL(10,2) DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO $DATABASE_USER;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO $DATABASE_USER;
EOF

# Create configuration
echo -e "${BLUE}⚙️ Creating configuration...${NC}"
cat > .env << EOF
# Telegram Bot Configuration
BOT_TOKEN=your_bot_token_here
ADMIN_CHAT_ID=123456789

# Database Configuration (Termux)
DATABASE_URL=postgres://${DATABASE_USER}:${DB_PASSWORD}@localhost:5432/${DATABASE_NAME}?sslmode=disable
DB_HOST=localhost
DB_PORT=5432
DB_NAME=${DATABASE_NAME}
DB_USER=${DATABASE_USER}
DB_PASSWORD=${DB_PASSWORD}

# App Configuration
LOG_LEVEL=info
MAX_PRODUCTS_FREE=10
MAX_PRODUCTS_PRO=200
MAX_PRODUCTS_VIP=999999
EOF

# Save database password
echo "$DB_PASSWORD" > .db_password
chmod 600 .db_password

# Download Go modules and build
echo -e "${BLUE}🔧 Building application...${NC}"
if [[ ! -f "go.mod" ]]; then
    go mod init telegram-store-hub
    go mod edit -require github.com/go-telegram-bot-api/telegram-bot-api/v5@latest
    go mod edit -require github.com/joho/godotenv@latest
    go mod edit -require gorm.io/driver/postgres@latest
    go mod edit -require gorm.io/gorm@latest
fi

go mod download
go mod tidy

mkdir -p bin
go build -o bin/telegram-store-hub .
chmod +x bin/telegram-store-hub

# Create run script for Termux
echo -e "${BLUE}📝 Creating run script...${NC}"
cat > run.sh << 'EOF'
#!/bin/bash

# Start PostgreSQL if not running
if ! pgrep -x "postgres" > /dev/null; then
    echo "🗄️ Starting PostgreSQL..."
    pg_ctl -D $PREFIX/var/lib/postgresql -l $PREFIX/var/lib/postgresql/logfile start
    sleep 2
fi

# Check configuration
if [[ "$BOT_TOKEN" == "your_bot_token_here" ]] || [[ -z "$BOT_TOKEN" ]]; then
    echo "❌ Bot token not configured!"
    echo "Please edit .env file and add your bot token from @BotFather"
    exit 1
fi

# Load environment
set -a
source .env
set +a

echo "🚀 Starting Telegram Store Hub..."
./bin/telegram-store-hub
EOF

chmod +x run.sh

# Create auto-start script
echo -e "${BLUE}🔧 Creating auto-start script...${NC}"
cat > start-postgres.sh << EOF
#!/bin/bash
# Auto-start PostgreSQL for Termux

if ! pgrep -x "postgres" > /dev/null; then
    echo "🗄️ Starting PostgreSQL..."
    pg_ctl -D \$PREFIX/var/lib/postgresql -l \$PREFIX/var/lib/postgresql/logfile start
else
    echo "✅ PostgreSQL already running"
fi
EOF

chmod +x start-postgres.sh

echo ""
echo -e "${GREEN}🎉 Termux installation completed successfully!${NC}"
echo "============================================"
echo ""
echo -e "${YELLOW}📋 Next steps:${NC}"
echo "  1. Get bot token from @BotFather"
echo "  2. Edit configuration: nano .env"
echo "  3. Set your BOT_TOKEN and ADMIN_CHAT_ID"
echo "  4. Run the bot: ./run.sh"
echo ""
echo -e "${YELLOW}📱 Termux-specific notes:${NC}"
echo "  - PostgreSQL must be running before starting the bot"
echo "  - Use './start-postgres.sh' to start PostgreSQL"
echo "  - The app will run in foreground - use screen/tmux for background"
echo "  - Database password saved in .db_password"
echo ""
echo -e "${YELLOW}🔧 Useful commands:${NC}"
echo "  Start PostgreSQL: ./start-postgres.sh"
echo "  Check PostgreSQL: pgrep postgres"
echo "  Run bot: ./run.sh"
echo "  View database: psql $DATABASE_NAME"
echo ""
echo -e "${GREEN}✅ Your Telegram Store Hub is ready for Termux!${NC}"