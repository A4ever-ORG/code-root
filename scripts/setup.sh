#!/bin/bash

# Telegram Store Hub - Complete Setup Script
# Installs all prerequisites, sets up the environment, and builds the application
# Supports Ubuntu/Debian, CentOS/RHEL, macOS, and Android Termux

set -e

APP_NAME="telegram-store-hub"
VERSION="1.0.0"
REQUIRED_GO_VERSION="1.22"
DATABASE_NAME="telegram_store_hub"
DATABASE_USER="telegram_hub"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Logging functions
log_info() { echo -e "${BLUE}ℹ️  $1${NC}"; }
log_success() { echo -e "${GREEN}✅ $1${NC}"; }
log_warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }
log_error() { echo -e "${RED}❌ $1${NC}"; }
log_step() { echo -e "${PURPLE}🔧 $1${NC}"; }

# Detect OS and distribution
detect_os() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        if [ -f /etc/os-release ]; then
            . /etc/os-release
            OS=$ID
            VER=$VERSION_ID
        elif [ -f /etc/redhat-release ]; then
            OS="centos"
        elif [ -f /etc/debian_version ]; then
            OS="debian"
        else
            OS="unknown"
        fi
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        OS="macos"
    elif [[ "$OSTYPE" == "cygwin" ]]; then
        OS="cygwin"
    elif [[ "$OSTYPE" == "msys" ]]; then
        OS="msys"
    else
        OS="unknown"
    fi
    
    # Check for Termux (Android)
    if [[ -n "$PREFIX" && "$PREFIX" == "/data/data/com.termux"* ]]; then
        OS="termux"
    fi
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Install PostgreSQL based on OS
install_postgresql() {
    log_step "Installing PostgreSQL..."
    
    case $OS in
        "ubuntu"|"debian")
            if ! command_exists psql; then
                sudo apt-get update
                sudo apt-get install -y postgresql postgresql-contrib postgresql-client
                sudo systemctl enable postgresql
                sudo systemctl start postgresql
            else
                log_success "PostgreSQL already installed"
            fi
            ;;
        "centos"|"rhel"|"fedora")
            if ! command_exists psql; then
                if command_exists dnf; then
                    sudo dnf install -y postgresql postgresql-server postgresql-contrib
                else
                    sudo yum install -y postgresql postgresql-server postgresql-contrib
                fi
                sudo postgresql-setup initdb
                sudo systemctl enable postgresql
                sudo systemctl start postgresql
            else
                log_success "PostgreSQL already installed"
            fi
            ;;
        "macos")
            if ! command_exists psql; then
                if command_exists brew; then
                    brew install postgresql@15
                    brew services start postgresql@15
                else
                    log_error "Homebrew not found. Please install Homebrew first: https://brew.sh"
                    exit 1
                fi
            else
                log_success "PostgreSQL already installed"
            fi
            ;;
        "termux")
            if ! command_exists psql; then
                pkg update
                pkg install -y postgresql
                # Initialize PostgreSQL in Termux
                initdb $PREFIX/var/lib/postgresql
                pg_ctl -D $PREFIX/var/lib/postgresql -l $PREFIX/var/lib/postgresql/logfile start
            else
                log_success "PostgreSQL already installed"
            fi
            ;;
        *)
            log_error "Unsupported OS for automatic PostgreSQL installation: $OS"
            log_info "Please install PostgreSQL manually and run this script again"
            exit 1
            ;;
    esac
}

