#!/bin/bash

# Telegram Store Hub - Quick Installation Script
# One-command installation for Ubuntu/Debian systems

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🚀 Telegram Store Hub - Quick Install${NC}"
echo "====================================="

# Check OS
if [[ ! -f /etc/os-release ]]; then
    echo -e "${RED}❌ Unsupported operating system${NC}"
    exit 1
fi

source /etc/os-release
if [[ "$ID" != "ubuntu" && "$ID" != "debian" ]]; then
    echo -e "${RED}❌ This quick installer only supports Ubuntu/Debian${NC}"
    echo -e "${YELLOW}Please use the full setup script: ./scripts/setup.sh${NC}"
    exit 1
fi

# Update system
echo -e "${BLUE}📦 Updating system packages...${NC}"
sudo apt-get update -qq

# Install prerequisites
echo -e "${BLUE}📦 Installing prerequisites...${NC}"
sudo apt-get install -y -qq curl wget git build-essential

# Install Go
echo -e "${BLUE}🔧 Installing Go...${NC}"
if ! command -v go &> /dev/null; then
    GO_VERSION="1.22.0"
    wget -q https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz
    rm go${GO_VERSION}.linux-amd64.tar.gz
    
    # Add to PATH
    echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee -a /etc/profile
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    export PATH=$PATH:/usr/local/go/bin
else
    echo -e "${GREEN}✅ Go already installed${NC}"
fi

# Install PostgreSQL
echo -e "${BLUE}🗄️ Installing PostgreSQL...${NC}"
if ! command -v psql &> /dev/null; then
    sudo apt-get install -y -qq postgresql postgresql-contrib
    sudo systemctl enable postgresql
    sudo systemctl start postgresql
else
    echo -e "${GREEN}✅ PostgreSQL already installed${NC}"
fi

# Run full setup
echo -e "${BLUE}🔧 Running full setup...${NC}"
chmod +x scripts/setup.sh
./scripts/setup.sh

echo -e "${GREEN}🎉 Quick installation completed!${NC}"