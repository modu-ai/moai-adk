#!/usr/bin/env bash
# MoAI-ADK Installation Script
# Automatically detects platform and installs the appropriate binary
#
# Usage:
#   curl -sSL https://raw.githubusercontent.com/anthropics/moai-adk-go/main/scripts/install.sh | bash
#   curl -sSL https://raw.githubusercontent.com/anthropics/moai-adk-go/main/scripts/install.sh | bash -s -- --version 1.0.0

set -e

VERSION="${VERSION:-latest}"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
BINARY_NAME="moai-adk"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

# Detect OS and Architecture
detect_platform() {
    local os=""
    local arch=""

    case "$(uname -s)" in
        Darwin)
            os="darwin"
            ;;
        Linux)
            os="linux"
            ;;
        MINGW*|MSYS*|CYGWIN*)
            os="windows"
            ;;
        *)
            error "Unsupported OS: $(uname -s)"
            exit 1
            ;;
    esac

    case "$(uname -m)" in
        x86_64|amd64)
            arch="amd64"
            ;;
        aarch64|arm64)
            arch="arm64"
            ;;
        *)
            error "Unsupported architecture: $(uname -m)"
            exit 1
            ;;
    esac

    echo "${os}-${arch}"
}

# Download binary from GitHub Releases
download_binary() {
    local platform="$1"
    local version="$2"
    local download_url=""

    if [ "$version" = "latest" ]; then
        download_url="https://github.com/anthropics/moai-adk-go/releases/latest/download/moai-adk-${platform}"
    else
        download_url="https://github.com/anthropics/moai-adk-go/releases/download/${version}/moai-adk-${platform}"
    fi

    info "Downloading moai-adk from ${download_url}"

    if [ "$(uname -s)" = "MINGW*" ] || [ "$(uname -s)" = "MSYS*" ] || [ "$(uname -s)" = "CYGWIN*" ]; then
        # Windows: add .exe extension
        download_url="${download_url}.exe"
        curl -fsSL "${download_url}" -o "${BINARY_NAME}.exe" || {
            error "Failed to download binary"
            exit 1
        }
    else
        curl -fsSL "${download_url}" -o "${BINARY_NAME}" || {
            error "Failed to download binary"
            exit 1
        }
    fi
}

# Verify binary
verify_binary() {
    local binary="$1"

    info "Verifying binary..."

    if [ "$(uname -s)" = "MINGW*" ] || [ "$(uname -s)" = "MSYS*" ] || [ "$(uname -s)" = "CYGWIN*" ]; then
        [ -f "${binary}.exe" ] || {
            error "Binary not found: ${binary}.exe"
            exit 1
        }
    else
        [ -f "${binary}" ] || {
            error "Binary not found: ${binary}"
            exit 1
        }
    fi
}

# Install binary
install_binary() {
    local binary="$1"
    local install_dir="$2"

    info "Installing ${BINARY_NAME} to ${install_dir}"

    # Create install directory if it doesn't exist
    mkdir -p "${install_dir}" || {
        error "Failed to create install directory: ${install_dir}"
        exit 1
    }

    # Move binary to install directory
    if [ "$(uname -s)" = "MINGW*" ] || [ "$(uname -s)" = "MSYS*" ] || [ "$(uname -s)" = "CYGWIN*" ]; then
        mv "${binary}.exe" "${install_dir}/${BINARY_NAME}.exe" || {
            error "Failed to move binary to ${install_dir}"
            exit 1
        }
        chmod +x "${install_dir}/${BINARY_NAME}.exe"
    else
        mv "${binary}" "${install_dir}/${BINARY_NAME}" || {
            error "Failed to move binary to ${install_dir}"
            exit 1
        }
        chmod +x "${install_dir}/${BINARY_NAME}"
    fi
}

# Check if install directory is in PATH
check_path() {
    local install_dir="$1"

    case ":$PATH:" in
        *:"$install_dir":*)
            info "Installation directory is already in PATH"
            return 0
            ;;
        *)
            warn "Installation directory is not in PATH"
            echo ""
            echo "Add the following to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
            echo ""
            echo "  export PATH=\"${install_dir}:\$PATH\""
            echo ""
            return 1
            ;;
    esac
}

# Display usage
usage() {
    cat << EOF
MoAI-ADK Installation Script

Usage:
    $0 [options]

Options:
    --version <version>    Install specific version (default: latest)
    --dir <directory>      Installation directory (default: \$HOME/.local/bin)
    -h, --help             Show this help message

Environment Variables:
    VERSION                Version to install (default: latest)
    INSTALL_DIR            Installation directory (default: \$HOME/.local/bin)

Examples:
    # Install latest version
    curl -sSL https://raw.githubusercontent.com/anthropics/moai-adk-go/main/scripts/install.sh | bash

    # Install specific version
    curl -sSL https://raw.githubusercontent.com/anthropics/moai-adk-go/main/scripts/install.sh | bash -s -- --version 1.0.0

    # Install to custom directory
    INSTALL_DIR=/usr/local/bin bash install.sh
EOF
}

# Parse command-line arguments
parse_args() {
    while [ $# -gt 0 ]; do
        case "$1" in
            --version)
                VERSION="$2"
                shift 2
                ;;
            --dir)
                INSTALL_DIR="$2"
                shift 2
                ;;
            -h|--help)
                usage
                exit 0
                ;;
            *)
                error "Unknown option: $1"
                usage
                exit 1
                ;;
        esac
    done
}

# Main installation flow
main() {
    parse_args "$@"

    info "Installing MoAI-ADK version: ${VERSION}"
    info "Installation directory: ${INSTALL_DIR}"

    # Detect platform
    platform=$(detect_platform)
    info "Detected platform: ${platform}"

    # Download binary
    download_binary "${platform}" "${VERSION}"

    # Verify binary
    binary_name="${BINARY_NAME}"
    verify_binary "${binary_name}"

    # Install binary
    install_binary "${binary_name}" "${INSTALL_DIR}"

    # Check PATH
    check_path "${INSTALL_DIR}"

    # Display success message
    echo ""
    info "MoAI-ADK installed successfully!"
    echo ""
    echo "Run the following to verify installation:"
    echo ""
    echo "  ${BINARY_NAME} version"
    echo ""
}

main "$@"
