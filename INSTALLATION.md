# Installation Guide - Telegram Store Hub

Complete installation guide for setting up the Telegram Store Hub bot system on various platforms.

## Quick Start (Ubuntu/Debian)

For Ubuntu/Debian systems, use the quick installer:

```bash
# Make scripts executable
chmod +x scripts/*.sh

# Run quick installation (Ubuntu/Debian only)
./scripts/quick-install.sh
```

## Complete Installation (All Platforms)

For all other platforms or custom installations:

```bash
# Run the comprehensive setup script
./scripts/setup.sh
```

This script will:
- ✅ Auto-detect your operating system
- ✅ Install PostgreSQL database
- ✅ Install Go programming language (1.22+)
- ✅ Install system dependencies (git, make, gcc, etc.)
- ✅ Create and configure database with all tables
- ✅ Generate secure database credentials
- ✅ Download Go modules and dependencies
- ✅ Build the application binary
- ✅ Create configuration files
- ✅ Set up run scripts and system services
- ✅ Configure proper permissions and security

## Platform-Specific Installation

### Android (Termux)

For Android devices using Termux:

```bash
# Install in Termux environment
./scripts/install-termux.sh
```

### Manual Installation

If automatic installation fails, follow these manual steps:

#### 1. Install Prerequisites

**Ubuntu/Debian:**
```bash
sudo apt-get update
sudo apt-get install -y postgresql postgresql-contrib git build-essential wget curl
```

**CentOS/RHEL/Fedora:**
```bash
# Fedora
sudo dnf install -y postgresql postgresql-server git make gcc wget curl
# OR CentOS/RHEL
sudo yum install -y postgresql postgresql-server git make gcc wget curl
```

**macOS:**
```bash
# Install Homebrew first: https://brew.sh
brew install postgresql go git make
```

**Termux (Android):**
```bash
pkg update
pkg install -y postgresql golang git make clang
```

#### 2. Install Go (if not installed)

```bash
# Download Go 1.22
wget https://golang.org/dl/go1.22.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz

# Add to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

#### 3. Setup PostgreSQL

```bash
# Start PostgreSQL service
sudo systemctl enable postgresql
sudo systemctl start postgresql

# Create database and user
sudo -u postgres createuser telegram_hub
sudo -u postgres createdb telegram_store_hub -O telegram_hub
sudo -u postgres psql -c "ALTER USER telegram_hub WITH ENCRYPTED PASSWORD 'your_secure_password';"
```

#### 4. Build Application

```bash
# Clone or ensure you're in the project directory
cd telegram-store-hub

# Initialize Go module (if needed)
go mod init telegram-store-hub
go mod tidy

# Build the application
make build
# OR
go build -o bin/telegram-store-hub .
```

#### 5. Configure Environment

```bash
# Copy environment template
cp .env.example .env

# Edit configuration
nano .env
```

## Configuration

After installation, you need to configure the bot:

### 1. Get Telegram Bot Token

1. Message @BotFather on Telegram
2. Create a new bot with `/newbot`
3. Save the bot token

### 2. Edit Configuration File

Edit the `.env` file with your settings:

```bash
nano .env
```

**Required Settings:**
```env
BOT_TOKEN=your_bot_token_from_botfather
ADMIN_CHAT_ID=your_telegram_chat_id
CHANNEL_USERNAME=@your_channel
DATABASE_URL=postgres://user:pass@localhost:5432/dbname
```

**To find your chat ID:**
1. Message @userinfobot on Telegram
2. It will reply with your chat ID

### 3. Database Tables

The setup script automatically creates these tables:
- `users` - User accounts and admin status
- `stores` - Bot stores and subscription plans
- `products` - Store products and inventory
- `orders` - Customer orders and status
- `payments` - Payment tracking and verification
- `subscriptions` - Plan subscriptions and renewals
- `bot_sessions` - Active bot sessions

## Running the Bot

### Development Mode

```bash
# Run directly
./run.sh

# OR using make
make dev

