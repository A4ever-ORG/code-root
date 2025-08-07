# Telegram Store Hub - Deployment Guide

## Quick Deployment (Recommended)

### 1. Prerequisites
- Linux server (Ubuntu 18.04+, CentOS 7+)
- Go 1.22+ installed
- PostgreSQL 12+ installed
- Root/sudo access

### 2. Build the Application
```bash
# Clone or download the project
git clone <repository_url>
cd telegram-store-hub

# Install dependencies
make deps

# Build for your platform
make build
```

### 3. Deploy Using Improved Installation Script
```bash
# Run the enhanced installation script
sudo ./scripts/improved-install.sh

# The script will:
# - Create system user and directories
# - Install the binary
# - Set up systemd service
# - Configure logging
# - Set up database (if PostgreSQL is available)
# - Create uninstall script
```

### 4. Configure Your Bot
```bash
# Edit the configuration file
sudo nano /etc/telegram-store-hub/telegram-store-hub.env

# Required settings:
BOT_TOKEN=your_telegram_bot_token_here
ADMIN_CHAT_ID=your_admin_telegram_id
DATABASE_URL=postgres://user:password@localhost:5432/database
```

### 5. Start the Service
```bash
# Start the bot service
sudo systemctl start telegram-store-hub

# Enable auto-start on boot
sudo systemctl enable telegram-store-hub

# Check status
sudo systemctl status telegram-store-hub

# View logs
sudo journalctl -u telegram-store-hub -f
```

## Manual Deployment

### 1. System Setup
```bash
# Create user
sudo useradd --system --no-create-home telegram-hub

# Create directories
sudo mkdir -p /etc/telegram-store-hub
sudo mkdir -p /var/log/telegram-store-hub
sudo mkdir -p /var/lib/telegram-store-hub
```

### 2. Database Setup
```bash
# Connect to PostgreSQL
sudo -u postgres psql

# Create database and user
CREATE USER telegram_hub WITH PASSWORD 'secure_password';
CREATE DATABASE telegram_store_hub OWNER telegram_hub;
GRANT ALL PRIVILEGES ON DATABASE telegram_store_hub TO telegram_hub;
\q
```

### 3. Install Binary
```bash
# Build the application
make build

# Install binary
sudo cp build/telegram-store-hub /usr/local/bin/
sudo chmod +x /usr/local/bin/telegram-store-hub
```

### 4. Configuration
```bash
# Copy configuration
sudo cp .env.example /etc/telegram-store-hub/telegram-store-hub.env
sudo chown root:telegram-hub /etc/telegram-store-hub/telegram-store-hub.env
sudo chmod 640 /etc/telegram-store-hub/telegram-store-hub.env

# Edit configuration
sudo nano /etc/telegram-store-hub/telegram-store-hub.env
```

### 5. Systemd Service
```bash
# Create service file
sudo tee /etc/systemd/system/telegram-store-hub.service > /dev/null <<EOF
[Unit]
Description=Telegram Store Hub Bot
After=network.target postgresql.service

[Service]
Type=simple
User=telegram-hub
WorkingDirectory=/var/lib/telegram-store-hub
EnvironmentFile=/etc/telegram-store-hub/telegram-store-hub.env
ExecStart=/usr/local/bin/telegram-store-hub
Restart=always

[Install]
WantedBy=multi-user.target
EOF

# Reload and start
sudo systemctl daemon-reload
sudo systemctl enable telegram-store-hub
sudo systemctl start telegram-store-hub
```

## Testing Deployment

### 1. System Health Check
```bash
# Check service status
sudo systemctl status telegram-store-hub

# Check logs
sudo journalctl -u telegram-store-hub --no-pager -l

# Test database connection
sudo -u telegram-hub psql $DATABASE_URL -c "SELECT 1;"
```

### 2. Bot Functionality Test
```bash
# Send /start to your bot in Telegram
# Verify welcome message appears
# Test menu buttons
# Try creating a store with free plan
```

### 3. Performance Test
```bash
# Monitor resource usage
sudo htop

# Check memory usage
sudo systemctl show --property=MemoryUsage telegram-store-hub

# Monitor logs for errors
sudo journalctl -u telegram-store-hub -f
```

## Cross-Platform Builds

### Build for Multiple Platforms
```bash
# Build for Linux
make build-linux

# Build for Windows
make build-windows

# Build for macOS
make build-mac

# Build for ARM (Raspberry Pi, Android Termux)
make build-arm

# Build all platforms
make build-all
```

### Docker Deployment
```bash
# Build Docker image
make docker-build

# Run with Docker
make docker-run

# Or use docker-compose (create docker-compose.yml first)
docker-compose up -d
```

## Monitoring and Maintenance

### Log Management
```bash
# View real-time logs
sudo journalctl -u telegram-store-hub -f

# View last 100 lines
sudo journalctl -u telegram-store-hub -n 100

# View logs from today
sudo journalctl -u telegram-store-hub --since today
```

### Backup
```bash
# Backup database
pg_dump $DATABASE_URL > backup_$(date +%Y%m%d).sql

# Backup configuration
sudo cp /etc/telegram-store-hub/telegram-store-hub.env backup_config_$(date +%Y%m%d).env
```

### Updates
```bash
# Stop service
sudo systemctl stop telegram-store-hub

# Backup current binary
sudo cp /usr/local/bin/telegram-store-hub /usr/local/bin/telegram-store-hub.backup

# Install new binary
sudo cp build/telegram-store-hub /usr/local/bin/

# Start service
sudo systemctl start telegram-store-hub
```

### Uninstall
```bash
# Use the uninstall script created during installation
sudo /usr/local/bin/telegram-store-hub-uninstall

# Or manual cleanup:
sudo systemctl stop telegram-store-hub
sudo systemctl disable telegram-store-hub
sudo rm /etc/systemd/system/telegram-store-hub.service
sudo rm /usr/local/bin/telegram-store-hub
sudo rm -rf /etc/telegram-store-hub
sudo rm -rf /var/log/telegram-store-hub
sudo rm -rf /var/lib/telegram-store-hub
sudo userdel telegram-hub
```

## Troubleshooting

### Common Issues

1. **Bot not responding**
   - Check bot token in configuration
   - Verify internet connectivity
   - Check service logs for errors

2. **Database connection failed**
   - Verify PostgreSQL is running
   - Check database URL format
   - Ensure user has proper permissions

3. **Permission denied errors**
   - Check file ownership and permissions
   - Verify user has access to directories
   - Check SELinux/AppArmor policies

4. **Service won't start**
   - Check systemd service file syntax
   - Verify binary permissions
   - Check configuration file format

### Getting Help
- Check logs: `sudo journalctl -u telegram-store-hub`
- Test configuration: `telegram-store-hub --test-config`
- Validate bot token: Use Telegram Bot API
- Community support: [GitHub Issues](link-to-repository)

## Security Considerations

1. **File Permissions**
   - Configuration files: 640 (root:telegram-hub)
   - Binary: 755 (root:root)
   - Data directories: 750 (telegram-hub:telegram-hub)

2. **Network Security**
   - Use HTTPS for webhooks
   - Configure firewall rules
   - Use VPN for database access

3. **Bot Token Security**
   - Store securely in configuration
   - Never commit to version control
   - Rotate tokens periodically

4. **Database Security**
   - Use strong passwords
   - Enable SSL connections
   - Regular backups
   - Access logging