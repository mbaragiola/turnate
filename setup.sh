#!/bin/bash

# Turnate Setup Script
# This script helps set up the development environment for Turnate

echo "🚀 Setting up Turnate development environment..."

# Check if Go is installed
if command -v go >/dev/null 2>&1; then
    echo "✅ Go is already available in PATH"
    echo "   Version: $(go version)"
elif [ -f "$HOME/go/bin/go" ]; then
    echo "✅ Go found in $HOME/go/bin"
    echo "   Version: $($HOME/go/bin/go version)"
    echo "   Adding to PATH for this session..."
    export PATH="$HOME/go/bin:$PATH"
    echo "   Run: export PATH=\"\$HOME/go/bin:\$PATH\" to make permanent"
else
    echo "❌ Go not found. Installing Go 1.23.4..."
    
    # Download and install Go
    cd /tmp
    curl -LO "https://go.dev/dl/go1.23.4.linux-amd64.tar.gz"
    
    if [ -d "$HOME/go" ]; then
        echo "   Backing up existing $HOME/go to $HOME/go.backup"
        mv "$HOME/go" "$HOME/go.backup"
    fi
    
    tar -C "$HOME" -xzf go1.23.4.linux-amd64.tar.gz
    rm go1.23.4.linux-amd64.tar.gz
    
    export PATH="$HOME/go/bin:$PATH"
    echo "✅ Go installed to $HOME/go"
    echo "   Version: $($HOME/go/bin/go version)"
    echo ""
    echo "📝 Add this to your ~/.bashrc or ~/.zshrc:"
    echo "   export PATH=\"\$HOME/go/bin:\$PATH\""
fi

# Go back to project directory
cd "$(dirname "$0")"

# Install dependencies
echo ""
echo "📦 Installing Go dependencies..."
go mod tidy
go mod download

# Build the application
echo ""
echo "🔨 Building Turnate..."
make build

# Check if build succeeded
if [ -f "./bin/turnate" ]; then
    echo ""
    echo "🎉 Setup complete! Turnate is ready to go."
    echo ""
    echo "🚀 To start Turnate:"
    echo "   make run"
    echo "   # or"
    echo "   ./bin/turnate"
    echo ""
    echo "🌐 Then visit: http://localhost:8080"
    echo "👤 Default admin: admin / admin123"
    echo ""
    echo "🛠️  Development commands:"
    echo "   make help      # Show all available commands"
    echo "   make test      # Run tests"
    echo "   make dev       # Run with auto-reload (requires air)"
    echo ""
else
    echo "❌ Build failed. Please check the errors above."
    exit 1
fi