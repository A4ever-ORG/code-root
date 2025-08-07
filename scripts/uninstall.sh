#!/bin/bash

# Telegram Store Hub - Uninstall Script
# Safely removes the application while optionally preserving data

set -e

APP_NAME="telegram-store-hub"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}ℹ️  $1${NC}"; }
log_success() { echo -e "${GREEN}✅ $1${NC}"; }
log_warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }
log_error() { echo -e "${RED}❌ $1${NC}"; }

echo -e "${RED}🗑️  Telegram Store Hub - Uninstall Script${NC}"
echo "========================================="
echo ""
log_warning "This will remove the Telegram Store Hub from your system"
echo ""

# Ask for confirmation
read -p "Are you sure you want to uninstall? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    log_info "Uninstall cancelled"
    exit 0
fi

# Ask about data preservation
echo ""
log_info "What would you like to do with your data?"
echo "1) Remove everything (application, database, logs, config)"
echo "2) Keep database and configuration (remove only application)"
echo "3) Create backup before removing everything"
echo ""
read -p "Choose option (1/2/3): " -n 1 -r
echo ""

KEEP_DATA=false
CREATE_BACKUP=false

case $REPLY in
    1)
        log_warning "Will remove everything"
        ;;
    2)
        KEEP_DATA=true
        log_info "Will keep database and configuration"
        ;;
    3)
        CREATE_BACKUP=true
        log_info "Will create backup before removal"
        ;;
    *)
        log_error "Invalid option"
        exit 1
        ;;
esac

# Create backup if requested
if [[ "$CREATE_BACKUP" == true ]]; then
    BACKUP_DIR="uninstall_backup_$(date +%Y%m%d_%H%M%S)"
    log_info "Creating backup in $BACKUP_DIR..."
    
    mkdir -p $BACKUP_DIR
    
    # Backup configuration
    [[ -f ".env" ]] && cp .env $BACKUP_DIR/
    [[ -f ".db_password" ]] && cp .db_password $BACKUP_DIR/
    
    # Backup database
    if command -v pg_dump >/dev/null 2>&1 && [[ -f ".env" ]]; then
        source .env 2>/dev/null || true
        if [[ -n "$DB_PASSWORD" ]]; then
            export PGPASSWORD="$DB_PASSWORD"
            pg_dump -h "${DB_HOST:-localhost}" -p "${DB_PORT:-5432}" -U "${DB_USER:-telegram_hub}" "${DB_NAME:-telegram_store_hub}" > $BACKUP_DIR/database_backup.sql 2>/dev/null || log_warning "Database backup failed"
        fi
    fi
    
    # Backup logs
    [[ -d "logs" ]] && cp -r logs $BACKUP_DIR/
    
    log_success "Backup created in $BACKUP_DIR"
fi

# Stop running services
log_info "Stopping services..."

if systemctl is-active --quiet $APP_NAME 2>/dev/null; then
    sudo systemctl stop $APP_NAME
    sudo systemctl disable $APP_NAME
    log_success "Systemd service stopped and disabled"
fi

# Kill any running processes
pkill -f $APP_NAME 2>/dev/null || true

# Remove systemd service file
if [[ -f "/etc/systemd/system/${APP_NAME}.service" ]]; then
    sudo rm -f "/etc/systemd/system/${APP_NAME}.service"
    sudo systemctl daemon-reload
    log_success "Systemd service file removed"
fi

# Remove binary from system
if [[ -f "/usr/local/bin/${APP_NAME}" ]]; then
    sudo rm -f "/usr/local/bin/${APP_NAME}"
    log_success "System binary removed"
fi

# Remove application directory content
log_info "Removing application files..."

# Remove binaries
rm -rf bin/
rm -rf build/
rm -rf builds/

# Remove build artifacts
rm -f ${APP_NAME}
rm -f *.exe

if [[ "$KEEP_DATA" != true ]]; then
    # Remove configuration
    rm -f .env
    rm -f .db_password
    
    # Remove logs
    rm -rf logs/
    
    log_success "Configuration and logs removed"
    
    # Remove database
    if command -v psql >/dev/null 2>&1; then
        log_warning "Removing database..."
        
        # Try to load config for database details
        if [[ -f "$BACKUP_DIR/.env" ]]; then
            source "$BACKUP_DIR/.env" 2>/dev/null || true
        fi
        
        DATABASE_NAME="${DB_NAME:-telegram_store_hub}"
        DATABASE_USER="${DB_USER:-telegram_hub}"
        
        # Remove database and user
        sudo -u postgres psql << EOF 2>/dev/null || true
DROP DATABASE IF EXISTS ${DATABASE_NAME};
DROP USER IF EXISTS ${DATABASE_USER};
EOF
        log_success "Database removed"
    fi
else
    log_info "Database and configuration preserved"
fi

# Remove system configuration directories (if they exist)
if [[ -d "/etc/${APP_NAME}" ]]; then
    if [[ "$KEEP_DATA" != true ]]; then
        sudo rm -rf "/etc/${APP_NAME}"
        log_success "System configuration directory removed"
    else
        log_info "System configuration preserved"
    fi
fi

# Remove log directories (if they exist)
if [[ -d "/var/log/${APP_NAME}" ]]; then
    if [[ "$KEEP_DATA" != true ]]; then
        sudo rm -rf "/var/log/${APP_NAME}"
        log_success "System log directory removed"
    fi
fi

# Remove system user (if exists)
if id "${APP_NAME//-/_}" &>/dev/null; then
    if [[ "$KEEP_DATA" != true ]]; then
        sudo userdel "${APP_NAME//-/_}" 2>/dev/null || true
        log_success "System user removed"
    fi
fi

echo ""
log_success "Uninstall completed!"
echo ""

if [[ "$CREATE_BACKUP" == true ]]; then
    log_info "Your data has been backed up to: $BACKUP_DIR"
    echo "To restore:"
    echo "  - Configuration: cp $BACKUP_DIR/.env ."
    echo "  - Database: psql -U username -d dbname < $BACKUP_DIR/database_backup.sql"
    echo ""
fi

if [[ "$KEEP_DATA" == true ]]; then
    log_info "Your data has been preserved:"
    echo "  - Configuration: .env"
    echo "  - Database: telegram_store_hub (PostgreSQL)"
    echo "  - Database user: telegram_hub"
    echo ""
    log_info "To reinstall: run ./scripts/setup.sh"
fi

# Remove Go modules (optional)
read -p "Remove Go modules cache? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    go clean -modcache 2>/dev/null || true
    log_success "Go modules cache cleared"
fi

echo ""
log_info "Optional cleanup:"
echo "  - Remove PostgreSQL (if no longer needed): sudo apt remove postgresql"
echo "  - Remove Go (if no longer needed): sudo rm -rf /usr/local/go"
echo ""
log_success "Telegram Store Hub has been uninstalled from your system"