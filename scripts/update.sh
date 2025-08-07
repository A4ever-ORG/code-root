#!/bin/bash

# Telegram Store Hub - Update Script
# Updates the application while preserving data and configuration

set -e

APP_NAME="telegram-store-hub"
BACKUP_DIR="backups/$(date +%Y%m%d_%H%M%S)"

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

echo -e "${BLUE}🔄 Telegram Store Hub - Update Script${NC}"
echo "====================================="

# Check if running in project directory
if [[ ! -f "go.mod" ]]; then
    log_error "Please run this script from the project root directory"
    exit 1
fi

# Check if bot is currently running
if pgrep -f "$APP_NAME" > /dev/null; then
    log_warning "Bot is currently running. Stopping it..."
    
    # Try graceful stop first
    if systemctl is-active --quiet $APP_NAME 2>/dev/null; then
        sudo systemctl stop $APP_NAME
        log_info "Stopped systemd service"
    else
        # Kill running process
        pkill -f $APP_NAME || true
        log_info "Stopped running process"
    fi
    
    sleep 2
fi

# Create backup
log_info "Creating backup..."
mkdir -p $BACKUP_DIR

# Backup configuration and database
if [[ -f ".env" ]]; then
    cp .env $BACKUP_DIR/
    log_success "Configuration backed up"
fi

if [[ -f ".db_password" ]]; then
    cp .db_password $BACKUP_DIR/
fi

# Backup database
if command -v pg_dump >/dev/null 2>&1 && [[ -f ".env" ]]; then
    source .env
    export PGPASSWORD="$DB_PASSWORD"
    pg_dump -h "${DB_HOST:-localhost}" -p "${DB_PORT:-5432}" -U "${DB_USER:-telegram_hub}" "${DB_NAME:-telegram_store_hub}" > $BACKUP_DIR/database_backup.sql 2>/dev/null || log_warning "Database backup failed"
fi

# Backup current binary
if [[ -f "bin/$APP_NAME" ]]; then
    cp bin/$APP_NAME $BACKUP_DIR/${APP_NAME}.old
    log_success "Current binary backed up"
fi

# Pull latest changes (if git repository)
if [[ -d ".git" ]]; then
    log_info "Pulling latest changes from repository..."
    git stash push -m "Auto-stash before update $(date)"
    git pull origin main || git pull origin master || log_warning "Git pull failed"
fi

# Update Go modules
log_info "Updating Go modules..."
go mod download
go mod tidy

# Build new version
log_info "Building updated application..."
mkdir -p bin
go build -v -o bin/${APP_NAME} .
chmod +x bin/${APP_NAME}

# Run database migrations (if migration system exists)
if [[ -f "migrations" ]] || grep -q "migrate" Makefile 2>/dev/null; then
    log_info "Running database migrations..."
    make migrate 2>/dev/null || log_warning "No migrations to run"
fi

# Restart service
log_info "Starting updated application..."
if systemctl list-unit-files | grep -q "$APP_NAME.service"; then
    sudo systemctl start $APP_NAME
    sudo systemctl status $APP_NAME --no-pager -l
    log_success "Systemd service started"
else
    log_info "Systemd service not found. You can start manually with: ./run.sh"
fi

# Verify update
log_info "Verifying update..."
sleep 3

if [[ -f "bin/$APP_NAME" ]]; then
    VERSION_INFO=$(./bin/$APP_NAME -version 2>/dev/null || echo "Version info not available")
    log_success "Binary updated successfully"
    log_info "Version: $VERSION_INFO"
else
    log_error "Binary build failed"
    
    # Restore backup
    if [[ -f "$BACKUP_DIR/${APP_NAME}.old" ]]; then
        log_warning "Restoring previous version..."
        cp $BACKUP_DIR/${APP_NAME}.old bin/$APP_NAME
        chmod +x bin/$APP_NAME
        log_info "Previous version restored"
    fi
    exit 1
fi

# Check if service is running
if systemctl is-active --quiet $APP_NAME 2>/dev/null; then
    log_success "Service is running successfully"
elif pgrep -f $APP_NAME > /dev/null; then
    log_success "Application is running successfully"
else
    log_warning "Service may not be running. Check with: sudo systemctl status $APP_NAME"
fi

echo ""
log_success "Update completed successfully!"
echo ""
log_info "Backup created at: $BACKUP_DIR"
log_info "To rollback if needed: cp $BACKUP_DIR/${APP_NAME}.old bin/$APP_NAME"
echo ""
log_info "Useful commands:"
echo "  Check status: sudo systemctl status $APP_NAME"
echo "  View logs: sudo journalctl -u $APP_NAME -f"
echo "  Restart: sudo systemctl restart $APP_NAME"
echo ""
log_warning "If you encounter issues, restore from backup and check logs"