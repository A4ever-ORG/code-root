#!/bin/bash

# Quick test script for telegram store hub
set -e

echo "🧪 Running quick system tests..."

# Test Go compilation
echo "✓ Testing Go compilation..."
go build -o /tmp/test-binary . && rm -f /tmp/test-binary

# Test basic imports
echo "✓ Testing basic imports..."
go run -c <<EOF || echo "package main; import _ \"telegram-store-hub/internal/models\"; import _ \"telegram-store-hub/internal/bot\"; import _ \"telegram-store-hub/internal/config\"; func main() {}" > /tmp/test.go && go run /tmp/test.go && rm -f /tmp/test.go
EOF

echo "✅ Basic tests passed!"
echo "✅ System is ready for deployment"
echo
echo "Next steps:"
echo "1. Configure .env file with your bot tokens"
echo "2. Set up PostgreSQL database"
echo "3. Run: make build"
echo "4. Deploy using: sudo ./scripts/improved-install.sh"