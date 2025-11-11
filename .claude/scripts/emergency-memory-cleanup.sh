#!/bin/bash
# Emergency Memory Cleanup Script
# Use this when memory usage gets critical

echo "🚨 Emergency Memory Cleanup Started..."

# Kill MCP server processes
echo "🧹 Killing MCP servers..."
pkill -f "context7-mcp" 2>/dev/null || true
pkill -f "mcp-server-playwright" 2>/dev/null || true  
pkill -f "mcp-server-sequential-thinking" 2>/dev/null || true

# Kill Playwright processes
echo "🧹 Killing Playwright browsers..."
pkill -f "headless_shell" 2>/dev/null || true
pkill -f "chrome-mac" 2>/dev/null || true

# Clean npm cache
echo "🧹 Cleaning npm cache..."
npm cache clean --force 2>/dev/null || true
rm -rf ~/.npm/_npx/* 2>/dev/null || true

# Python memory manager backup
if [ -f "$(dirname "$0")/mcp-memory-manager.py" ]; then
    echo "🧹 Running Python memory manager..."
    python3 "$(dirname "$0")/mcp-memory-manager.py" cleanup
fi

echo "✅ Emergency cleanup completed!"
echo "💾 Memory usage after cleanup:"
python3 -c "import psutil; m=psutil.virtual_memory(); print(f'  {m.percent:.1f}% used, {m.available//1024//1024}MB free')"
