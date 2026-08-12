#!/bin/bash
# MoAI-ADK Go Edition Installer
# This script detects your platform and downloads the appropriate binary

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print colored message
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Detect OS and architecture
detect_platform() {
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    local arch=$(uname -m)

    case $os in
        darwin)
            OS="darwin"
            ;;
        linux)
            OS="linux"
            # Detect WSL
            if [ -n "$WSL_DISTRO_NAME" ]; then
                print_info "Detected WSL environment: $WSL_DISTRO_NAME"
            fi
            ;;
        *)
            print_error "Unsupported OS: $os"
            print_info "Supported operating systems: macOS, Linux"
            exit 1
            ;;
    esac

    case $arch in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        *)
            print_error "Unsupported architecture: $arch"
            print_info "Supported architectures: x86_64 (amd64), arm64"
            exit 1
            ;;
    esac

    PLATFORM="${OS}_${ARCH}"
    print_success "Detected platform: $PLATFORM"
}

# Get latest Go edition version from GitHub
get_latest_version() {
    local version_url="https://api.github.com/repos/modu-ai/moai-adk/releases"

    if command -v curl &> /dev/null; then
        # Try go-v* tags first, then fall back to v* tags
        VERSION=$(curl -s "$version_url" | grep -o '"tag_name":\s*"[^"]*"' | head -n 1 | sed -E 's/.*"([^"]+)".*/\1/' | sed 's/^go-//' | sed 's/^v//')
    elif command -v wget &> /dev/null; then
        VERSION=$(wget -qO- "$version_url" | grep -o '"tag_name":\s*"[^"]*"' | head -n 1 | sed -E 's/.*"([^"]+)".*/\1/' | sed 's/^go-//' | sed 's/^v//')
    else
        print_error "Neither curl nor wget found. Please install one of them."
        exit 1
    fi

    if [ -z "$VERSION" ]; then
        print_error "Failed to fetch latest Go edition version from GitHub"
        print_info "No releases found. You can:"
        echo "  1. Install a specific version: $0 --version 2.0.0"
        echo "  2. Install from source: go install github.com/modu-ai/moai-adk/cmd/moai@latest"
        exit 1
    fi

    print_success "Latest Go edition version: $VERSION"
}

# Download binary
download_binary() {
    local version=$1
    local os_arch=$2

    # Extract OS and ARCH from platform (e.g., "linux_amd64")
    local os=$(echo "$os_arch" | cut -d'_' -f1)
    local arch=$(echo "$os_arch" | cut -d'_' -f2)

    # Determine archive extension based on OS
    local ext="tar.gz"
    if [ "$os" = "windows" ]; then
        ext="zip"
    fi

    # Build archive filename matching goreleaser format
    local archive_name="moai-adk_${version}_${os}_${arch}.${ext}"
    local download_url="https://github.com/modu-ai/moai-adk/releases/download/v${version}/${archive_name}"
    local checksum_url="https://github.com/modu-ai/moai-adk/releases/download/v${version}/checksums.txt"

    # Create temp directory
    TMP_DIR=$(mktemp -d)
    ARCHIVE_FILE="$TMP_DIR/$archive_name"
    CHECKSUM_FILE="$TMP_DIR/checksums.txt"

    print_info "Downloading from: $download_url"

    # Download archive
    if command -v curl &> /dev/null; then
        if ! curl -fsSL "$download_url" -o "$ARCHIVE_FILE"; then
            print_error "Download failed"
            rm -rf "$TMP_DIR"
            exit 1
        fi
        # Download checksums
        if ! curl -fsSL "$checksum_url" -o "$CHECKSUM_FILE"; then
            print_warning "Failed to download checksums (verification skipped)"
        fi
    elif command -v wget &> /dev/null; then
        if ! wget -q "$download_url" -O "$ARCHIVE_FILE"; then
            print_error "Download failed"
            rm -rf "$TMP_DIR"
            exit 1
        fi
        # Download checksums
        if ! wget -q "$checksum_url" -O "$CHECKSUM_FILE"; then
            print_warning "Failed to download checksums (verification skipped)"
        fi
    else
        print_error "Neither curl nor wget found. Please install one of them."
        rm -rf "$TMP_DIR"
        exit 1
    fi

    print_success "Download completed"

    # Verify checksum if checksums.txt was downloaded
    if [ -f "$CHECKSUM_FILE" ]; then
        print_info "Verifying checksum..."
        local expected_checksum=$(grep "$archive_name" "$CHECKSUM_FILE" | awk '{print $1}')

        if [ -n "$expected_checksum" ]; then
            if command -v sha256sum &> /dev/null; then
                local actual_checksum=$(sha256sum "$ARCHIVE_FILE" | awk '{print $1}')
            elif command -v shasum &> /dev/null; then
                local actual_checksum=$(shasum -a 256 "$ARCHIVE_FILE" | awk '{print $1}')
            else
                print_warning "sha256sum/shasum not found (checksum verification skipped)"
            fi

            if [ -n "$actual_checksum" ]; then
                if [ "$expected_checksum" = "$actual_checksum" ]; then
                    print_success "Checksum verified"
                else
                    print_error "Checksum mismatch!"
                    print_error "Expected: $expected_checksum"
                    print_error "Actual:   $actual_checksum"
                    rm -rf "$TMP_DIR"
                    exit 1
                fi
            fi
        fi
    fi

    # Extract archive
    print_info "Extracting archive..."
    if tar -xzf "$ARCHIVE_FILE" -C "$TMP_DIR"; then
        print_success "Extraction completed"
    else
        print_error "Failed to extract archive"
        rm -rf "$TMP_DIR"
        exit 1
    fi

    # Find the binary
    BINARY_PATH="$TMP_DIR/moai"
    if [ ! -f "$BINARY_PATH" ]; then
        print_error "Binary not found in archive"
        rm -rf "$TMP_DIR"
        exit 1
    fi

    # Make executable
    chmod +x "$BINARY_PATH"

    # Install to target location
    install_binary "$BINARY_PATH"
}