# Install Go based on OS
install_go() {
    log_step "Installing Go programming language..."
    
    if command_exists go; then
        CURRENT_GO_VERSION=$(go version | cut -d' ' -f3 | sed 's/go//')
        if [[ "$CURRENT_GO_VERSION" == "$REQUIRED_GO_VERSION"* ]]; then
            log_success "Go $CURRENT_GO_VERSION already installed"
            return
        else
            log_warning "Go $CURRENT_GO_VERSION found, but $REQUIRED_GO_VERSION is required"
        fi
    fi
    
    case $OS in
        "ubuntu"|"debian")
            # Install from official Go repository
            sudo apt-get update
            sudo apt-get install -y wget
            wget -q https://golang.org/dl/go${REQUIRED_GO_VERSION}.linux-amd64.tar.gz
            sudo rm -rf /usr/local/go
            sudo tar -C /usr/local -xzf go${REQUIRED_GO_VERSION}.linux-amd64.tar.gz
            rm go${REQUIRED_GO_VERSION}.linux-amd64.tar.gz
            ;;
        "centos"|"rhel"|"fedora")
            if command_exists dnf; then
                sudo dnf install -y wget
            else
                sudo yum install -y wget
            fi
            wget -q https://golang.org/dl/go${REQUIRED_GO_VERSION}.linux-amd64.tar.gz
            sudo rm -rf /usr/local/go
            sudo tar -C /usr/local -xzf go${REQUIRED_GO_VERSION}.linux-amd64.tar.gz
            rm go${REQUIRED_GO_VERSION}.linux-amd64.tar.gz
            ;;
        "macos")
            if command_exists brew; then
                brew install go
            else
                # Download and install Go manually
                curl -L https://golang.org/dl/go${REQUIRED_GO_VERSION}.darwin-amd64.tar.gz -o go.tar.gz
                sudo rm -rf /usr/local/go
                sudo tar -C /usr/local -xzf go.tar.gz
                rm go.tar.gz
            fi
            ;;
        "termux")
            pkg install -y golang
            ;;
        *)
            log_error "Unsupported OS for automatic Go installation: $OS"
            log_info "Please install Go manually from https://golang.org/dl/"
            exit 1
            ;;
    esac
    
    # Add Go to PATH
    if [[ ":$PATH:" != *":/usr/local/go/bin:"* ]]; then
        echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
        echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.profile
        export PATH=$PATH:/usr/local/go/bin
    fi
    
    # Set GOPATH
    if [[ -z "$GOPATH" ]]; then
        mkdir -p ~/go
        echo 'export GOPATH=$HOME/go' >> ~/.bashrc
        echo 'export GOPATH=$HOME/go' >> ~/.profile
        echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
        echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.profile
        export GOPATH=$HOME/go
        export PATH=$PATH:$GOPATH/bin
    fi
}

# Install system dependencies
install_dependencies() {
    log_step "Installing system dependencies..."
    
    case $OS in
        "ubuntu"|"debian")
            sudo apt-get update
            sudo apt-get install -y git curl wget build-essential make gcc openssl
            ;;
        "centos"|"rhel"|"fedora")
            if command_exists dnf; then
                sudo dnf groupinstall -y "Development Tools"
                sudo dnf install -y git curl wget make gcc openssl
            else
                sudo yum groupinstall -y "Development Tools"
                sudo yum install -y git curl wget make gcc openssl
            fi
            ;;
        "macos")
            # Ensure Xcode command line tools are installed
            xcode-select --install 2>/dev/null || true
            if command_exists brew; then
                brew install git make
            fi
            ;;
        "termux")
            pkg update
            pkg install -y git curl wget make clang openssl
            ;;
        *)
            log_warning "Unknown OS. Please ensure git, make, gcc, and openssl are installed."
            ;;
    esac
}

# Setup database
setup_database() {
    log_step "Setting up PostgreSQL database..."
    
    # Generate secure password
    DB_PASSWORD="tsh_$(openssl rand -hex 16)"
    
    # Create database setup script
    cat > /tmp/setup_db.sql << EOF
-- Create user and database
DO \$\$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = '${DATABASE_USER}') THEN
        CREATE USER ${DATABASE_USER} WITH ENCRYPTED PASSWORD '${DB_PASSWORD}';
    END IF;
END
\$\$;

-- Create database if it doesn't exist
SELECT 'CREATE DATABASE ${DATABASE_NAME} OWNER ${DATABASE_USER}'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '${DATABASE_NAME}');

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE ${DATABASE_NAME} TO ${DATABASE_USER};

-- Connect to the database
\c ${DATABASE_NAME};

-- Create tables
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    chat_id BIGINT UNIQUE NOT NULL,
    username VARCHAR(255),
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    is_admin BOOLEAN DEFAULT false,
    subscription_plan VARCHAR(50) DEFAULT 'free',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS stores (
    id SERIAL PRIMARY KEY,
    owner_chat_id BIGINT NOT NULL,
    store_name VARCHAR(255) NOT NULL,
    bot_token VARCHAR(255) UNIQUE,
    bot_username VARCHAR(255),
    plan_type VARCHAR(50) DEFAULT 'free',
    is_active BOOLEAN DEFAULT true,
    product_count INTEGER DEFAULT 0,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (owner_chat_id) REFERENCES users(chat_id)
);

CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    store_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10,2) DEFAULT 0,
    image_url VARCHAR(500),
    category VARCHAR(255),
    stock_quantity INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (store_id) REFERENCES stores(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    store_id INTEGER NOT NULL,
    customer_chat_id BIGINT NOT NULL,
    customer_name VARCHAR(255),
    customer_phone VARCHAR(50),
    customer_address TEXT,
    total_amount DECIMAL(10,2) DEFAULT 0,
    status VARCHAR(50) DEFAULT 'pending',
    payment_status VARCHAR(50) DEFAULT 'pending',
    order_items JSONB,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (store_id) REFERENCES stores(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS payments (
    id SERIAL PRIMARY KEY,
    store_id INTEGER,
    order_id INTEGER,
    plan_type VARCHAR(50),
    amount DECIMAL(10,2),
    status VARCHAR(50) DEFAULT 'pending',
    payment_method VARCHAR(100),
    transaction_id VARCHAR(255),
    payment_proof VARCHAR(500),
    admin_verified BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (store_id) REFERENCES stores(id),
    FOREIGN KEY (order_id) REFERENCES orders(id)
);

CREATE TABLE IF NOT EXISTS subscriptions (
    id SERIAL PRIMARY KEY,
    store_id INTEGER NOT NULL,
    plan_type VARCHAR(50) NOT NULL,
    starts_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    is_active BOOLEAN DEFAULT true,
    auto_renew BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (store_id) REFERENCES stores(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS bot_sessions (
    id SERIAL PRIMARY KEY,
    store_id INTEGER NOT NULL,
    session_data JSONB,
    is_active BOOLEAN DEFAULT true,
    last_activity TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (store_id) REFERENCES stores(id) ON DELETE CASCADE
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_users_chat_id ON users(chat_id);
CREATE INDEX IF NOT EXISTS idx_stores_owner_chat_id ON stores(owner_chat_id);
CREATE INDEX IF NOT EXISTS idx_stores_bot_token ON stores(bot_token);
CREATE INDEX IF NOT EXISTS idx_products_store_id ON products(store_id);
CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);
CREATE INDEX IF NOT EXISTS idx_orders_store_id ON orders(store_id);
CREATE INDEX IF NOT EXISTS idx_orders_customer_chat_id ON orders(customer_chat_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_payments_store_id ON payments(store_id);
CREATE INDEX IF NOT EXISTS idx_payments_order_id ON payments(order_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_store_id ON subscriptions(store_id);
CREATE INDEX IF NOT EXISTS idx_bot_sessions_store_id ON bot_sessions(store_id);

-- Grant all privileges to the user
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO ${DATABASE_USER};
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO ${DATABASE_USER};
GRANT CREATE ON SCHEMA public TO ${DATABASE_USER};

-- Insert default admin (will be updated with real chat ID later)
INSERT INTO users (chat_id, username, first_name, is_admin, subscription_plan) 
VALUES (123456789, 'admin', 'System Admin', true, 'vip') 
ON CONFLICT (chat_id) DO NOTHING;
EOF
    
    # Execute database setup
    if [[ "$OS" == "termux" ]]; then
        psql -d postgres -f /tmp/setup_db.sql
    else
        sudo -u postgres psql -f /tmp/setup_db.sql
    fi
    
    # Clean up
    rm -f /tmp/setup_db.sql
    
    # Store database credentials
    echo "$DB_PASSWORD" > .db_password
    chmod 600 .db_password
    
    log_success "Database setup completed"
    log_info "Database password saved to .db_password (keep this file secure)"
}

# Create configuration file
create_config() {
    log_step "Creating configuration file..."
    
    if [[ ! -f ".env" ]]; then
        DB_PASSWORD=$(cat .db_password 2>/dev/null || echo "your_db_password_here")
        
        cat > .env << EOF
# Telegram Bot Configuration
BOT_TOKEN=your_bot_token_here
CHANNEL_USERNAME=@your_channel
ADMIN_CHAT_ID=123456789

# Database Configuration
DATABASE_URL=postgres://${DATABASE_USER}:${DB_PASSWORD}@localhost:5432/${DATABASE_NAME}?sslmode=disable
DB_HOST=localhost
DB_PORT=5432
DB_NAME=${DATABASE_NAME}
DB_USER=${DATABASE_USER}
DB_PASSWORD=${DB_PASSWORD}

# Application Configuration
APP_NAME=${APP_NAME}
APP_VERSION=${VERSION}
LOG_LEVEL=info
DEBUG=false

# Subscription Plans
MAX_PRODUCTS_FREE=10
MAX_PRODUCTS_PRO=200
MAX_PRODUCTS_VIP=999999

# Pricing (in your local currency)
PRO_PLAN_PRICE=50000
VIP_PLAN_PRICE=150000

# Payment Configuration
PAYMENT_CARD_NUMBER=1234-5678-9012-3456
PAYMENT_CARD_NAME=CodeRoot Store
PAYMENT_INSTRUCTIONS=Please transfer to the above card and send payment receipt

# Bot Features
ENABLE_PRODUCT_IMAGES=true
ENABLE_ORDER_TRACKING=true
ENABLE_ANALYTICS=true
AUTO_APPROVE_PAYMENTS=false

# System Configuration
MAX_FILE_UPLOAD_SIZE=10MB
SESSION_TIMEOUT=3600
RATE_LIMIT_REQUESTS_PER_MINUTE=60

# Backup Configuration
ENABLE_AUTO_BACKUP=true
BACKUP_INTERVAL_HOURS=24
BACKUP_RETENTION_DAYS=30
EOF
        
        log_success "Configuration file created: .env"
        log_warning "Please edit .env file with your actual bot token and settings!"
    else
        log_success "Configuration file already exists"
    fi
}

# Download Go modules and build
build_application() {
    log_step "Downloading Go modules..."
    
    if [[ ! -f "go.mod" ]]; then
        log_info "Initializing Go module..."
        go mod init $APP_NAME
    fi
    
    # Ensure go.mod has required dependencies
    go mod edit -require github.com/go-telegram-bot-api/telegram-bot-api/v5@latest
    go mod edit -require github.com/joho/godotenv@latest
    go mod edit -require gorm.io/driver/postgres@latest
    go mod edit -require gorm.io/gorm@latest
    
    go mod download
    go mod tidy
    
    log_step "Building application..."
    
    # Build for current platform
    mkdir -p bin
    go build -v -o bin/${APP_NAME} .
    
    # Make executable
    chmod +x bin/${APP_NAME}
    
    log_success "Application built successfully: bin/${APP_NAME}"
}

# Create run script
create_run_script() {
    log_step "Creating run script..."
    
    cat > run.sh << 'EOF'
#!/bin/bash

# Telegram Store Hub - Run Script
APP_NAME="telegram-store-hub"

# Load environment variables
if [[ -f ".env" ]]; then
    set -a  # automatically export all variables
    source .env
    set +a
else
    echo "❌ Configuration file (.env) not found!"
    echo "Please run the setup script first: ./scripts/setup.sh"
    exit 1
fi

# Check if binary exists
if [[ ! -f "bin/${APP_NAME}" ]]; then
    echo "❌ Application binary not found!"
    echo "Please build the application first: make build"
    exit 1
fi

# Check if bot token is configured
if [[ "$BOT_TOKEN" == "your_bot_token_here" ]]; then
    echo "❌ Bot token not configured!"
    echo "Please edit .env file and add your bot token from @BotFather"
    exit 1
fi

# Check database connectivity
echo "🔍 Checking database connection..."
if ! command -v psql >/dev/null 2>&1; then
    echo "❌ PostgreSQL client not found!"
    exit 1
fi

# Test database connection
export PGPASSWORD="$DB_PASSWORD"
if ! psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c '\q' >/dev/null 2>&1; then
    echo "❌ Cannot connect to database!"
    echo "Please check your database configuration in .env file"
    exit 1
fi

echo "✅ Database connection successful"

# Create logs directory
mkdir -p logs

# Start the application
echo "🚀 Starting $APP_NAME..."
echo "📝 Logs will be saved to logs/app.log"
echo "⏹️  Press Ctrl+C to stop"

# Run with output to both console and log file
./bin/${APP_NAME} 2>&1 | tee logs/app.log
EOF
    
    chmod +x run.sh
    log_success "Run script created: ./run.sh"
}

# Create systemd service (Linux only)
create_systemd_service() {
    if [[ "$OS" != "ubuntu" && "$OS" != "debian" && "$OS" != "centos" && "$OS" != "rhel" && "$OS" != "fedora" ]]; then
        return
    fi
    
    log_step "Creating systemd service..."
    
    CURRENT_DIR=$(pwd)
    USER_NAME=$(whoami)
    
    sudo tee /etc/systemd/system/${APP_NAME}.service > /dev/null << EOF
[Unit]
Description=Telegram Store Hub Bot
After=network.target postgresql.service
Wants=postgresql.service

[Service]
Type=simple
User=${USER_NAME}
Group=${USER_NAME}
WorkingDirectory=${CURRENT_DIR}
EnvironmentFile=${CURRENT_DIR}/.env
ExecStart=${CURRENT_DIR}/bin/${APP_NAME}
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
ReadWritePaths=${CURRENT_DIR}/logs

[Install]
WantedBy=multi-user.target
EOF
    
    sudo systemctl daemon-reload
    log_success "Systemd service created"
    log_info "To enable: sudo systemctl enable ${APP_NAME}"
    log_info "To start: sudo systemctl start ${APP_NAME}"
}

# Main installation function
main() {
    echo "🚀 Telegram Store Hub - Complete Setup Script"
    echo "=============================================="
    echo ""
    
    # Check if running in project directory
    if [[ ! -f "go.mod" && ! -f "main.go" ]]; then
        log_error "Please run this script from the project root directory"
        exit 1
    fi
    
    # Detect operating system
    detect_os
    log_info "Detected OS: $OS"
    
    # Check if running as root (not recommended)
    if [[ $EUID -eq 0 ]]; then
        log_warning "Running as root. This is not recommended for development."
        read -p "Continue anyway? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
    
    # Step 1: Install system dependencies
    log_step "Step 1/7: Installing system dependencies..."
    install_dependencies
    
    # Step 2: Install PostgreSQL
    log_step "Step 2/7: Installing PostgreSQL..."
    install_postgresql
    
    # Step 3: Install Go
    log_step "Step 3/7: Installing Go programming language..."
    install_go
    
    # Reload bash profile to get Go in PATH
    source ~/.bashrc 2>/dev/null || true
    source ~/.profile 2>/dev/null || true
    
    # Step 4: Setup database
    log_step "Step 4/7: Setting up database..."
    setup_database
    
    # Step 5: Create configuration
    log_step "Step 5/7: Creating configuration..."
    create_config
    
    # Step 6: Build application
    log_step "Step 6/7: Building application..."
    build_application
    
    # Step 7: Create run scripts and services
    log_step "Step 7/7: Creating run scripts..."
    create_run_script
    create_systemd_service
    
    # Final success message
    echo ""
    echo "🎉 Installation completed successfully!"
    echo "======================================"
    echo ""
    log_success "All prerequisites installed and configured"
    log_success "Database created and tables initialized"
    log_success "Application built and ready to run"
    echo ""
    
    # Next steps
    echo "📋 Next steps:"
    echo "  1. Get your bot token from @BotFather on Telegram"
    echo "  2. Edit .env file: nano .env"
    echo "  3. Update BOT_TOKEN with your actual bot token"
    echo "  4. Update ADMIN_CHAT_ID with your Telegram chat ID"
    echo "  5. Update CHANNEL_USERNAME with your channel"
    echo "  6. Run the bot: ./run.sh"
    echo ""
    
    # Useful commands
    echo "🔧 Useful commands:"
    echo "  Start bot: ./run.sh"
    echo "  Build only: make build"
    echo "  View logs: tail -f logs/app.log"
    echo "  Test database: psql -h localhost -U ${DATABASE_USER} -d ${DATABASE_NAME}"
    echo ""
    
    if [[ "$OS" == "ubuntu" || "$OS" == "debian" || "$OS" == "centos" || "$OS" == "rhel" || "$OS" == "fedora" ]]; then
        echo "🔧 Systemd service commands:"
        echo "  Enable service: sudo systemctl enable ${APP_NAME}"
        echo "  Start service: sudo systemctl start ${APP_NAME}"
        echo "  Status: sudo systemctl status ${APP_NAME}"
        echo "  Logs: sudo journalctl -u ${APP_NAME} -f"
        echo ""
    fi
    
    # Final warnings
    log_warning "Important security notes:"
    echo "  - Keep your .env file secure and never commit it to git"
    echo "  - The database password is saved in .db_password file"
    echo "  - Change the default admin chat ID in .env"
    echo "  - Configure your firewall appropriately"
    echo ""
    
    log_success "Setup completed! Your Telegram Store Hub is ready to use."
}

# Run main function
main "$@"