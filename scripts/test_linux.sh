#!/bin/bash
# MoAI-ADK Linux Cross-Platform Compatibility Test
# @TASK:CROSS-PLATFORM-001 Linux environment test automation

set -e  # Exit on any error

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
python3 -m pip install dist/moai_adk-0.1.9-py3-none-any.whl --upgrade

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