# Install binary
install_binary() {
    local binary_path=$1

    # Determine install location
    if [ -n "$INSTALL_DIR" ]; then
        TARGET_DIR="$INSTALL_DIR"
    else
        # Check if Go bin path exists
        if command -v go &> /dev/null; then
            GOBIN=$(go env GOBIN)
            GOPATH=$(go env GOPATH)

            if [ -n "$GOBIN" ] && [ -d "$GOBIN" ]; then
                TARGET_DIR="$GOBIN"
            elif [ -n "$GOPATH" ] && [ -d "$GOPATH/bin" ]; then
                TARGET_DIR="$GOPATH/bin"
            else
                # Default to user's local bin
                TARGET_DIR="$HOME/.local/bin"
            fi
        else
            # Default to user's local bin
            TARGET_DIR="$HOME/.local/bin"
        fi
    fi

    # Create target directory if it doesn't exist
    if [ ! -d "$TARGET_DIR" ]; then
        print_info "Creating directory: $TARGET_DIR"
        mkdir -p "$TARGET_DIR"
    fi

    TARGET_PATH="$TARGET_DIR/moai"

    # Move binary to target location
    if mv "$binary_path" "$TARGET_PATH"; then
        print_success "Installed to: $TARGET_PATH"
    else
        # If mv fails, try cp
        if cp "$binary_path" "$TARGET_PATH"; then
            chmod +x "$TARGET_PATH"
            print_success "Installed to: $TARGET_PATH"
        else
            print_error "Failed to install binary to $TARGET_PATH"
            rm -rf "$(dirname "$binary_path")"
            exit 1
        fi
    fi

    # Clean up temp directory
    rm -rf "$(dirname "$binary_path")"
}

# Verify installation
verify_installation() {
    if command -v moai &> /dev/null; then
        print_success "MoAI-ADK installed successfully!"
        echo ""
        moai version
        echo ""
        print_info "To get started, run:"
        echo "  moai init          # Initialize a new project"
        echo "  moai doctor        # Check system health"
        echo "  moai update --project # Update project templates"
    else
        print_warning "Installation completed, but 'moai' command not found in PATH"
        print_info "Add the following to your ~/.bashrc or ~/.zshrc:"
        echo ""
        if [ -n "$TARGET_DIR" ]; then
            echo "  export PATH=\"\$PATH:$TARGET_DIR\""
        else
            echo "  export PATH=\"\$PATH:\$HOME/.local/bin\""
        fi
        echo ""
        print_info "Then run: source ~/.bashrc (or source ~/.zshrc)"
    fi
}

# Main installation flow
main() {
    echo ""
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║          MoAI's Agentic Development Kit Installer           ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo ""

    # Parse arguments
    VERSION=""
    INSTALL_DIR=""

    while [[ $# -gt 0 ]]; do
        case $1 in
            --version)
                VERSION="$2"
                shift 2
                ;;
            --install-dir)
                INSTALL_DIR="$2"
                shift 2
                ;;
            -h|--help)
                echo "Usage: $0 [OPTIONS]"
                echo ""
                echo "Options:"
                echo "  --version VERSION    Install specific version (default: latest)"
                echo "  --install-dir DIR     Install to custom directory"
                echo "  -h, --help            Show this help message"
                echo ""
                echo "Examples:"
                echo "  $0                                    # Install latest version"
                echo "  $0 --version 2.0.0                   # Install version 2.0.0"
                echo "  $0 --install-dir /usr/local/bin       # Install to custom directory"
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                echo "Use --help for usage information"
                exit 1
                ;;
        esac
    done

    # Detect platform
    detect_platform

    # Get version
    if [ -z "$VERSION" ]; then
        get_latest_version
    else
        print_info "Installing version: $VERSION"
    fi

    # Download and install
    download_binary "$VERSION" "$PLATFORM"

    # Verify installation
    verify_installation

    echo ""
    print_success "Installation complete!"
    echo ""
    print_info "Documentation: https://github.com/modu-ai/moai-adk"
}

# Run main function
main "$@"
