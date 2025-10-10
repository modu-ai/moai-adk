#!/bin/bash
# @CODE:BUILD-HOOKS-001 | SPEC: SPEC-BUILD-001.md
# MoAI-ADK Hooks Build Script
# Builds TypeScript hooks to CommonJS for Claude Code

set -e  # Exit on error

echo "🔨 Building MoAI-ADK Hooks..."

# Check if node_modules exists
if [ ! -d "node_modules" ]; then
  echo "📦 Installing dependencies..."
  npm install
fi

# Build hooks
echo "⚡ Running tsup..."
npm run build:hooks

# Verify output
if [ -f ".claude-plugin/hooks/scripts/session-notice.cjs" ]; then
  echo "✅ Build successful!"
  echo ""
  echo "📄 Generated files:"
  ls -lh .claude-plugin/hooks/scripts/*.cjs
else
  echo "❌ Build failed - output files not found"
  exit 1
fi
