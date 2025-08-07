#!/bin/bash

# Comprehensive Telegram Store Hub Deployment Script
# This script deploys the complete system with all 8 features

set -e

echo "🚀 Starting Telegram Store Hub Deployment..."
echo "================================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging function
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"
    exit 1
}

# Check if running as root
if [[ $EUID -eq 0 ]]; then
   error "This script should not be run as root for security reasons"
fi

# Check system requirements
check_requirements() {
    log "Checking system requirements..."
    
    # Check for required commands
    local required_commands=("go" "node" "npm" "psql" "git")
    
    for cmd in "${required_commands[@]}"; do
        if ! command -v $cmd &> /dev/null; then
            error "$cmd is required but not installed. Please install it first."
        fi
    done
    
    # Check Go version
    local go_version=$(go version | cut -d' ' -f3 | sed 's/go//')
    if ! [[ "$go_version" =~ ^1\.(19|20|21) ]]; then
        warn "Go version $go_version detected. Recommended: 1.19+"
    fi
    
    # Check Node version
    local node_version=$(node --version | sed 's/v//')
    if ! [[ "$node_version" =~ ^(18|20|21) ]]; then
        warn "Node version $node_version detected. Recommended: 18+"
    fi
    
    log "✅ System requirements check completed"
}

# Setup environment
setup_environment() {
    log "Setting up environment..."
    
    # Create .env file if it doesn't exist
    if [[ ! -f .env ]]; then
        log "Creating .env file from template..."
        cp .env.example .env
        
        echo -e "${BLUE}================================================${NC}"
        echo -e "${BLUE}🔧 IMPORTANT: Please configure your .env file!${NC}"
        echo -e "${BLUE}================================================${NC}"
        echo ""
        echo "Required configurations:"
        echo "1. BOT_TOKEN - Get from @BotFather on Telegram"
        echo "2. ADMIN_CHAT_ID - Your Telegram chat ID"
        echo "3. DATABASE_URL - PostgreSQL connection string"
        echo "4. FORCE_JOIN_CHANNEL_ID - Channel for forced join (optional)"
        echo ""
        echo "Press Enter to continue after configuring .env..."
        read -r
    fi
    
    # Source environment variables
    if [[ -f .env ]]; then
        set -a
        source .env
        set +a
    fi
    
    log "✅ Environment setup completed"
}

# Install dependencies
install_dependencies() {
    log "Installing dependencies..."
    
    # Install Go dependencies
    log "Installing Go modules..."
    go mod download
    go mod tidy
    
    # Install Node.js dependencies
    log "Installing Node.js packages..."
    npm install
    
    log "✅ Dependencies installation completed"
}

# Setup database
setup_database() {
    log "Setting up database..."
    
    if [[ -z "$DATABASE_URL" ]]; then
        warn "DATABASE_URL not set. Using PostgreSQL environment variables..."
        
        # Check for PostgreSQL environment variables
        if [[ -z "$PGHOST" || -z "$PGDATABASE" || -z "$PGUSER" ]]; then
            error "Database configuration missing. Please set DATABASE_URL or PostgreSQL environment variables."
        fi
        
        # Construct DATABASE_URL
        local password_part=""
        if [[ -n "$PGPASSWORD" ]]; then
            password_part=":$PGPASSWORD"
        fi
        
        export DATABASE_URL="postgresql://$PGUSER$password_part@$PGHOST:${PGPORT:-5432}/$PGDATABASE"
    fi
    
    # Test database connection
    log "Testing database connection..."
    if ! psql "$DATABASE_URL" -c "SELECT version();" &> /dev/null; then
        error "Cannot connect to database. Please check your DATABASE_URL."
    fi
    
    log "✅ Database setup completed"
}

# Build the application
build_application() {
    log "Building application..."
    
    # Build Go binary
    log "Building Go binary..."
    go build -o bin/telegram-store-hub cmd/main.go
    
    # Build frontend (if applicable)
    log "Building frontend..."
    npm run build
    
    log "✅ Application build completed"
}

# Run tests
run_tests() {
    log "Running tests..."
    
    # Run Go tests
    log "Running Go tests..."
    if ! go test ./tests/... -v; then
        warn "Some Go tests failed. Check the output above."
    fi
    
    # Run integration tests if database is available
    if [[ -n "$DATABASE_URL" ]]; then
        log "Running integration tests..."
        if ! go test ./tests/comprehensive_test.go -v; then
            warn "Some integration tests failed. Check the output above."
        fi
    fi
    
    log "✅ Tests completed"
}

# Create systemd service
create_systemd_service() {
    if [[ "$1" == "--no-service" ]]; then
        log "Skipping systemd service creation"
        return
    fi
    
    log "Creating systemd service..."
    
    local service_file="/tmp/telegram-store-hub.service"
    local current_user=$(whoami)
    local current_dir=$(pwd)
    
    cat > "$service_file" << EOF
[Unit]
Description=Telegram Store Hub - CodeRoot Mother Bot
After=network.target postgresql.service
Wants=postgresql.service

[Service]
Type=simple
User=$current_user
WorkingDirectory=$current_dir
Environment=NODE_ENV=production
EnvironmentFile=$current_dir/.env
ExecStart=$current_dir/bin/telegram-store-hub
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=telegram-store-hub

# Security settings
NoNewPrivileges=yes
PrivateTmp=yes
ProtectSystem=strict
ProtectHome=yes
ReadWritePaths=$current_dir

[Install]
WantedBy=multi-user.target
EOF
    
    echo ""
    echo "Systemd service file created at: $service_file"
    echo ""
    echo "To install the service, run as root:"
    echo "  sudo cp $service_file /etc/systemd/system/"
    echo "  sudo systemctl daemon-reload"
    echo "  sudo systemctl enable telegram-store-hub"
    echo "  sudo systemctl start telegram-store-hub"
    echo ""
    
    log "✅ Systemd service file created"
}

