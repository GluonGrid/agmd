#!/bin/bash
# Setup script for agent-config on M4 MacBook Pro
# This script automates the complete installation process

set -e # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
print_step() {
    echo -e "${BLUE}==>${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# Check if running on macOS
if [[ "$OSTYPE" != "darwin"* ]]; then
    print_error "This script is designed for macOS. Detected: $OSTYPE"
    exit 1
fi

# Check if running on ARM64 (M4)
ARCH=$(uname -m)
if [[ "$ARCH" != "arm64" ]]; then
    print_warning "Expected ARM64 (M4 Mac), detected: $ARCH"
    print_warning "This might still work, but the script is optimized for M4."
fi

echo ""
echo "╔════════════════════════════════════════════════╗"
echo "║   agent-config Setup for M4 MacBook Pro       ║"
echo "╚════════════════════════════════════════════════╝"
echo ""

# Step 1: Check/Install Homebrew
print_step "Checking Homebrew..."
if ! command -v brew &>/dev/null; then
    print_warning "Homebrew not found. Installing..."
    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

    # Add Homebrew to PATH for M-series Macs
    echo 'eval "$(/opt/homebrew/bin/brew shellenv)"' >>~/.zprofile
    eval "$(/opt/homebrew/bin/brew shellenv)"

    print_success "Homebrew installed"
else
    print_success "Homebrew already installed"
fi

# Step 2: Check/Install Go
print_step "Checking Go installation..."
if ! command -v go &>/dev/null; then
    print_warning "Go not found. Installing..."
    brew install go
    print_success "Go installed"
else
    GO_VERSION=$(go version | awk '{print $3}')
    print_success "Go already installed: $GO_VERSION"
fi

# Verify Go architecture
GO_ARCH=$(go env GOARCH)
if [[ "$GO_ARCH" != "arm64" ]]; then
    print_warning "Go is not configured for ARM64. Detected: $GO_ARCH"
fi

# Step 3: Setup Go environment
print_step "Setting up Go environment..."
GOPATH=$(go env GOPATH)
if ! grep -q "GOPATH" ~/.zshrc 2>/dev/null; then
    echo "" >>~/.zshrc
    echo "# Go environment" >>~/.zshrc
    echo "export GOPATH=\$HOME/go" >>~/.zshrc
    echo "export PATH=\$PATH:/usr/local/go/bin:\$GOPATH/bin" >>~/.zshrc
    print_success "Added Go to ~/.zshrc"
else
    print_success "Go environment already configured"
fi

# Source the updated shell config
export GOPATH=$HOME/go
export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin

# Step 4: Navigate to project directory
print_step "Setting up project..."
PROJECT_DIR=$(pwd)
if [[ ! -f "$PROJECT_DIR/go.mod" ]]; then
    print_error "go.mod not found. Are you in the agent-config directory?"
    exit 1
fi
print_success "Found project at: $PROJECT_DIR"

# Step 5: Download dependencies
print_step "Downloading Go dependencies..."
go mod download
print_success "Dependencies downloaded"

# Step 6: Verify and clean dependencies
print_step "Cleaning up dependencies..."
go mod tidy
print_success "Dependencies cleaned"

# Step 7: Build the project
print_step "Building agent-config for ARM64..."
go build -o agent-config .
if [[ -f "agent-config" ]]; then
    print_success "Binary built successfully"

    # Show binary info
    SIZE=$(ls -lh agent-config | awk '{print $5}')
    print_success "Binary size: $SIZE"
else
    print_error "Build failed"
    exit 1
fi

# Step 8: Test the binary
print_step "Testing binary..."
if ./agent-config --help &>/dev/null; then
    print_success "Binary works correctly"
else
    print_error "Binary test failed"
    exit 1
fi

# Step 9: Install globally
print_step "Installing to GOPATH/bin..."
go install .
if command -v agent-config &>/dev/null; then
    print_success "agent-config installed globally"
    INSTALL_PATH=$(which agent-config)
    print_success "Installed at: $INSTALL_PATH"
else
    print_warning "Installation completed but agent-config not in PATH"
    print_warning "You may need to restart your terminal or run: source ~/.zshrc"
fi

# Step 10: Optional - Install dev tools
echo ""
read -p "Install Go development tools? (gopls, golangci-lint, delve) [y/N] " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    print_step "Installing Go development tools..."

    print_step "Installing gopls (language server)..."
    go install golang.org/x/tools/gopls@latest

    print_step "Installing golangci-lint (linter)..."
    brew install golangci-lint

    print_step "Installing delve (debugger)..."
    go install github.com/go-delve/delve/cmd/dlv@latest

    print_success "Development tools installed"
fi

# Final summary
echo ""
echo "╔════════════════════════════════════════════════╗"
echo "║              Setup Complete! 🎉                ║"
echo "╚════════════════════════════════════════════════╝"
echo ""
print_success "Go version: $(go version)"
print_success "Architecture: $(go env GOARCH)"
print_success "Project built successfully"
print_success "agent-config installed globally"
echo ""
echo "Next steps:"
echo "  1. Restart your terminal or run: source ~/.zshrc"
echo "  2. Test: agent-config --help"
echo "  3. Try it in a project: agent-config init --claude"
echo ""
echo "Quick start:"
echo "  cd ~/your-project"
echo "  agent-config init --claude"
echo "  agent-config add \"Your first rule\""
echo "  agent-config list"
echo ""
print_success "Happy coding! 🚀"
