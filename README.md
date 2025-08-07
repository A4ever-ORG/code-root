# 🤖 Telegram Store Hub

**Complete multi-store e-commerce bot management platform for Telegram**

A comprehensive Go-based system for creating and managing multiple Telegram shopping bots. Features a mother bot (CodeRoot) that allows sellers to create their own sub-bots with different subscription plans, built for maximum performance and cross-platform compatibility.

## ✨ Features

### 🏪 Multi-Store Management
- **Mother Bot System**: Central CodeRoot bot manages all sub-bots
- **Individual Store Bots**: Each seller gets their own branded Telegram bot
- **Subscription Plans**: Free (10 products), Pro (200 products), VIP (unlimited)
- **Persian Language Support**: Complete UI in Persian for Iranian market

### 💼 Business Features
- **Product Management**: Full inventory system with images, categories, pricing
- **Order Processing**: Complete order lifecycle with customer management
- **Payment Integration**: Payment verification and subscription management
- **Admin Panel**: Comprehensive administration through Telegram interface
- **Analytics**: Store performance and sales tracking

### 🛠️ Technical Excellence
- **Pure Go Implementation**: Single binary, no dependencies, maximum performance
- **Cross-Platform**: Windows, Linux, macOS, Android (Termux) support
- **PostgreSQL Database**: Robust data storage with GORM integration
- **Concurrent Bot Handling**: Manages multiple bots simultaneously
- **Secure Architecture**: Input validation, XSS protection, proper authentication
- **Production Ready**: Systemd service, logging, monitoring, backups

## 🚀 Quick Installation

### One-Command Setup (Ubuntu/Debian)

```bash
# Clone the repository
git clone <repository-url>
cd telegram-store-hub

# Run quick installation
chmod +x scripts/*.sh
./scripts/quick-install.sh
```

### Complete Setup (All Platforms)

```bash
# Run comprehensive setup script
./scripts/setup.sh
```

**The setup script automatically:**
- ✅ Installs PostgreSQL database
- ✅ Installs Go 1.22+ programming language  
- ✅ Creates database with all tables and indexes
- ✅ Generates secure credentials
- ✅ Downloads dependencies and builds application
- ✅ Creates configuration files and run scripts
- ✅ Sets up system services (Linux)

## 📋 Supported Platforms

| Platform | Installation | Status |
|----------|-------------|---------|
| **Ubuntu 20.04+** | `./scripts/quick-install.sh` | ✅ Full Support |
| **Debian 11+** | `./scripts/quick-install.sh` | ✅ Full Support |
| **CentOS 8+** | `./scripts/setup.sh` | ✅ Full Support |
| **RHEL 8+** | `./scripts/setup.sh` | ✅ Full Support |
| **Fedora 35+** | `./scripts/setup.sh` | ✅ Full Support |
| **macOS 11+** | `./scripts/setup.sh` | ✅ Full Support |
| **Android (Termux)** | `./scripts/install-termux.sh` | ✅ Full Support |
| **Windows 10+** | Manual + WSL | ⚠️ Limited Support |
| **FreeBSD 13+** | Manual | ⚠️ Limited Support |

## 🔧 Configuration

After installation, configure your bot:

### 1. Get Bot Token from @BotFather

1. Message @BotFather on Telegram
2. Create bot with `/newbot`
3. Save the provided token

### 2. Edit Configuration

```bash
nano .env
```

**Required settings:**
```env
# Your bot token from @BotFather
BOT_TOKEN=1234567890:ABCdefGHIjklMNOpqrsTUVwxyz

# Your Telegram chat ID (get from @userinfobot)
ADMIN_CHAT_ID=123456789

# Your channel username
CHANNEL_USERNAME=@your_channel
```

The database settings are automatically configured during installation.

### 3. Start the Bot

```bash
# Run directly
./run.sh

# OR as system service (Linux)
sudo systemctl start telegram-store-hub
sudo systemctl enable telegram-store-hub  # Auto-start on boot
```

## 📊 System Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Mother Bot    │    │   Store Bot 1   │    │   Store Bot N   │
│   (CodeRoot)    │◄───┤   (@shop1_bot)  │    │  (@shopN_bot)   │
│                 │    │                 │    │                 │
│ • User Management│    │ • Products      │    │ • Products      │
│ • Subscriptions │    │ • Orders        │    │ • Orders        │
│ • Sub-bot Creation│   │ • Payments     │    │ • Payments      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
          │                        │                        │
          └────────────────────────┼────────────────────────┘
                                   │
                    ┌─────────────────────────┐
                    │   PostgreSQL Database   │
                    │                         │
                    │ • Users & Stores        │
                    │ • Products & Orders     │
                    │ • Payments & Sessions   │
                    └─────────────────────────┘
