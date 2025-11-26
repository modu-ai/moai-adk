#!/bin/bash

# 버전 동기화 스크립트
# 모든 버전 파일을 지정된 버전으로 업데이트합니다

set -e

if [ -z "$1" ]; then
  echo "❌ Usage: sync-versions.sh <VERSION>"
  echo "   Example: sync-versions.sh 0.31.0"
  exit 1
fi

VERSION="$1"
VERSION_FILE="${2:-.}"

echo "🔄 Synchronizing version files to: $VERSION"
echo ""

# Validate version format
if ! [[ "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "❌ Invalid version format: $VERSION"
  echo "   Must be: MAJOR.MINOR.PATCH (e.g., 0.31.0)"
  exit 1
fi

# Update src/moai_adk/__init__.py
echo "📝 Updating src/moai_adk/__init__.py..."
if [ "$(uname)" = "Darwin" ]; then
  # macOS requires empty string for -i
  sed -i '' "s/__version__ = \".*\"/__version__ = \"$VERSION\"/" "$VERSION_FILE/src/moai_adk/__init__.py"
else
  # Linux
  sed -i "s/__version__ = \".*\"/__version__ = \"$VERSION\"/" "$VERSION_FILE/src/moai_adk/__init__.py"
fi
echo "   ✅ Updated __init__.py"

# Update src/moai_adk/version.py (fallback)
echo "📝 Updating src/moai_adk/version.py..."
if [ "$(uname)" = "Darwin" ]; then
  sed -i '' "s/MOAI_VERSION = \".*\"/MOAI_VERSION = \"$VERSION\"/" "$VERSION_FILE/src/moai_adk/version.py"
else
  sed -i "s/MOAI_VERSION = \".*\"/MOAI_VERSION = \"$VERSION\"/" "$VERSION_FILE/src/moai_adk/version.py"
fi
echo "   ✅ Updated version.py fallback"

# Verify synchronization
echo ""
echo "🔍 Verifying synchronization..."

INIT_VERSION=$(grep '__version__ = ' "$VERSION_FILE/src/moai_adk/__init__.py" | sed 's/__version__ = "//' | sed 's/"//')
VERSION_PY=$(grep 'MOAI_VERSION = ' "$VERSION_FILE/src/moai_adk/version.py" | sed 's/MOAI_VERSION = "//' | sed 's/"//')

if [ "$INIT_VERSION" != "$VERSION" ] || [ "$VERSION_PY" != "$VERSION" ]; then
  echo "❌ Version sync verification failed!"
  echo "   __init__.py: $INIT_VERSION"
  echo "   version.py:  $VERSION_PY"
  echo "   Expected:    $VERSION"
  exit 1
fi

echo "✅ Version files synchronized successfully"
echo "   src/moai_adk/__init__.py: $INIT_VERSION"
echo "   src/moai_adk/version.py:  $VERSION_PY"
echo ""

exit 0
