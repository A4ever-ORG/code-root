#!/bin/bash

# Telegram Store Hub - Improved Installation Script
# Complete system installation with enhanced error handling and validation

set -e

# Script configuration
APP_NAME="telegram-store-hub"
APP_VERSION="1.0.0"
INSTALL_DIR="/usr/local/bin"
SERVICE_DIR="/etc/systemd/system"
CONFIG_DIR="/etc/telegram-store-hub"
LOG_DIR="/var/log/telegram-store-hub"
DATA_DIR="/var/lib/telegram-store-hub"
BACKUP_DIR="/var/backups/telegram-store-hub"

# User and group for service
SERVICE_USER="telegram-hub"
SERVICE_GROUP="telegram-hub"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_debug() {
    if [[ "${DEBUG:-0}" == "1" ]]; then
        echo -e "${PURPLE}[DEBUG]${NC} $1"
    fi
}

# Error handling
error_exit() {
    log_error "$1"
    cleanup_on_error
    exit 1
}

cleanup_on_error() {
    log_warning "Cleaning up partial installation..."
    # Remove partially installed files
    rm -f "${INSTALL_DIR}/${APP_NAME}"
    rm -f "${SERVICE_DIR}/${APP_NAME}.service"
    # Don't remove config directory as it might contain user data
}

# Check if running as root
check_root() {
    if [[ $EUID -ne 0 ]]; then
        error_exit "This script must be run as root (use sudo)"
    fi
}

# Check system requirements
check_system_requirements() {
    log_info "Checking system requirements..."
    
    # Check OS
    if ! command -v systemctl &> /dev/null; then
        error_exit "This script requires systemd (Ubuntu 16.04+, CentOS 7+, etc.)"
    fi
    
    # Check available space
    AVAILABLE_SPACE=$(df / | awk 'NR==2 {print $4}')
    REQUIRED_SPACE=102400 # 100MB in KB
    if [[ $AVAILABLE_SPACE -lt $REQUIRED_SPACE ]]; then
        error_exit "Insufficient disk space. Need at least 100MB free."
    fi
    
    # Check PostgreSQL
    if ! command -v psql &> /dev/null; then
        log_warning "PostgreSQL client not found. Database setup will be manual."
    fi
    
    log_success "System requirements check passed"
}

# Install system dependencies
install_dependencies() {
    log_info "Installing system dependencies..."
    
    # Detect package manager and install dependencies
    if command -v apt-get &> /dev/null; then
        apt-get update
        apt-get install -y curl wget postgresql-client sudo systemd
    elif command -v yum &> /dev/null; then
        yum update -y
        yum install -y curl wget postgresql sudo systemd
    elif command -v pacman &> /dev/null; then
        pacman -Sy
        pacman -S --noconfirm curl wget postgresql sudo systemd
    else
        log_warning "Unknown package manager. Please install dependencies manually:"
        echo "  - curl"
        echo "  - wget" 
        echo "  - postgresql-client"
        echo "  - sudo"
        echo "  - systemd"
    fi
    
    log_success "Dependencies installed"
}

# Create system user and group
create_service_user() {
    log_info "Creating service user and group..."
    
    # Create group if it doesn't exist
    if ! getent group "$SERVICE_GROUP" &> /dev/null; then
        groupadd --system "$SERVICE_GROUP"
        log_success "Created group: $SERVICE_GROUP"
    else
        log_info "Group $SERVICE_GROUP already exists"
    fi
    
    # Create user if it doesn't exist
    if ! getent passwd "$SERVICE_USER" &> /dev/null; then
        useradd --system --gid "$SERVICE_GROUP" --no-create-home \
                --home-dir "$DATA_DIR" --shell /bin/false \
                --comment "Telegram Store Hub service user" "$SERVICE_USER"
        log_success "Created user: $SERVICE_USER"
    else
        log_info "User $SERVICE_USER already exists"
    fi
}

