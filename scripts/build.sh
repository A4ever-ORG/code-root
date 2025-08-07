#!/bin/bash

# Telegram Store Hub - Cross-Platform Build Script
# Builds the application for all supported platforms

set -e

APP_NAME="telegram-store-hub"
VERSION=${VERSION:-"1.0.0"}
BUILD_DIR="builds"

echo "🔨 Building Telegram Store Hub v${VERSION} for all platforms..."

# Clean build directory
rm -rf ${BUILD_DIR}
mkdir -p ${BUILD_DIR}

# Build information
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS="-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.Commit=${COMMIT}"

# Build for different platforms
echo "📦 Building binaries..."

# Linux AMD64
echo "  → Linux (AMD64)"
GOOS=linux GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o ${BUILD_DIR}/${APP_NAME}-linux-amd64 cmd/bot/main.go

# Linux ARM64 (Raspberry Pi)
echo "  → Linux (ARM64)"
GOOS=linux GOARCH=arm64 go build -ldflags="${LDFLAGS}" -o ${BUILD_DIR}/${APP_NAME}-linux-arm64 cmd/bot/main.go

# Windows AMD64
echo "  → Windows (AMD64)"
GOOS=windows GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o ${BUILD_DIR}/${APP_NAME}-windows-amd64.exe cmd/bot/main.go

# macOS AMD64 (Intel)
echo "  → macOS (Intel)"
GOOS=darwin GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o ${BUILD_DIR}/${APP_NAME}-macos-amd64 cmd/bot/main.go

# macOS ARM64 (Apple Silicon)
echo "  → macOS (Apple Silicon)"
GOOS=darwin GOARCH=arm64 go build -ldflags="${LDFLAGS}" -o ${BUILD_DIR}/${APP_NAME}-macos-arm64 cmd/bot/main.go

# Android ARM64 (Termux)
echo "  → Android (ARM64)"
GOOS=linux GOARCH=arm64 go build -ldflags="${LDFLAGS}" -o ${BUILD_DIR}/${APP_NAME}-android-arm64 cmd/bot/main.go

# FreeBSD AMD64 (optional)
echo "  → FreeBSD (AMD64)"
GOOS=freebsd GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o ${BUILD_DIR}/${APP_NAME}-freebsd-amd64 cmd/bot/main.go

echo ""
echo "✅ Build completed successfully!"
echo ""
echo "📋 Generated binaries:"
ls -la ${BUILD_DIR}/

echo ""
echo "🚀 Usage examples:"
echo "  Linux/macOS: ./${BUILD_DIR}/${APP_NAME}-linux-amd64"
echo "  Windows: ${BUILD_DIR}\\${APP_NAME}-windows-amd64.exe"
echo "  Termux: ./${BUILD_DIR}/${APP_NAME}-android-arm64"
echo ""
echo "📝 Don't forget to:"
echo "  1. Configure .env file with your bot token"
echo "  2. Set up PostgreSQL database"
echo "  3. Configure admin chat ID"
echo ""

# Create checksums
echo "🔐 Generating checksums..."
cd ${BUILD_DIR}
sha256sum * > checksums.txt
cd ..

echo "✨ All done! Binaries are in the ${BUILD_DIR}/ directory"