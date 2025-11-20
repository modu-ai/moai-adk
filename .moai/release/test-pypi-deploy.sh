#!/bin/bash

# TestPyPI Deployment Script for MoAI-ADK
# Local-only development tool for testing deployments

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() { echo -e "${BLUE}ℹ️  $1${NC}"; }
log_success() { echo -e "${GREEN}✅ $1${NC}"; }
log_error() { echo -e "${RED}❌ $1${NC}"; }
log_warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }

# Project configuration
PROJECT_NAME="moai-adk"
TEST_PYPI_URL="https://test.pypi.org/legacy/"

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if we're in a git repo
    if ! git rev-parse --git-head > /dev/null 2>&1; then
        log_error "Not in a git repository"
        exit 1
    fi
    
    # Check if uv is installed
    if ! command -v uv &> /dev/null; then
        log_error "uv is not installed"
        exit 1
    fi
    
    # Check if pyproject.toml exists
    if [[ ! -f "pyproject.toml" ]]; then
        log_error "pyproject.toml not found"
        exit 1
    fi
    
    # Check if TestPyPI token is set
    if [[ -z "${TEST_PYPI_TOKEN:-}" ]]; then
        log_error "TEST_PYPI_TOKEN environment variable not set"
        log_info "Set it with: export TEST_PYPI_TOKEN=your_token"
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

# Run quality checks
run_quality_checks() {
    log_info "Running quality checks..."
    
    # Run tests
    log_info "Running tests..."
    if ! uv run pytest --tb=short; then
        log_error "Tests failed"
        exit 1
    fi
    
    # Type checking
    log_info "Running type checking..."
    if ! uv run mypy src/; then
        log_error "Type checking failed"
        exit 1
    fi
    
    # Linting
    log_info "Running linting..."
    if ! uv run ruff check src/; then
        log_error "Linting failed"
        exit 1
    fi
    
    # Format checking
    log_info "Checking formatting..."
    if ! uv run black --check src/; then
        log_error "Format check failed"
        exit 1
    fi
    
    log_success "Quality checks passed"
}

# Get current version
get_current_version() {
    uv run python -c "
import tomllib
with open('pyproject.toml', 'rb') as f:
    data = tomllib.load(f)
    print(data['project']['version'])
"
}

# Clean previous builds
clean_build() {
    log_info "Cleaning previous builds..."
    rm -rf dist/ build/ *.egg-info/
    log_success "Build cleaned"
}

# Build package
build_package() {
    log_info "Building package..."
    uv build
    log_success "Package built successfully"
}

# Deploy to TestPyPI
deploy_to_test_pypi() {
    local version=$(get_current_version)
    log_info "Deploying ${PROJECT_NAME} v${version} to TestPyPI..."
    
    # Deploy using uv publish with TestPyPI
    uv publish --publish-url "$TEST_PYPI_URL" --token "$TEST_PYPI_TOKEN"
    
    log_success "Package deployed to TestPyPI successfully"
}

# Verify deployment
verify_deployment() {
    local version=$(get_current_version)
    log_info "Verifying deployment to TestPyPI..."
    
    # Wait a moment for TestPyPI to process
    sleep 5
    
    # Check if package is available on TestPyPI
    if curl -s "https://test.pypi.org/pypi/$PROJECT_NAME/$version/" > /dev/null; then
        log_success "Package verified on TestPyPI: https://test.pypi.org/project/$PROJECT_NAME/$version/"
    else
        log_warning "Package not yet visible on TestPyPI (may take a few minutes)"
        log_info "Check manually at: https://test.pypi.org/project/$PROJECT_NAME/"
    fi
}

# Main execution
main() {
    echo
    echo "🚀 MoAI-ADK TestPyPI Deployment"
    echo "================================"
    echo
    
    check_prerequisites
    run_quality_checks
    clean_build
    build_package
    deploy_to_test_pypi
    verify_deployment
    
    echo
    log_success "TestPyPI deployment completed successfully!"
    echo
    echo "📦 Package details:"
    echo "   Name: $PROJECT_NAME"
    echo "   Version: $(get_current_version)"
    echo "   TestPyPI: https://test.pypi.org/project/$PROJECT_NAME/"
    echo
    echo "🧪 To test the package:"
    echo "   pip install --index-url https://test.pypi.org/simple/ $PROJECT_NAME==$(get_current_version)"
    echo
}

# Run main function
main "$@"