# Create directory structure
create_directories() {
    log_info "Creating directory structure..."
    
    # Create directories with proper permissions
    mkdir -p "$CONFIG_DIR"
    mkdir -p "$LOG_DIR"
    mkdir -p "$DATA_DIR/db"
    mkdir -p "$DATA_DIR/uploads"
    mkdir -p "$DATA_DIR/backups"
    mkdir -p "$BACKUP_DIR"
    
    # Set ownership and permissions
    chown root:root "$CONFIG_DIR"
    chmod 755 "$CONFIG_DIR"
    
    chown "$SERVICE_USER:$SERVICE_GROUP" "$LOG_DIR"
    chmod 750 "$LOG_DIR"
    
    chown "$SERVICE_USER:$SERVICE_GROUP" "$DATA_DIR"
    chmod 750 "$DATA_DIR"
    
    chown "$SERVICE_USER:$SERVICE_GROUP" "$DATA_DIR/db"
    chmod 700 "$DATA_DIR/db"
    
    chown "$SERVICE_USER:$SERVICE_GROUP" "$DATA_DIR/uploads"
    chmod 750 "$DATA_DIR/uploads"
    
    chown "$SERVICE_USER:$SERVICE_GROUP" "$DATA_DIR/backups"
    chmod 750 "$DATA_DIR/backups"
    
    chown root:root "$BACKUP_DIR"
    chmod 755 "$BACKUP_DIR"
    
    log_success "Directory structure created"
}

# Install binary
install_binary() {
    log_info "Installing application binary..."
    
    # Check if binary exists
    if [[ ! -f "./${APP_NAME}" ]]; then
        error_exit "Binary ${APP_NAME} not found. Run 'make build' first."
    fi
    
    # Backup existing binary if it exists
    if [[ -f "${INSTALL_DIR}/${APP_NAME}" ]]; then
        BACKUP_NAME="${APP_NAME}_$(date +%Y%m%d_%H%M%S)"
        cp "${INSTALL_DIR}/${APP_NAME}" "${BACKUP_DIR}/${BACKUP_NAME}"
        log_info "Backed up existing binary to ${BACKUP_DIR}/${BACKUP_NAME}"
    fi
    
    # Install new binary
    cp "./${APP_NAME}" "${INSTALL_DIR}/"
    chmod 755 "${INSTALL_DIR}/${APP_NAME}"
    chown root:root "${INSTALL_DIR}/${APP_NAME}"
    
    # Verify installation
    if "${INSTALL_DIR}/${APP_NAME}" --version &> /dev/null || true; then
        log_success "Binary installed successfully"
    else
        log_warning "Binary installed but version check failed"
    fi
}

# Setup configuration
setup_configuration() {
    log_info "Setting up configuration..."
    
    # Install configuration template
    if [[ -f ".env.example" ]]; then
        cp ".env.example" "${CONFIG_DIR}/${APP_NAME}.env.example"
        chown root:root "${CONFIG_DIR}/${APP_NAME}.env.example"
        chmod 644 "${CONFIG_DIR}/${APP_NAME}.env.example"
        
        # Create actual config if it doesn't exist
        if [[ ! -f "${CONFIG_DIR}/${APP_NAME}.env" ]]; then
            cp ".env.example" "${CONFIG_DIR}/${APP_NAME}.env"
            chown root:"$SERVICE_GROUP" "${CONFIG_DIR}/${APP_NAME}.env"
            chmod 640 "${CONFIG_DIR}/${APP_NAME}.env"
            log_success "Created configuration file: ${CONFIG_DIR}/${APP_NAME}.env"
            log_warning "Please edit this file with your settings!"
        else
            log_info "Configuration file already exists"
        fi
    else
        log_warning ".env.example not found, creating basic configuration"
        create_basic_config
    fi
}

# Create basic configuration if template doesn't exist
create_basic_config() {
    cat > "${CONFIG_DIR}/${APP_NAME}.env" << EOF
# Telegram Bot Configuration
BOT_TOKEN=your_telegram_bot_token_here
ADMIN_CHAT_ID=your_admin_telegram_id

# Database Configuration
DATABASE_URL=postgres://telegram_hub:password@localhost:5432/telegram_store_hub?sslmode=disable

# Channel Configuration (for force join) - Optional
FORCE_JOIN_CHANNEL_ID=-1001234567890
FORCE_JOIN_CHANNEL_USERNAME=@your_channel

# Payment Configuration
PAYMENT_CARD_NUMBER=1234-5678-9012-3456
PAYMENT_CARD_HOLDER=Your Name Here

# Subscription Prices (in Toman)
FREE_PLAN_PRICE=0
PRO_PLAN_PRICE=50000
VIP_PLAN_PRICE=150000

# Commission Rates (percentage)
FREE_PLAN_COMMISSION=5
PRO_PLAN_COMMISSION=5
VIP_PLAN_COMMISSION=0

# Security
DEBUG=false
EOF
    
    chown root:"$SERVICE_GROUP" "${CONFIG_DIR}/${APP_NAME}.env"
    chmod 640 "${CONFIG_DIR}/${APP_NAME}.env"
}

