#!/bin/bash

# Telegram Store Hub - System Installation Script
# Installs the bot as a system service

set -e

APP_NAME="telegram-store-hub"
INSTALL_DIR="/usr/local/bin"
SERVICE_DIR="/etc/systemd/system"
CONFIG_DIR="/etc/telegram-store-hub"
LOG_DIR="/var/log/telegram-store-hub"

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   echo "❌ This script must be run as root (use sudo)" 
   exit 1
fi

echo "📥 Installing Telegram Store Hub..."

# Create directories
echo "📁 Creating directories..."
mkdir -p ${CONFIG_DIR}
mkdir -p ${LOG_DIR}
mkdir -p /var/lib/telegram-store-hub/db

# Copy binary
if [[ -f "./${APP_NAME}" ]]; then
    echo "📋 Installing binary to ${INSTALL_DIR}..."
    cp ./${APP_NAME} ${INSTALL_DIR}/
    chmod +x ${INSTALL_DIR}/${APP_NAME}
else
    echo "❌ Binary ${APP_NAME} not found. Run 'make build' first."
    exit 1
fi

# Copy configuration template
if [[ -f ".env.example" ]]; then
    echo "⚙️ Installing configuration template..."
    cp .env.example ${CONFIG_DIR}/telegram-store-hub.env.example
    
    if [[ ! -f "${CONFIG_DIR}/telegram-store-hub.env" ]]; then
        cp .env.example ${CONFIG_DIR}/telegram-store-hub.env
        echo "📝 Created configuration file: ${CONFIG_DIR}/telegram-store-hub.env"
        echo "⚠️  Please edit this file with your settings!"
    fi
fi

# Create systemd service
echo "🔧 Creating systemd service..."
cat > ${SERVICE_DIR}/${APP_NAME}.service << EOF
[Unit]
Description=Telegram Store Hub Bot
After=network.target postgresql.service
Wants=postgresql.service

[Service]
Type=simple
User=telegram-hub
Group=telegram-hub
WorkingDirectory=${CONFIG_DIR}
Environment="CONFIG_FILE=${CONFIG_DIR}/telegram-store-hub.env"
EnvironmentFile=${CONFIG_DIR}/telegram-store-hub.env
ExecStart=${INSTALL_DIR}/${APP_NAME}
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=${APP_NAME}

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=${LOG_DIR}

[Install]
WantedBy=multi-user.target
EOF

# Create user for the service
echo "👤 Creating system user..."
if ! id "telegram-hub" &>/dev/null; then
    useradd --system --no-create-home --shell /bin/false telegram-hub
fi

# Install PostgreSQL if not present
echo "🗄️ Setting up PostgreSQL database..."
if ! command -v psql &> /dev/null; then
    echo "📦 Installing PostgreSQL..."
    apt-get update
    apt-get install -y postgresql postgresql-contrib
fi

# Start PostgreSQL service
systemctl enable postgresql
systemctl start postgresql

# Generate secure password
DB_PASSWORD="secure_$(openssl rand -hex 12)"

# Create database and user
echo "🔧 Creating database and user..."
sudo -u postgres psql << EOF
-- Create database user
CREATE USER telegram_hub WITH ENCRYPTED PASSWORD '${DB_PASSWORD}';

-- Create database
CREATE DATABASE telegram_store_hub OWNER telegram_hub;

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE telegram_store_hub TO telegram_hub;

-- Connect to the database and create tables
\c telegram_store_hub;

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    chat_id BIGINT UNIQUE NOT NULL,
    username VARCHAR(255),
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    is_admin BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Stores table
CREATE TABLE IF NOT EXISTS stores (
    id SERIAL PRIMARY KEY,
    owner_chat_id BIGINT NOT NULL,
    store_name VARCHAR(255) NOT NULL,
    bot_token VARCHAR(255) UNIQUE,
    plan_type VARCHAR(50) DEFAULT 'free',
    is_active BOOLEAN DEFAULT true,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (owner_chat_id) REFERENCES users(chat_id)
);

-- Products table
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    store_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10,2),
    image_url VARCHAR(500),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (store_id) REFERENCES stores(id) ON DELETE CASCADE
);

-- Orders table
CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    store_id INTEGER NOT NULL,
    customer_chat_id BIGINT NOT NULL,
    customer_name VARCHAR(255),
    total_amount DECIMAL(10,2),
    status VARCHAR(50) DEFAULT 'pending',
    order_data JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (store_id) REFERENCES stores(id) ON DELETE CASCADE
);