# Start the application
start_application() {
    if [[ "$1" == "--no-start" ]]; then
        log "Skipping application start"
        return
    fi
    
    log "Starting application..."
    
    # Kill any existing process
    if pgrep -f "telegram-store-hub" > /dev/null; then
        log "Stopping existing application..."
        pkill -f "telegram-store-hub" || true
        sleep 2
    fi
    
    # Start in background
    log "Starting Telegram Store Hub..."
    nohup ./bin/telegram-store-hub > logs/app.log 2>&1 &
    local app_pid=$!
    
    # Create logs directory
    mkdir -p logs
    
    echo "Application started with PID: $app_pid"
    echo "Logs: tail -f logs/app.log"
    
    # Wait a moment and check if process is still running
    sleep 3
    if ps -p $app_pid > /dev/null; then
        log "✅ Application started successfully"
    else
        error "Application failed to start. Check logs/app.log for details."
    fi
}

# Generate deployment report
generate_report() {
    log "Generating deployment report..."
    
    local report_file="deployment-report-$(date +%Y%m%d-%H%M%S).txt"
    
    cat > "$report_file" << EOF
Telegram Store Hub Deployment Report
=====================================
Date: $(date)
User: $(whoami)
Directory: $(pwd)

System Information:
- OS: $(uname -s) $(uname -r)
- Go Version: $(go version)
- Node Version: $(node --version)
- NPM Version: $(npm --version)

Application Features:
✅ 1. Mother Bot (CodeRoot) with Persian interface
✅ 2. Forced channel join system
✅ 3. Seller management panel
✅ 4. Three-tier subscription plans (Free/Pro/VIP)
✅ 5. Admin management panel
✅ 6. Automatic sub-bot creation system
✅ 7. Subscription renewal reminder system
✅ 8. Database integration and testing

Configuration:
- Database: $(echo $DATABASE_URL | sed 's/:[^@]*@/:***@/')
- Bot Token: $(echo $BOT_TOKEN | sed 's/.*:/***:/')
- Admin Chat ID: $ADMIN_CHAT_ID

Binary Location: $(pwd)/bin/telegram-store-hub
Logs Location: $(pwd)/logs/app.log

Service Management:
- Start: systemctl start telegram-store-hub
- Stop: systemctl stop telegram-store-hub
- Status: systemctl status telegram-store-hub
- Logs: journalctl -u telegram-store-hub -f

Manual Start:
- Foreground: ./bin/telegram-store-hub
- Background: nohup ./bin/telegram-store-hub > logs/app.log 2>&1 &

Deployment completed successfully!
EOF
    
    echo ""
    echo "📋 Deployment report saved to: $report_file"
    echo ""
    
    log "✅ Deployment report generated"
}

# Main deployment function
main() {
    echo ""
    echo "🎯 Telegram Store Hub - Complete Deployment"
    echo "Features: Mother Bot, Channel Join, Seller Panel, 3-Tier Plans,"
    echo "          Admin Panel, Auto Sub-Bots, Smart Reminders, Full Testing"
    echo ""
    
    # Parse command line arguments
    local no_service=false
    local no_start=false
    local skip_tests=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --no-service)
                no_service=true
                shift
                ;;
            --no-start)
                no_start=true
                shift
                ;;
            --skip-tests)
                skip_tests=true
                shift
                ;;
            --help)
                echo "Usage: $0 [OPTIONS]"
                echo "Options:"
                echo "  --no-service   Skip systemd service creation"
                echo "  --no-start     Skip starting the application"
                echo "  --skip-tests   Skip running tests"
                echo "  --help         Show this help"
                exit 0
                ;;
            *)
                error "Unknown option: $1"
                ;;
        esac
    done
    
    # Run deployment steps
    check_requirements
    setup_environment
    install_dependencies
    setup_database
    build_application
    
    if [[ "$skip_tests" != true ]]; then
        run_tests
    fi
    
    if [[ "$no_service" != true ]]; then
        create_systemd_service
    fi
    
    if [[ "$no_start" != true ]]; then
        start_application
    fi
    
    generate_report
    
    echo ""
    echo "🎉 Deployment completed successfully!"
    echo ""
    echo "🔗 Quick Start:"
    echo "1. Configure your bot token and admin settings in .env"
    echo "2. Set up your Telegram channel for forced join (optional)"
    echo "3. Start using your mother bot: /start"
    echo ""
    echo "📚 Documentation:"
    echo "- README.md - General information"
    echo "- INSTALLATION.md - Detailed installation guide"
    echo "- USAGE.md - Usage instructions"
    echo ""
    echo "🆘 Support:"
    echo "- Check logs: tail -f logs/app.log"
    echo "- Test database: psql \$DATABASE_URL"
    echo "- System status: systemctl status telegram-store-hub"
    echo ""
}

# Run main function with all arguments
main "$@"