# Create systemd service
create_systemd_service() {
    log_info "Creating systemd service..."
    
    cat > "${SERVICE_DIR}/${APP_NAME}.service" << EOF
[Unit]
Description=Telegram Store Hub Bot
Documentation=https://github.com/yourusername/telegram-store-hub
After=network.target postgresql.service
Wants=postgresql.service
StartLimitIntervalSec=0

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_GROUP
WorkingDirectory=$DATA_DIR
Environment="CONFIG_FILE=${CONFIG_DIR}/${APP_NAME}.env"
EnvironmentFile=${CONFIG_DIR}/${APP_NAME}.env
ExecStart=${INSTALL_DIR}/${APP_NAME}
ExecReload=/bin/kill -USR1 \$MAINPID
Restart=always
RestartSec=10
StartLimitBurst=5

# Output to journal
StandardOutput=journal
StandardError=journal
SyslogIdentifier=${APP_NAME}

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ReadWritePaths=$DATA_DIR $LOG_DIR
ProtectHome=true
ProtectKernelTunables=true
ProtectKernelModules=true
ProtectControlGroups=true
RestrictRealtime=true
RestrictSUIDSGID=true
RemoveIPC=true
PrivateDevices=true

# Resource limits
LimitNOFILE=65536
LimitNPROC=4096

[Install]
WantedBy=multi-user.target
EOF
    
    # Set proper permissions
    chown root:root "${SERVICE_DIR}/${APP_NAME}.service"
    chmod 644 "${SERVICE_DIR}/${APP_NAME}.service"
    
    # Reload systemd
    systemctl daemon-reload
    
    log_success "Systemd service created"
}

