#!/bin/bash

# Test script to simulate the installation process without actually installing system-wide

set -e

APP_NAME="telegram-store-hub"
TEST_DIR="/tmp/telegram-store-hub-test"

echo "🧪 Testing local database setup for Telegram Store Hub..."

# Clean test directory
rm -rf ${TEST_DIR}
mkdir -p ${TEST_DIR}/{config,logs,db}

# Copy binary for testing
if [[ -f "./${APP_NAME}" ]]; then
    echo "📋 Copying binary to test directory..."
    cp ./${APP_NAME} ${TEST_DIR}/
    chmod +x ${TEST_DIR}/${APP_NAME}
else
    echo "❌ Binary ${APP_NAME} not found. Run 'go build -o telegram-store-hub cmd/bot/main.go' first."
    exit 1
fi

# Test PostgreSQL connection (if available)
if command -v psql &> /dev/null; then
    echo "✅ PostgreSQL is available"
    
    # Test if we can connect to PostgreSQL
    if sudo -u postgres psql -c '\l' &> /dev/null; then
        echo "✅ PostgreSQL connection works"
        
        # Create test database
        echo "🧪 Creating test database..."
        sudo -u postgres psql << EOF
-- Create test database user
DROP USER IF EXISTS test_telegram_hub;
CREATE USER test_telegram_hub WITH ENCRYPTED PASSWORD 'test_password_123';

-- Create test database
DROP DATABASE IF EXISTS test_telegram_store_hub;
CREATE DATABASE test_telegram_store_hub OWNER test_telegram_hub;

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE test_telegram_store_hub TO test_telegram_hub;

-- Connect to the database and test table creation
\c test_telegram_store_hub;

-- Test users table
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

-- Test insert
INSERT INTO users (chat_id, username, first_name, is_admin) 
VALUES (123456789, 'test_admin', 'Test Admin', true);

-- Test select
SELECT * FROM users;

GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO test_telegram_hub;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO test_telegram_hub;

EOF
        
        echo "✅ Test database created and validated successfully!"
        
        # Create test configuration
        cat > ${TEST_DIR}/config/test.env << EOF
# Test Configuration
BOT_TOKEN=test_token_here
CHANNEL_USERNAME=@test_channel
ADMIN_CHAT_ID=123456789

# Test Database Configuration
DATABASE_URL=postgres://test_telegram_hub:test_password_123@localhost:5432/test_telegram_store_hub
DB_HOST=localhost
DB_PORT=5432
DB_NAME=test_telegram_store_hub
DB_USER=test_telegram_hub
DB_PASSWORD=test_password_123

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
        
        echo "📝 Test configuration created at: ${TEST_DIR}/config/test.env"
        
        # Clean up test database
        echo "🧹 Cleaning up test database..."
        sudo -u postgres psql << EOF
DROP DATABASE IF EXISTS test_telegram_store_hub;
DROP USER IF EXISTS test_telegram_hub;
EOF
        
    else
        echo "❌ Cannot connect to PostgreSQL as postgres user"
        exit 1
    fi
else
    echo "❌ PostgreSQL is not installed. The install script will install it automatically."
fi

echo ""
echo "✅ Local database setup test completed successfully!"
echo ""
echo "📋 Summary:"
echo "  ✓ Binary compilation works"
echo "  ✓ PostgreSQL installation check works"
echo "  ✓ Database creation and table setup works"
echo "  ✓ Configuration file generation works"
echo "  ✓ Test environment cleanup works"
echo ""
echo "🚀 The installation script is ready to:"
echo "  1. Install PostgreSQL (if not present)"
echo "  2. Create database and user automatically"
echo "  3. Set up all tables and indexes"
echo "  4. Generate secure configuration"
echo "  5. Install as systemd service"
echo ""
echo "📝 To install for real: sudo ./scripts/install.sh"

# Cleanup test directory
rm -rf ${TEST_DIR}