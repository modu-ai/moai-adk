#!/bin/sh
# MoAI-ADK Installer
# Version: 1.0.0
# Website: https://adk.mo.ai.kr

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print banner
print_banner() {
    echo ""
    echo "🗿 MoAI-ADK Installer"
    echo "===================="
    echo ""
}

# Detect OS and Architecture
detect_platform() {
    OS="$(uname -s)"
    ARCH="$(uname -m)"

    case "$OS" in
        Linux*)     OS="linux";;
        Darwin*)    OS="macos";;
        MINGW*|MSYS*|CYGWIN*)
            echo "${YELLOW}⚠️  Windows detected: Please use UV installer (PowerShell)${NC}"
            echo "Run: powershell -c \"irm https://astral.sh/uv/install.ps1 | iex\""
            echo "Then: uv tool install moai-adk"
            exit 1
            ;;
        *)
            echo "${RED}❌ Unsupported OS: $OS${NC}"
            exit 1
            ;;
    esac

    case "$ARCH" in
        x86_64|amd64) ARCH="x86_64";;
        arm64|aarch64) ARCH="arm64";;
        *)
            echo "${RED}❌ Unsupported architecture: $ARCH${NC}"
            exit 1
            ;;
    esac

    echo "${GREEN}✓${NC} Platform detected: ${BLUE}$OS $ARCH${NC}"
}

# Check if UV is installed, install if not
check_install_uv() {
    if ! command -v uv &> /dev/null; then
        echo "${YELLOW}⚠️  UV not found. Installing UV...${NC}"
        curl -LsSf https://astral.sh/uv/install.sh | sh

        # Add UV to PATH for current session
        export PATH="$HOME/.local/bin:$PATH"

        # Verify UV installation
        if command -v uv &> /dev/null; then
            echo "${GREEN}✓${NC} UV installed successfully: $(uv --version)"
        else
            echo "${RED}❌ UV installation failed${NC}"
            echo "Please install UV manually: https://docs.astral.sh/uv/getting-started/installation/"
            exit 1
        fi
    else
        UV_VERSION=$(uv --version)
        echo "${GREEN}✓${NC} UV already installed: ${BLUE}$UV_VERSION${NC}"
    fi
}

# Get latest MoAI-ADK version from PyPI
get_latest_version() {
    echo "${BLUE}📦${NC} Checking latest MoAI-ADK version..."

    VERSION=$(curl -s https://pypi.org/pypi/moai-adk/json 2>/dev/null | \
              grep -o '"version":"[^"]*"' | \
              head -1 | \
              cut -d'"' -f4)

    if [ -z "$VERSION" ]; then
        echo "${YELLOW}⚠️  Could not fetch version from PyPI${NC}"
        VERSION="latest"
    else
        echo "${GREEN}✓${NC} Latest version: ${BLUE}$VERSION${NC}"
    fi
}

# Install MoAI-ADK using UV
install_moai_adk() {
    echo ""
    echo "${BLUE}🚀 Installing MoAI-ADK...${NC}"
    echo ""

    if uv tool install moai-adk; then
        echo ""
        echo "${GREEN}✅ Installation complete!${NC}"
    else
        echo ""
        echo "${RED}❌ Installation failed${NC}"
        echo "Please try manual installation:"
        echo "  uv tool install moai-adk"
        exit 1
    fi
}

# Print success message and next steps
print_success() {
    echo ""
    echo "═══════════════════════════════════════════════════════"
    echo "  ${GREEN}Welcome to MoAI-ADK!${NC} 🎉"
    echo "═══════════════════════════════════════════════════════"
    echo ""
    echo "MoAI-ADK has been installed successfully."
    echo ""
    echo "📚 Verify installation:"
    echo "      ${BLUE}moai --version${NC}"
    echo ""
    echo "🚀 Get started:"
    echo "      ${BLUE}moai init my-project${NC}"
    echo ""
    echo "📖 Documentation:"
    echo "      ${BLUE}https://adk.mo.ai.kr${NC}"
    echo ""
    echo "═══════════════════════════════════════════════════════"
}

# Check PATH configuration
check_path() {
    # Check if ~/.local/bin is in PATH
    if ! echo ":$PATH:" | grep -q ":$HOME/.local/bin:"; then
        echo ""
        echo "${YELLOW}⚠️  Note:${NC}"
        echo "Make sure ${BLUE}\$HOME/.local/bin${NC} is in your PATH."
        echo ""
        echo "Add to your ${BLUE}~/.bashrc${NC} or ${BLUE}~/.zshrc${NC}:"
        echo "  ${GREEN}export PATH=\"\$HOME/.local/bin:\$PATH\"${NC}"
        echo ""
        echo "Then restart your terminal or run:"
        echo "  ${GREEN}source ~/.bashrc${NC}  # or ${BLUE}source ~/.zshrc${NC}"
        echo ""
        return 1
    fi

    # macOS: check /etc/paths.d/ for GUI app compatibility (VS Code, Cursor)
    # Claude Code's process inherits PATH from launchd, not shell config files
    if [ "$OS" = "macos" ]; then
        LOCAL_BIN="$HOME/.local/bin"
        IN_PATHS_D=false

        # Check /etc/paths
        if [ -f /etc/paths ] && grep -q "^${LOCAL_BIN}$" /etc/paths 2>/dev/null; then
            IN_PATHS_D=true
        fi

        # Check /etc/paths.d/*
        if [ "$IN_PATHS_D" = "false" ] && [ -d /etc/paths.d ]; then
            for f in /etc/paths.d/*; do
                if [ -f "$f" ] && grep -q "^${LOCAL_BIN}$" "$f" 2>/dev/null; then
                    IN_PATHS_D=true
                    break
                fi
            done
        fi

        if [ "$IN_PATHS_D" = "false" ]; then
            echo ""
            echo "${YELLOW}⚠️  macOS System PATH Note:${NC}"
            echo "~/.local/bin is not in /etc/paths.d/ (macOS system PATH)."
            echo "GUI apps (VS Code, Cursor) may show false PATH warnings."
            echo ""
            echo "To fix, run:"
            echo "  ${GREEN}sudo sh -c 'echo \"${LOCAL_BIN}\" > /etc/paths.d/local-bin'${NC}"
            echo ""
            echo "Then restart VS Code/terminal to apply."
            echo ""
        fi
    fi

    return 0
}

# Main installation flow
main() {
    print_banner
    detect_platform
    check_install_uv
    get_latest_version
    install_moai_adk
    check_path
    print_success
}

# Run main function
main "$@"
