#!/bin/bash
#
# agmd installer
# Usage: curl -fsSL https://gluongrid.dev/agmd/install.sh | bash
#

set -e

REPO="GluonGrid/agmd"
BINARY_NAME="agmd"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

info() {
    echo -e "${BLUE}→${NC} $1"
}

success() {
    echo -e "${GREEN}✓${NC} $1"
}

error() {
    echo -e "${RED}✗${NC} $1"
    exit 1
}

# Detect OS and architecture
detect_platform() {
    OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
    ARCH="$(uname -m)"

    case "$OS" in
        linux)
            OS="linux"
            ;;
        darwin)
            OS="darwin"
            ;;
        *)
            error "Unsupported operating system: $OS"
            ;;
    esac

    case "$ARCH" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        *)
            error "Unsupported architecture: $ARCH"
            ;;
    esac

    PLATFORM="${OS}_${ARCH}"
}

# Get latest release version
get_latest_version() {
    VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')

    if [ -z "$VERSION" ]; then
        # Fallback: try to get from tags if no release
        VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/tags" 2>/dev/null | grep '"name"' | head -1 | sed -E 's/.*"([^"]+)".*/\1/')
    fi

    if [ -z "$VERSION" ]; then
        error "Could not determine latest version. Please install manually from https://github.com/${REPO}/releases"
    fi
}

# Download and install
install_binary() {
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY_NAME}_${PLATFORM}.tar.gz"

    info "Downloading agmd ${VERSION} for ${PLATFORM}..."

    # Create temp directory
    TMP_DIR=$(mktemp -d)
    trap "rm -rf $TMP_DIR" EXIT

    # Download
    if ! curl -fsSL "$DOWNLOAD_URL" -o "$TMP_DIR/agmd.tar.gz" 2>/dev/null; then
        # Try without platform suffix (single binary)
        DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY_NAME}"
        if ! curl -fsSL "$DOWNLOAD_URL" -o "$TMP_DIR/agmd" 2>/dev/null; then
            error "Failed to download from ${DOWNLOAD_URL}\nPlease check https://github.com/${REPO}/releases for available downloads"
        fi
        chmod +x "$TMP_DIR/agmd"
    else
        # Extract
        tar -xzf "$TMP_DIR/agmd.tar.gz" -C "$TMP_DIR"
    fi

    # Create install directory if it doesn't exist
    mkdir -p "$INSTALL_DIR"

    # Install
    mv "$TMP_DIR/agmd" "$INSTALL_DIR/agmd"
    chmod +x "$INSTALL_DIR/agmd"

    success "Installed agmd to $INSTALL_DIR/agmd"
}

# Check if install dir is in PATH
check_path() {
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        echo ""
        echo -e "${BLUE}Note:${NC} $INSTALL_DIR is not in your PATH"
        echo ""
        echo "Add it to your shell profile:"
        echo ""
        echo "  # For bash (~/.bashrc or ~/.bash_profile)"
        echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
        echo ""
        echo "  # For zsh (~/.zshrc)"
        echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
        echo ""
        echo "  # For fish (~/.config/fish/config.fish)"
        echo "  set -gx PATH \$HOME/.local/bin \$PATH"
        echo ""
    fi
}

# Alternative: install from source
install_from_source() {
    info "Installing from source..."

    if ! command -v go &> /dev/null; then
        error "Go is required for source installation. Install Go from https://go.dev"
    fi

    go install "github.com/${REPO}@latest"
    success "Installed agmd via 'go install'"
}

main() {
    echo ""
    echo "  agmd installer"
    echo "  https://github.com/${REPO}"
    echo ""

    detect_platform
    info "Detected platform: ${PLATFORM}"

    # Check for --source flag
    if [ "$1" = "--source" ]; then
        install_from_source
    else
        get_latest_version
        info "Latest version: ${VERSION}"
        install_binary
    fi

    check_path

    echo ""
    success "Installation complete!"
    echo ""
    echo "Get started:"
    echo "  agmd setup    # Initialize your registry"
    echo "  agmd init     # Create directives.md in a project"
    echo "  agmd --help   # See all commands"
    echo ""
}

main "$@"