# OR with Go
go run .
```

### Production Mode (Linux)

```bash
# Enable and start systemd service
sudo systemctl enable telegram-store-hub
sudo systemctl start telegram-store-hub

# Check status
sudo systemctl status telegram-store-hub

# View logs
sudo journalctl -u telegram-store-hub -f
```

### Background Mode (Termux)

```bash
# Install screen or tmux
pkg install -y screen

# Run in background
screen -S telegram-bot ./run.sh

# Detach: Ctrl+A, then D
# Reattach: screen -r telegram-bot
```

## Building for Different Platforms

### Cross-Platform Build

```bash
# Build for all platforms
make build-all

# OR use the build script
./scripts/build.sh
```

**Generated binaries:**
- `telegram-store-hub-linux-amd64` - Linux 64-bit
- `telegram-store-hub-linux-arm64` - Linux ARM64 (Raspberry Pi)
- `telegram-store-hub-windows-amd64.exe` - Windows 64-bit
- `telegram-store-hub-macos-amd64` - macOS Intel
- `telegram-store-hub-macos-arm64` - macOS Apple Silicon
- `telegram-store-hub-android-arm64` - Android/Termux

### Docker Installation

```bash
# Build Docker image
make docker-build

# Run with Docker
make docker-run
```

## Directory Structure After Installation

```
telegram-store-hub/
├── bin/                    # Compiled binaries
│   └── telegram-store-hub
├── logs/                   # Application logs
├── scripts/                # Installation scripts
├── internal/               # Go source code
├── .env                    # Configuration (create from .env.example)
├── .db_password           # Database password (auto-generated)
├── run.sh                 # Run script
└── README.md
```

## Troubleshooting

### Common Issues

**"Permission denied" errors:**
```bash
chmod +x scripts/*.sh
chmod +x run.sh
chmod +x bin/telegram-store-hub
```

**Database connection failed:**
```bash
# Check PostgreSQL status
sudo systemctl status postgresql

# Test connection manually
psql -h localhost -U telegram_hub -d telegram_store_hub
```

**Go not found:**
```bash
# Add Go to PATH
export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
```

**Bot not responding:**
1. Check bot token in `.env`
2. Ensure bot is not already running elsewhere
3. Check network connectivity
4. Verify database connection

### Log Files

- Application logs: `logs/app.log`
- System service logs: `sudo journalctl -u telegram-store-hub`
- PostgreSQL logs: `/var/log/postgresql/`

### Database Management

```bash
# Connect to database
psql -h localhost -U telegram_hub -d telegram_store_hub

# Backup database
pg_dump -h localhost -U telegram_hub telegram_store_hub > backup.sql

# Restore database
psql -h localhost -U telegram_hub -d telegram_store_hub < backup.sql
```

## Security Notes

1. **Keep your `.env` file secure** - Never commit it to git
2. **Database password** is auto-generated and saved in `.db_password`
3. **Change default admin chat ID** in configuration
4. **Configure firewall** appropriately for your setup
5. **Regular backups** of database recommended

## Getting Help

If you encounter issues:

1. Check the logs: `tail -f logs/app.log`
2. Verify configuration: `cat .env`
3. Test database: `psql -h localhost -U telegram_hub -d telegram_store_hub`
4. Check service status: `sudo systemctl status telegram-store-hub`

## System Requirements

**Minimum:**
- 512MB RAM
- 100MB disk space
- PostgreSQL 10+
- Go 1.22+

**Recommended:**
- 1GB+ RAM
- 1GB+ disk space (for logs and database)
- PostgreSQL 13+
- Go 1.22+
- SSD storage for better performance

## Supported Platforms

- ✅ Ubuntu 20.04+
- ✅ Debian 11+
- ✅ CentOS 8+
- ✅ RHEL 8+
- ✅ Fedora 35+
- ✅ macOS 11+ (Intel & Apple Silicon)
- ✅ Android (Termux)
- ✅ Windows 10+ (with WSL recommended)
- ✅ FreeBSD 13+

Installation completed successfully! Your Telegram Store Hub is ready to manage multiple bot stores.