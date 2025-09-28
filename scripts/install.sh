#!/bin/bash
# MoAI-ADK Unix Installation Script (macOS/Linux)
# @TASK:UNIX-INSTALL-001
#
# This script automatically installs uv and MoAI-ADK on Unix systems.
# Usage: curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/scripts/install.sh | bash

set -euo pipefail

# Configuration
SKIP_UV_INSTALL=false
VERSION="latest"
VERBOSE=false

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --skip-uv)
            SKIP_UV_INSTALL=true
            shift
            ;;
        --version)
            VERSION="$2"
            shift 2
            ;;
        --verbose)
            VERBOSE=true
            shift
            ;;
        -h|--help)
            echo "MoAI-ADK Installation Script"
            echo ""
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --skip-uv      Skip uv installation"
            echo "  --version VER  Install specific version (default: latest)"
            echo "  --verbose      Enable verbose output"
            echo "  -h, --help     Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

log() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

debug() {
    if [[ "$VERBOSE" == "true" ]]; then
        echo -e "${BLUE}[DEBUG]${NC} $1"
    fi
}

command_exists() {
    command -v "$1" >/dev/null 2>&1
}

detect_os() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        echo "macOS"
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        echo "Linux"
    else
        echo "Unknown"
    fi
}

main() {
    log "🗿 MoAI-ADK Installation Script"
    log "============================="

    OS=$(detect_os)
    debug "Detected OS: $OS"

    # Step 1: Install uv if not present
    if [[ "$SKIP_UV_INSTALL" == "false" ]]; then
        if command_exists uv; then
            log "✅ uv is already installed"
            debug "uv version: $(uv --version 2>/dev/null || echo 'unknown')"
        else
            log "📦 Installing uv (Astral Python package manager)..."
            if curl -LsSf https://astral.sh/uv/install.sh | sh; then
                log "✅ uv installed successfully"
                # Source the shell configuration to make uv available
                if [[ -f "$HOME/.cargo/env" ]]; then
                    source "$HOME/.cargo/env"
                fi
                # Add to PATH for current session
                export PATH="$HOME/.cargo/bin:$PATH"
            else
                error "Failed to install uv"
                error "Please install uv manually from https://astral.sh/uv/"
                exit 1
            fi
        fi
    fi

    # Verify uv is available
    if ! command_exists uv; then
        error "uv is not available in PATH"
        warn "You may need to restart your shell or run: source ~/.cargo/env"
        exit 1
    fi

    # Step 2: Install/run MoAI-ADK
    log "🗿 Installing MoAI-ADK..."

    if [[ "$VERSION" == "latest" ]]; then
        # Install latest version from PyPI
        if uv tool install moai-adk; then
            log "✅ MoAI-ADK installed successfully"
        else
            warn "Failed to install MoAI-ADK with uv tool install"
            warn "Trying alternative installation with uvx..."

            if uvx --from moai-adk moai-adk doctor >/dev/null 2>&1; then
                log "✅ MoAI-ADK is working via uvx"
                log "💡 You can use 'uvx --from moai-adk moai-adk [command]' to run commands"
            else
                error "Both installation methods failed"
                exit 1
            fi
        fi
    else
        # Install specific version
        if uv tool install "moai-adk==$VERSION"; then
            log "✅ MoAI-ADK $VERSION installed successfully"
        else
            error "Failed to install MoAI-ADK version $VERSION"
            exit 1
        fi
    fi

    # Step 3: Verify installation
    log "🔍 Verifying installation..."

    # Check if uv tool installation worked
    if command_exists moai-adk; then
        VERSION_OUTPUT=$(moai-adk --version 2>/dev/null || echo "unknown")
        log "✅ MoAI-ADK is ready: $VERSION_OUTPUT"
    else
        # Try uvx fallback
        if uvx --from moai-adk moai-adk --version >/dev/null 2>&1; then
            VERSION_OUTPUT=$(uvx --from moai-adk moai-adk --version)
            log "✅ MoAI-ADK is ready via uvx: $VERSION_OUTPUT"
        else
            warn "Installation completed but verification failed"
            warn "Please try running 'moai-adk --version' or 'uvx --from moai-adk moai-adk --version'"
        fi
    fi

    # Step 4: Show next steps
    echo ""
    log "🎉 Installation completed successfully!"
    echo ""
    echo -e "${BLUE}Next steps:${NC}"
    echo -e "${BLUE}1. Create a new project: moai-adk init my-project${NC}"
    echo -e "${BLUE}2. Or navigate to existing project and run: moai-adk init${NC}"
    echo -e "${BLUE}3. For help: moai-adk --help${NC}"
    echo ""

    if ! command_exists moai-adk; then
        echo -e "${YELLOW}💡 If 'moai-adk' command is not found, use:${NC}"
        echo -e "${YELLOW}   uvx --from moai-adk moai-adk [command]${NC}"
        echo ""
    fi

    log "🗿 Welcome to Spec-First TDD development with MoAI-ADK!"
    log "💡 For Windows users: Download moai-adk.exe from GitHub releases for offline usage"
}

# Trap to handle interruption
trap 'error "Installation interrupted by user"; exit 130' INT

# Run main function
main "$@"