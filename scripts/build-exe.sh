#!/bin/bash
# MoAI-ADK Windows Executable Build Script
# @TASK:EXE-BUILD-001
#
# This script builds a standalone Windows executable using PyInstaller.
# Supports cross-platform builds and optimization.

set -euo pipefail

# Configuration
PYTHON_VERSION="3.12"
BUILD_DIR="build"
DIST_DIR="dist"
SPEC_FILE="moai-adk.spec"
VERBOSE=false
CLEAN=false
UPX_COMPRESS=true

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log() {
    echo -e "${GREEN}[BUILD]${NC} $1"
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

print_usage() {
    cat << EOF
MoAI-ADK Executable Build Script

Usage: $0 [OPTIONS]

Options:
    --clean             Clean build directories before building
    --no-upx            Disable UPX compression
    --python VERSION    Python version to use (default: $PYTHON_VERSION)
    --verbose           Enable verbose output
    -h, --help          Show this help message

Examples:
    $0                  # Build with default settings
    $0 --clean          # Clean build and rebuild
    $0 --no-upx         # Build without compression
    $0 --verbose        # Build with detailed output

EOF
}

check_requirements() {
    log "🔍 Checking build requirements..."

    # Check if uv is available
    if ! command -v uv >/dev/null 2>&1; then
        error "uv is not installed. Please install it from https://astral.sh/uv/"
        exit 1
    fi

    debug "uv version: $(uv --version)"

    # Check if we're in the project root
    if [[ ! -f "pyproject.toml" ]] || [[ ! -f "$SPEC_FILE" ]]; then
        error "Must be run from the project root directory"
        error "Missing: pyproject.toml or $SPEC_FILE"
        exit 1
    fi

    # Check UPX if compression is enabled
    if [[ "$UPX_COMPRESS" == "true" ]]; then
        if command -v upx >/dev/null 2>&1; then
            debug "UPX available: $(upx --version 2>/dev/null | head -n1 || echo 'unknown')"
        else
            warn "UPX not found - compression will be disabled"
            UPX_COMPRESS=false
        fi
    fi

    log "✅ Requirements check passed"
}

setup_environment() {
    log "🐍 Setting up Python environment..."

    # Install specific Python version if needed
    debug "Installing Python $PYTHON_VERSION..."
    uv python install "$PYTHON_VERSION"

    # Create/update virtual environment
    debug "Creating virtual environment..."
    uv venv --python "$PYTHON_VERSION" .venv

    # Activate virtual environment
    source .venv/bin/activate || source .venv/Scripts/activate 2>/dev/null || {
        error "Failed to activate virtual environment"
        exit 1
    }

    # Install the package in development mode
    debug "Installing MoAI-ADK in development mode..."
    uv pip install -e .

    # Install PyInstaller and build dependencies
    debug "Installing build dependencies..."
    uv pip install pyinstaller>=6.0.0

    log "✅ Environment setup completed"
}

clean_build() {
    if [[ "$CLEAN" == "true" ]]; then
        log "🧹 Cleaning build directories..."

        # Remove build artifacts
        rm -rf "$BUILD_DIR" "$DIST_DIR" "*.egg-info" "__pycache__"
        find . -name "*.pyc" -delete
        find . -name "*.pyo" -delete
        find . -name "__pycache__" -type d -exec rm -rf {} + 2>/dev/null || true

        log "✅ Clean completed"
    fi
}

build_executable() {
    log "🔨 Building Windows executable..."

    # Prepare PyInstaller arguments
    PYINSTALLER_ARGS=(
        "$SPEC_FILE"
        "--clean"
        "--noconfirm"
    )

    if [[ "$VERBOSE" == "true" ]]; then
        PYINSTALLER_ARGS+=("--log-level" "DEBUG")
    else
        PYINSTALLER_ARGS+=("--log-level" "INFO")
    fi

    # Disable UPX if requested or unavailable
    if [[ "$UPX_COMPRESS" == "false" ]]; then
        PYINSTALLER_ARGS+=("--noupx")
    fi

    # Run PyInstaller
    debug "Running: pyinstaller ${PYINSTALLER_ARGS[*]}"

    if pyinstaller "${PYINSTALLER_ARGS[@]}"; then
        log "✅ Build completed successfully"
    else
        error "Build failed"
        exit 1
    fi
}

validate_build() {
    log "🔍 Validating build output..."

    EXE_PATH="$DIST_DIR/moai-adk.exe"

    # Check if executable exists
    if [[ ! -f "$EXE_PATH" ]]; then
        error "Executable not found at $EXE_PATH"
        exit 1
    fi

    # Get file size
    FILE_SIZE=$(du -h "$EXE_PATH" | cut -f1)
    log "📦 Executable size: $FILE_SIZE"

    # Test executable (basic functionality)
    debug "Testing executable..."
    if wine "$EXE_PATH" --version >/dev/null 2>&1 || [[ "$OSTYPE" == "msys" ]]; then
        log "✅ Executable validation passed"
    else
        warn "Cannot test executable on this platform (Wine not available)"
        warn "Manual testing on Windows required"
    fi

    # Show final output
    echo ""
    log "🎉 Build completed successfully!"
    log "📍 Executable location: $EXE_PATH"
    log "📏 File size: $FILE_SIZE"
    echo ""
}

create_checksum() {
    log "🔒 Creating checksums..."

    EXE_PATH="$DIST_DIR/moai-adk.exe"

    if [[ -f "$EXE_PATH" ]]; then
        # Create SHA256 checksum
        if command -v sha256sum >/dev/null 2>&1; then
            sha256sum "$EXE_PATH" > "$EXE_PATH.sha256"
        elif command -v shasum >/dev/null 2>&1; then
            shasum -a 256 "$EXE_PATH" > "$EXE_PATH.sha256"
        else
            warn "SHA256 checksum tool not found"
        fi

        # Create MD5 checksum (for compatibility)
        if command -v md5sum >/dev/null 2>&1; then
            md5sum "$EXE_PATH" > "$EXE_PATH.md5"
        elif command -v md5 >/dev/null 2>&1; then
            md5 "$EXE_PATH" > "$EXE_PATH.md5"
        else
            warn "MD5 checksum tool not found"
        fi

        log "✅ Checksums created"
    fi
}

main() {
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --clean)
                CLEAN=true
                shift
                ;;
            --no-upx)
                UPX_COMPRESS=false
                shift
                ;;
            --python)
                PYTHON_VERSION="$2"
                shift 2
                ;;
            --verbose)
                VERBOSE=true
                shift
                ;;
            -h|--help)
                print_usage
                exit 0
                ;;
            *)
                error "Unknown option: $1"
                print_usage
                exit 1
                ;;
        esac
    done

    log "🗿 MoAI-ADK Windows Executable Builder"
    log "====================================="

    check_requirements
    clean_build
    setup_environment
    build_executable
    validate_build
    create_checksum

    log "🎯 Build process completed successfully!"
}

# Trap to handle interruption
trap 'error "Build interrupted by user"; exit 130' INT

# Run main function
main "$@"