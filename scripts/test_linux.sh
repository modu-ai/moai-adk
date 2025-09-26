#!/bin/bash
# MoAI-ADK Linux Cross-Platform Compatibility Test
# @TASK:CROSS-PLATFORM-001 Linux environment test automation

set -e  # Exit on any error

# @TASK:VERSION-SYNC-002 Dynamic version extraction from _version.py
get_current_version() {
    python3 -c "
import sys; sys.path.insert(0, 'src')
from moai_adk._version import __version__
print(__version__)
" 2>/dev/null || echo "0.1.22"
}

CURRENT_VERSION=$(get_current_version)

echo ""
echo "===================================================================="
echo "🗿 MoAI-ADK Linux Cross-Platform Compatibility Test"
echo "===================================================================="
echo ""

# Check Python installation
echo "Checking Python installation..."
if ! command -v python3 &> /dev/null; then
    echo "❌ Python3 not found! Please install Python 3.11+"
    exit 1
fi

python3 --version

# Check pip installation
echo "Checking pip installation..."
if ! command -v pip3 &> /dev/null && ! python3 -m pip --version &> /dev/null; then
    echo "❌ pip not found! Please install pip"
    exit 1
fi

# Use python3 -m pip for better compatibility
python3 -m pip --version

# Install MoAI-ADK from built package
echo ""
echo "Installing MoAI-ADK from local package..."
if [ -f "dist/moai_adk-${CURRENT_VERSION}-py3-none-any.whl" ]; then
    python3 -m pip install dist/moai_adk-${CURRENT_VERSION}-py3-none-any.whl --upgrade
else
    echo "⚠️ 빌드 파일 moai_adk-${CURRENT_VERSION}-py3-none-any.whl이 없습니다."
    echo "먼저 빌드를 실행하세요: make build"
    exit 1
fi

# Test CLI commands
echo ""
echo "Testing CLI commands..."
echo ""

echo "Testing version command..."
moai --version

echo ""
echo "Testing help command..."
moai --help

echo ""
echo "Testing doctor command..."
moai doctor

# Test Python tools
echo ""
echo "Testing Python tools..."
echo ""

echo "Testing version manager..."
python3 scripts/version_manager.py status

echo ""
echo "Testing test runner..."
python3 scripts/test_runner.py --help

echo ""
echo "Testing build system..."
python3 scripts/build.py --help

# Run comprehensive cross-platform test
echo ""
echo "Running comprehensive cross-platform test..."
python3 scripts/cross_platform_test.py

echo ""
echo "===================================================================="
echo "✅ All Linux compatibility tests passed!"
echo "✅ Linux is fully supported by MoAI-ADK!"
echo "===================================================================="
echo ""