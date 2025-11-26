#!/bin/bash

# 버전 파일 동기화 검증 스크립트
# 모든 버전 파일이 일치하는지 확인합니다

set -e

VERSION_FILE="${1:-.}"

echo "🔍 Validating version file synchronization..."
echo ""

# Extract versions from each file
PYPROJECT_VERSION=$(grep '^version = ' "$VERSION_FILE/pyproject.toml" 2>/dev/null | sed 's/version = "//' | sed 's/"//')
INIT_VERSION=$(grep '__version__ = ' "$VERSION_FILE/src/moai_adk/__init__.py" 2>/dev/null | sed 's/__version__ = "//' | sed 's/"//')
VERSION_PY=$(grep 'MOAI_VERSION = ' "$VERSION_FILE/src/moai_adk/version.py" 2>/dev/null | sed 's/MOAI_VERSION = "//' | sed 's/"//')

# Display versions
echo "📋 Current versions:"
echo "  📦 pyproject.toml:        $PYPROJECT_VERSION"
echo "  🐍 __init__.py:           $INIT_VERSION"
echo "  🔧 version.py (fallback): $VERSION_PY"
echo ""

# Check if all versions match
if [ -z "$PYPROJECT_VERSION" ]; then
  echo "❌ ERROR: Could not read version from pyproject.toml"
  exit 1
fi

if [ -z "$INIT_VERSION" ]; then
  echo "❌ ERROR: Could not read __version__ from __init__.py"
  exit 1
fi

if [ -z "$VERSION_PY" ]; then
  echo "❌ ERROR: Could not read MOAI_VERSION from version.py"
  exit 1
fi

# Validate synchronization
SYNC_ERROR=0

if [ "$PYPROJECT_VERSION" != "$INIT_VERSION" ]; then
  echo "❌ VERSION MISMATCH: pyproject.toml ($PYPROJECT_VERSION) ≠ __init__.py ($INIT_VERSION)"
  SYNC_ERROR=1
fi

if [ "$PYPROJECT_VERSION" != "$VERSION_PY" ]; then
  echo "❌ VERSION MISMATCH: pyproject.toml ($PYPROJECT_VERSION) ≠ version.py ($VERSION_PY)"
  SYNC_ERROR=1
fi

if [ "$INIT_VERSION" != "$VERSION_PY" ]; then
  echo "❌ VERSION MISMATCH: __init__.py ($INIT_VERSION) ≠ version.py ($VERSION_PY)"
  SYNC_ERROR=1
fi

if [ $SYNC_ERROR -eq 1 ]; then
  echo ""
  echo "❌ Version synchronization FAILED!"
  echo "   All version files must be identical"
  exit 1
fi

echo "✅ All version files are synchronized!"
echo "   Version: $PYPROJECT_VERSION"
echo ""
exit 0