# Setup logging
setup_logging() {
    log_info "Setting up logging..."
    
    # Create logrotate configuration
    cat > "/etc/logrotate.d/${APP_NAME}" << EOF
$LOG_DIR/*.log {
    daily
    missingok
    rotate 52
    compress
    delaycompress
    notifempty
    create 640 $SERVICE_USER $SERVICE_GROUP
    postrotate
        systemctl reload-or-restart ${APP_NAME} > /dev/null 2>&1 || true
    endscript
}
EOF
    
    # Create rsyslog configuration for structured logging
    if [[ -d "/etc/rsyslog.d" ]]; then
        cat > "/etc/rsyslog.d/49-${APP_NAME}.conf" << EOF
# Telegram Store Hub logging
:programname, isequal, "${APP_NAME}" $LOG_DIR/${APP_NAME}.log
& stop
EOF
        systemctl restart rsyslog &> /dev/null || true
    fi
    
    log_success "Logging configuration completed"
}

# Setup database
setup_database() {
    log_info "Setting up database..."
    
    # Check if PostgreSQL is running
    if systemctl is-active --quiet postgresql; then
        log_info "PostgreSQL is running"
        
        # Create database and user if PostgreSQL is available
        if command -v createdb &> /dev/null; then
            setup_postgresql_database
        else
            log_warning "PostgreSQL tools not available. Manual database setup required."
            show_manual_database_setup
        fi
    else
        log_warning "PostgreSQL is not running. Manual database setup required."
        show_manual_database_setup
    fi
}

# Setup PostgreSQL database automatically
setup_postgresql_database() {
    log_info "Configuring PostgreSQL database..."
    
    # Generate secure password
    DB_PASSWORD=$(openssl rand -base64 32)
    DB_NAME="telegram_store_hub"
    DB_USER="telegram_hub"
    
    # Create database and user
    sudo -u postgres psql << EOF
CREATE USER $DB_USER WITH PASSWORD '$DB_PASSWORD';
CREATE DATABASE $DB_NAME OWNER $DB_USER;
GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;
\q
EOF
    
    # Update configuration with database URL
    sed -i "s|DATABASE_URL=.*|DATABASE_URL=postgres://$DB_USER:$DB_PASSWORD@localhost:5432/$DB_NAME?sslmode=disable|" \
        "${CONFIG_DIR}/${APP_NAME}.env"
    
    log_success "Database setup completed"
    log_info "Database: $DB_NAME"
    log_info "User: $DB_USER"
    log_info "Password: $DB_PASSWORD"
}

# Show manual database setup instructions
show_manual_database_setup() {
    log_info "Manual database setup required:"
    echo
    echo "1. Install and start PostgreSQL:"
    echo "   Ubuntu/Debian: sudo apt install postgresql postgresql-contrib"
    echo "   CentOS/RHEL:   sudo yum install postgresql postgresql-server"
    echo
    echo "2. Create database and user:"
    echo "   sudo -u postgres psql"
    echo "   CREATE USER telegram_hub WITH PASSWORD 'your_secure_password';"
    echo "   CREATE DATABASE telegram_store_hub OWNER telegram_hub;"
    echo "   GRANT ALL PRIVILEGES ON DATABASE telegram_store_hub TO telegram_hub;"
    echo "   \\q"
    echo
    echo "3. Update configuration in ${CONFIG_DIR}/${APP_NAME}.env"
    echo "   DATABASE_URL=postgres://telegram_hub:your_secure_password@localhost:5432/telegram_store_hub?sslmode=disable"
    echo
}

# Enable and start service
enable_service() {
    log_info "Enabling and starting service..."
    
    # Enable service
    systemctl enable "${APP_NAME}.service"
    
    # Start service
    if systemctl start "${APP_NAME}.service"; then
        log_success "Service started successfully"
        
        # Check service status
        sleep 2
        if systemctl is-active --quiet "${APP_NAME}.service"; then
            log_success "Service is running"
        else
            log_warning "Service started but may have issues. Check: journalctl -u ${APP_NAME}"
        fi
    else
        log_error "Failed to start service. Check configuration and logs."
        systemctl status "${APP_NAME}.service" --no-pager || true
    fi
}

# Create uninstall script
create_uninstall_script() {
    log_info "Creating uninstall script..."
    
    cat > "${INSTALL_DIR}/${APP_NAME}-uninstall" << 'EOF'
#!/bin/bash
# Telegram Store Hub - Uninstall Script

APP_NAME="telegram-store-hub"
INSTALL_DIR="/usr/local/bin"
SERVICE_DIR="/etc/systemd/system"
CONFIG_DIR="/etc/telegram-store-hub"
LOG_DIR="/var/log/telegram-store-hub"
DATA_DIR="/var/lib/telegram-store-hub"

echo "Uninstalling Telegram Store Hub..."

# Stop and disable service
systemctl stop "${APP_NAME}.service" 2>/dev/null || true
systemctl disable "${APP_NAME}.service" 2>/dev/null || true

# Remove service file
rm -f "${SERVICE_DIR}/${APP_NAME}.service"

# Remove binary
rm -f "${INSTALL_DIR}/${APP_NAME}"
rm -f "${INSTALL_DIR}/${APP_NAME}-uninstall"

# Remove logrotate config
rm -f "/etc/logrotate.d/${APP_NAME}"

# Remove rsyslog config
rm -f "/etc/rsyslog.d/49-${APP_NAME}.conf"

# Ask about data removal
read -p "Remove all data and configuration? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    rm -rf "$CONFIG_DIR"
    rm -rf "$LOG_DIR"
    rm -rf "$DATA_DIR"
    userdel telegram-hub 2>/dev/null || true
    groupdel telegram-hub 2>/dev/null || true
    echo "All data removed"
else
    echo "Data preserved in:"
    echo "  - Configuration: $CONFIG_DIR"
    echo "  - Logs: $LOG_DIR"
    echo "  - Data: $DATA_DIR"
fi

systemctl daemon-reload
echo "Uninstallation completed"
EOF
    
    chmod +x "${INSTALL_DIR}/${APP_NAME}-uninstall"
    log_success "Uninstall script created at ${INSTALL_DIR}/${APP_NAME}-uninstall"
}

# Post-installation configuration check
post_install_check() {
    log_info "Running post-installation checks..."
    
    # Check binary
    if [[ -x "${INSTALL_DIR}/${APP_NAME}" ]]; then
        log_success "Binary is executable"
    else
        log_error "Binary is not executable"
    fi
    
    # Check service file
    if [[ -f "${SERVICE_DIR}/${APP_NAME}.service" ]]; then
        log_success "Service file exists"
    else
        log_error "Service file missing"
    fi
    
    # Check configuration
    if [[ -f "${CONFIG_DIR}/${APP_NAME}.env" ]]; then
        log_success "Configuration file exists"
        
        # Check for default values that need to be changed
        if grep -q "your_telegram_bot_token_here" "${CONFIG_DIR}/${APP_NAME}.env"; then
            log_warning "Bot token needs to be configured"
        fi
        
        if grep -q "your_admin_telegram_id" "${CONFIG_DIR}/${APP_NAME}.env"; then
            log_warning "Admin chat ID needs to be configured"
        fi
    else
        log_error "Configuration file missing"
    fi
    
    # Check permissions
    CONFIG_PERMS=$(stat -c "%a" "${CONFIG_DIR}/${APP_NAME}.env" 2>/dev/null || echo "000")
    if [[ "$CONFIG_PERMS" == "640" ]]; then
        log_success "Configuration file permissions correct"
    else
        log_warning "Configuration file permissions may be incorrect: $CONFIG_PERMS"
    fi
}

# Show post-installation instructions
show_post_install_instructions() {
    echo
    log_success "🎉 Installation completed successfully!"
    echo
    echo "📋 Next steps:"
    echo "1. Edit configuration file: ${CONFIG_DIR}/${APP_NAME}.env"
    echo "   - Set your Telegram bot token"
    echo "   - Set admin chat ID"
    echo "   - Configure database connection"
    echo
    echo "2. Start the service:"
    echo "   sudo systemctl start ${APP_NAME}"
    echo
    echo "3. Check service status:"
    echo "   sudo systemctl status ${APP_NAME}"
    echo "   sudo journalctl -u ${APP_NAME} -f"
    echo
    echo "📁 Important directories:"
    echo "   Configuration: ${CONFIG_DIR}"
    echo "   Logs:          ${LOG_DIR}"
    echo "   Data:          ${DATA_DIR}"
    echo
    echo "🔧 Management commands:"
    echo "   Start:     sudo systemctl start ${APP_NAME}"
    echo "   Stop:      sudo systemctl stop ${APP_NAME}"
    echo "   Restart:   sudo systemctl restart ${APP_NAME}"
    echo "   Status:    sudo systemctl status ${APP_NAME}"
    echo "   Logs:      sudo journalctl -u ${APP_NAME}"
    echo "   Uninstall: sudo ${INSTALL_DIR}/${APP_NAME}-uninstall"
    echo
    echo "⚠️  Remember to:"
    echo "   - Configure your bot token and settings"
    echo "   - Set up SSL certificates for production"
    echo "   - Configure firewall rules if needed"
    echo "   - Set up monitoring and alerting"
    echo
}

# Main installation function
main() {
    echo "🚀 Telegram Store Hub - Enhanced Installation"
    echo "============================================="
    echo "Version: $APP_VERSION"
    echo "Target: Production deployment"
    echo
    
    # Pre-installation checks
    check_root
    check_system_requirements
    
    # Installation steps
    install_dependencies
    create_service_user
    create_directories
    install_binary
    setup_configuration
    create_systemd_service
    setup_logging
    setup_database
    create_uninstall_script
    
    # Post-installation
    post_install_check
    enable_service
    show_post_install_instructions
}

# Handle script arguments
case "${1:-}" in
    --help|-h)
        echo "Usage: $0 [OPTIONS]"
        echo "Options:"
        echo "  --help, -h     Show this help message"
        echo "  --debug        Enable debug output"
        exit 0
        ;;
    --debug)
        DEBUG=1
        main
        ;;
    *)
        main
        ;;
esac