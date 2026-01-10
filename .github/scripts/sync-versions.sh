#!/bin/bash

# Version Synchronization Script
# Updates all version files to the specified version
# Single Source of Truth: pyproject.toml

set -e

if [ -z "$1" ]; then
  echo "❌ Usage: sync-versions.sh <VERSION>"
  echo "   Example: sync-versions.sh 0.41.2"
  exit 1
fi

VERSION="$1"
VERSION_FILE="${2:-.}"

echo "🔄 Synchronizing version files to: $VERSION"
echo ""

# Validate version format
if ! [[ "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "❌ Invalid version format: $VERSION"
  echo "   Must be: MAJOR.MINOR.PATCH (e.g., 0.41.2)"
  exit 1
fi

# Helper function for sed (macOS vs Linux compatibility)
sed_inplace() {
  if [ "$(uname)" = "Darwin" ]; then
    sed -i '' "$1" "$2"
  else
    sed -i "$1" "$2"
  fi
}

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 1. Python Package Files
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
echo "📦 Updating Python package files..."

# Update src/moai_adk/__init__.py (if version is hardcoded)
if [ -f "$VERSION_FILE/src/moai_adk/__init__.py" ]; then
  if grep -q '__version__ = "' "$VERSION_FILE/src/moai_adk/__init__.py"; then
    sed_inplace "s/__version__ = \".*\"/__version__ = \"$VERSION\"/" "$VERSION_FILE/src/moai_adk/__init__.py"
    echo "   ✅ Updated __init__.py"
  fi
fi

# Update src/moai_adk/version.py (fallback version)
if [ -f "$VERSION_FILE/src/moai_adk/version.py" ]; then
  sed_inplace "s/_FALLBACK_VERSION = \".*\"/_FALLBACK_VERSION = \"$VERSION\"/" "$VERSION_FILE/src/moai_adk/version.py"
  echo "   ✅ Updated version.py (_FALLBACK_VERSION)"
fi

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 2. Configuration Files (Local .moai/)
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
echo "📝 Updating local configuration files..."

# Update .moai/config/config.yaml (only moai.version, not template_version)
if [ -f "$VERSION_FILE/.moai/config/config.yaml" ]; then
  # Use awk to update only moai.version, not template_version or other version fields
  awk -v ver="$VERSION" '
    /^moai:/ { in_moai=1 }
    /^[a-z]/ && !/^moai:/ { in_moai=0 }
    in_moai && /^  version:/ { sub(/version: "[^"]*"/, "version: \"" ver "\"") }
    { print }
  ' "$VERSION_FILE/.moai/config/config.yaml" > "$VERSION_FILE/.moai/config/config.yaml.tmp"
  mv "$VERSION_FILE/.moai/config/config.yaml.tmp" "$VERSION_FILE/.moai/config/config.yaml"
  echo "   ✅ Updated .moai/config/config.yaml (moai.version)"
fi

# Update .moai/config/sections/system.yaml
if [ -f "$VERSION_FILE/.moai/config/sections/system.yaml" ]; then
  sed_inplace "s/version: \"[0-9]*\.[0-9]*\.[0-9]*\"/version: \"$VERSION\"/" "$VERSION_FILE/.moai/config/sections/system.yaml"
  echo "   ✅ Updated .moai/config/sections/system.yaml"
fi

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 3. Template Configuration Files
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
echo "📄 Checking template configuration files..."

# Templates should use {{MOAI_VERSION}} placeholder, not hardcoded versions
# Only warn if hardcoded version found

TEMPLATE_CONFIG="$VERSION_FILE/src/moai_adk/templates/.moai/config/config.yaml"
if [ -f "$TEMPLATE_CONFIG" ]; then
  if grep -q "version: \"[0-9]*\.[0-9]*\.[0-9]*\"" "$TEMPLATE_CONFIG"; then
    echo "   ⚠️  Warning: Template config.yaml has hardcoded version (should use {{MOAI_VERSION}})"
  else
    echo "   ✅ Template config.yaml uses placeholder"
  fi
fi

TEMPLATE_SYSTEM="$VERSION_FILE/src/moai_adk/templates/.moai/config/sections/system.yaml"
if [ -f "$TEMPLATE_SYSTEM" ]; then
  if grep -q "version: \"[0-9]*\.[0-9]*\.[0-9]*\"" "$TEMPLATE_SYSTEM"; then
    echo "   ⚠️  Warning: Template system.yaml has hardcoded version (should use {{MOAI_VERSION}})"
  else
    echo "   ✅ Template system.yaml uses placeholder"
  fi
fi

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 4. Documentation Files (README versions)
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
echo "📚 Checking documentation files..."

# Note: README versions are typically updated manually or through release notes
# This section can be expanded to auto-update version badges if needed

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 5. Verification
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
echo ""
echo "🔍 Verifying synchronization..."

ERRORS=0

# Check version.py
if [ -f "$VERSION_FILE/src/moai_adk/version.py" ]; then
  VERSION_PY=$(grep '_FALLBACK_VERSION = ' "$VERSION_FILE/src/moai_adk/version.py" | sed 's/_FALLBACK_VERSION = "//' | sed 's/"//')
  if [ "$VERSION_PY" != "$VERSION" ]; then
    echo "   ❌ version.py mismatch: $VERSION_PY (expected: $VERSION)"
    ERRORS=$((ERRORS + 1))
  else
    echo "   ✅ version.py: $VERSION_PY"
  fi
fi

# Check local config.yaml (moai.version only)
if [ -f "$VERSION_FILE/.moai/config/config.yaml" ]; then
  # Extract moai.version specifically (not template_version or other versions)
  CONFIG_VER=$(awk '/^moai:/{found=1} found && /^  version:/{match($0, /[0-9]+\.[0-9]+\.[0-9]+/); print substr($0, RSTART, RLENGTH); exit}' "$VERSION_FILE/.moai/config/config.yaml")
  if [ "$CONFIG_VER" != "$VERSION" ]; then
    echo "   ❌ config.yaml moai.version mismatch: $CONFIG_VER (expected: $VERSION)"
    ERRORS=$((ERRORS + 1))
  else
    echo "   ✅ config.yaml moai.version: $CONFIG_VER"
  fi
fi

# Check local system.yaml
if [ -f "$VERSION_FILE/.moai/config/sections/system.yaml" ]; then
  SYSTEM_VER=$(grep 'version:' "$VERSION_FILE/.moai/config/sections/system.yaml" | head -1 | sed 's/.*version: "//' | sed 's/".*//')
  if [ "$SYSTEM_VER" != "$VERSION" ]; then
    echo "   ❌ system.yaml mismatch: $SYSTEM_VER (expected: $VERSION)"
    ERRORS=$((ERRORS + 1))
  else
    echo "   ✅ system.yaml: $SYSTEM_VER"
  fi
fi

echo ""

if [ $ERRORS -gt 0 ]; then
  echo "❌ Version sync verification failed with $ERRORS error(s)!"
  exit 1
fi

echo "✅ Version files synchronized successfully to $VERSION"
echo ""
echo "📋 Summary:"
echo "   - Python package version updated"
echo "   - Local configuration files updated"
echo "   - Template files checked for placeholders"
echo ""
echo "💡 Next steps:"
echo "   1. Review changes: git diff"
echo "   2. Commit changes: git add -A && git commit -m 'chore: sync versions to $VERSION'"
echo "   3. Create tag: git tag v$VERSION"
echo ""

exit 0