```

## 🎯 Subscription Plans

| Plan | Products | Price | Features |
|------|----------|-------|----------|
| **Free** | 10 | Free | Basic store functionality |
| **Pro** | 200 | 50,000 | Advanced features, priority support |
| **VIP** | Unlimited | 150,000 | All features, custom branding |

## 📁 Project Structure

```
telegram-store-hub/
├── cmd/                    # Application entry points
│   ├── bot/               # Main bot application
│   └── web/               # Web dashboard (optional)
├── internal/              # Internal packages
│   ├── bot/               # Bot logic and handlers
│   ├── config/            # Configuration management
│   ├── database/          # Database operations
│   ├── handlers/          # HTTP handlers
│   ├── models/            # Data models
│   └── services/          # Business logic
├── scripts/               # Installation and management scripts
│   ├── setup.sh          # Complete setup script
│   ├── quick-install.sh  # Quick Ubuntu/Debian install
│   ├── install-termux.sh # Android/Termux installer
│   ├── update.sh         # Update script
│   └── uninstall.sh      # Uninstall script
├── client/                # React dashboard (development)
├── server/                # Express API server (development)
├── .env.example          # Configuration template
└── README.md
```

## 🔨 Development

### Prerequisites
- Go 1.22+
- PostgreSQL 13+
- Git

### Building from Source

```bash
# Install dependencies
go mod download

# Build for current platform
make build

# Build for all platforms
make build-all

# Run tests
make test

# Development mode
make dev
```

### Cross-Platform Builds

```bash
# Use the build script
./scripts/build.sh

# Generated binaries:
# - telegram-store-hub-linux-amd64
# - telegram-store-hub-windows-amd64.exe
# - telegram-store-hub-macos-amd64
# - telegram-store-hub-android-arm64
```

## 🛡️ Security Features

- **Input Validation**: All user input validated and sanitized
- **XSS Protection**: HTML/script tag removal
- **SQL Injection Prevention**: GORM ORM with prepared statements
- **Secure Credentials**: Auto-generated strong passwords
- **Rate Limiting**: Protection against spam and abuse
- **Systemd Hardening**: Secure service configuration

## 📈 Monitoring & Logging

### View Logs

```bash
# Application logs
tail -f logs/app.log

# System service logs (Linux)
sudo journalctl -u telegram-store-hub -f

# PostgreSQL logs
sudo tail -f /var/log/postgresql/postgresql-*.log
```

### Service Status

```bash
# Check service status
sudo systemctl status telegram-store-hub

# Restart service
sudo systemctl restart telegram-store-hub

# Enable auto-start
sudo systemctl enable telegram-store-hub
```

## 🔄 Maintenance

### Update Application

```bash
# Run update script
./scripts/update.sh
```

### Backup Data

```bash
# Manual database backup
pg_dump -h localhost -U telegram_hub telegram_store_hub > backup.sql

# Restore from backup
psql -h localhost -U telegram_hub -d telegram_store_hub < backup.sql
```

### Uninstall

```bash
# Uninstall with options to preserve data
./scripts/uninstall.sh
```

## 🐛 Troubleshooting

### Common Issues

**Bot not starting:**
```bash
# Check configuration
cat .env

# Check database connection
psql -h localhost -U telegram_hub -d telegram_store_hub

# Check logs
tail -f logs/app.log
```

**Permission errors:**
```bash
# Fix script permissions
chmod +x scripts/*.sh
chmod +x run.sh
chmod +x bin/telegram-store-hub
```

**Database issues:**
```bash
# Check PostgreSQL status
sudo systemctl status postgresql

# Restart PostgreSQL
sudo systemctl restart postgresql
```

### Getting Help

1. **Check logs**: `tail -f logs/app.log`
2. **Verify config**: `cat .env`
3. **Test database**: Connection test in logs
4. **Service status**: `sudo systemctl status telegram-store-hub`

## 📋 System Requirements

**Minimum:**
- 512MB RAM
- 100MB disk space
- PostgreSQL 10+
- Go 1.22+

**Recommended:**
- 1GB+ RAM
- 1GB+ disk space
- PostgreSQL 13+
- SSD storage
- Monitoring setup

## 🤝 Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 📞 Support

- **Documentation**: Complete installation guide in `INSTALLATION.md`
- **Issues**: Report bugs via GitHub Issues
- **Telegram**: Contact support through the bot itself

---

**Built with ❤️ for the Telegram ecosystem**

*Ready to revolutionize e-commerce on Telegram? Get started in minutes!*