-- Payments table
CREATE TABLE IF NOT EXISTS payments (
    id SERIAL PRIMARY KEY,
    store_id INTEGER,
    plan_type VARCHAR(50),
    amount DECIMAL(10,2),
    status VARCHAR(50) DEFAULT 'pending',
    payment_proof VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (store_id) REFERENCES stores(id)
);

-- Subscriptions table
CREATE TABLE IF NOT EXISTS subscriptions (
    id SERIAL PRIMARY KEY,
    store_id INTEGER NOT NULL,
    plan_type VARCHAR(50) NOT NULL,
    starts_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (store_id) REFERENCES stores(id) ON DELETE CASCADE
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_users_chat_id ON users(chat_id);
CREATE INDEX IF NOT EXISTS idx_stores_owner_chat_id ON stores(owner_chat_id);
CREATE INDEX IF NOT EXISTS idx_products_store_id ON products(store_id);
CREATE INDEX IF NOT EXISTS idx_orders_store_id ON orders(store_id);
CREATE INDEX IF NOT EXISTS idx_orders_customer_chat_id ON orders(customer_chat_id);
CREATE INDEX IF NOT EXISTS idx_payments_store_id ON payments(store_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_store_id ON subscriptions(store_id);

-- Insert default admin user (you'll need to update this with your actual chat ID)
INSERT INTO users (chat_id, username, first_name, is_admin) 
VALUES (123456789, 'admin', 'System Admin', true) 
ON CONFLICT (chat_id) DO NOTHING;

GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO telegram_hub;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO telegram_hub;

EOF

# Update configuration with database connection
echo "📝 Updating configuration with database connection..."
cat > ${CONFIG_DIR}/telegram-store-hub.env << EOF
# Telegram Bot Configuration
BOT_TOKEN=your_bot_token_here
CHANNEL_USERNAME=@your_channel
ADMIN_CHAT_ID=123456789

# Database Configuration (Auto-generated during installation)
DATABASE_URL=postgres://telegram_hub:${DB_PASSWORD}@localhost:5432/telegram_store_hub
DB_HOST=localhost
DB_PORT=5432
DB_NAME=telegram_store_hub
DB_USER=telegram_hub
DB_PASSWORD=${DB_PASSWORD}

# Application Configuration
LOG_LEVEL=info
MAX_PRODUCTS_FREE=10
MAX_PRODUCTS_PRO=200
MAX_PRODUCTS_VIP=999999

# Payment Configuration
PAYMENT_CARD_NUMBER=1234-5678-9012-3456
PAYMENT_CARD_NAME=CodeRoot Store
PRO_PLAN_PRICE=50000
VIP_PLAN_PRICE=150000
EOF

# Set permissions
echo "🔒 Setting permissions..."
chown -R telegram-hub:telegram-hub ${CONFIG_DIR}
chown -R telegram-hub:telegram-hub ${LOG_DIR}
chown -R telegram-hub:telegram-hub /var/lib/telegram-store-hub
chmod 600 ${CONFIG_DIR}/telegram-store-hub.env*
chmod 755 ${INSTALL_DIR}/${APP_NAME}

# Reload systemd
echo "🔄 Reloading systemd..."
systemctl daemon-reload

echo "✅ Installation completed successfully!"
echo "🗄️ PostgreSQL database created and configured automatically!"
echo ""
echo "📋 Next steps:"
echo "  1. Edit configuration: sudo nano ${CONFIG_DIR}/telegram-store-hub.env"
echo "  2. Add your Telegram bot token (get from @BotFather)"
echo "  3. Update ADMIN_CHAT_ID with your Telegram chat ID"
echo "  4. Update CHANNEL_USERNAME with your channel"
echo "  5. Start the service: sudo systemctl start ${APP_NAME}"
echo "  6. Enable auto-start: sudo systemctl enable ${APP_NAME}"
echo ""
echo "🔍 Useful commands:"
echo "  Status: sudo systemctl status ${APP_NAME}"
echo "  Logs: sudo journalctl -u ${APP_NAME} -f"
echo "  Stop: sudo systemctl stop ${APP_NAME}"
echo "  Restart: sudo systemctl restart ${APP_NAME}"
echo ""
echo "📁 Files installed:"
echo "  Binary: ${INSTALL_DIR}/${APP_NAME}"
echo "  Config: ${CONFIG_DIR}/telegram-store-hub.env"
echo "  Service: ${SERVICE_DIR}/${APP_NAME}.service"
echo "  Logs: ${LOG_DIR}/"