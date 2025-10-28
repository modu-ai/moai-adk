#!/bin/bash
# Initialize development repository's .moai/config.json with actual version values
# This script should be run after: pip install -e . OR uv pip install -e .
#
# Purpose: Replace template placeholders ({{MOAI_VERSION}}) with actual versions
# from pyproject.toml, ensuring the update command works correctly in dev environments

set -e  # Exit on error

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
CONFIG_FILE="$PROJECT_ROOT/.moai/config.json"
PYPROJECT_FILE="$PROJECT_ROOT/pyproject.toml"

echo "🔧 Initializing development repository configuration..."
echo ""

# Verify required files exist
if [ ! -f "$CONFIG_FILE" ]; then
    echo "❌ Error: .moai/config.json not found at $CONFIG_FILE"
    exit 1
fi

if [ ! -f "$PYPROJECT_FILE" ]; then
    echo "❌ Error: pyproject.toml not found at $PYPROJECT_FILE"
    exit 1
fi

# Extract version from pyproject.toml
VERSION=$(grep '^version = ' "$PYPROJECT_FILE" | sed 's/version = "\(.*\)"/\1/' | head -1)

if [ -z "$VERSION" ]; then
    echo "❌ Error: Could not extract version from pyproject.toml"
    exit 1
fi

echo "📌 Version extracted from pyproject.toml: $VERSION"
echo "📝 Updating .moai/config.json..."

# Use Python to safely update the JSON file
python3 << EOF
import json
from pathlib import Path

config_path = Path("$CONFIG_FILE")
config = json.loads(config_path.read_text(encoding="utf-8"))

# Update moai.version
config['moai']['version'] = "$VERSION"

# Ensure project section exists and add template_version
if 'project' not in config:
    config['project'] = {}
config['project']['template_version'] = "$VERSION"

# Write back with proper formatting
config_path.write_text(json.dumps(config, indent=2, ensure_ascii=False) + '\n', encoding="utf-8")

print(f"✅ Updated moai.version: {config['moai']['version']}")
print(f"✅ Updated project.template_version: {config['project']['template_version']}")
EOF

echo ""
echo "✅ Development configuration initialized!"
echo ""
echo "📋 Verification:"
grep -A 1 '"moai"' "$CONFIG_FILE" | head -2
echo ""
echo "✨ You can now run 'moai-adk update' and it will work correctly on the first try!"
echo ""
echo "💡 Tip: Add this to your development setup checklist after 'pip install -e .'"
