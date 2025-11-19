#!/bin/bash
# Pre-commit hook - Agent permissionMode 검증

echo "🔍 Checking agent permissionMode values..."

# 변경된 .md 파일 중 agents/ 디렉토리만 검사
changed_agents=$(git diff --cached --name-only | grep '\.claude/agents/.*\.md$')

if [ -z "$changed_agents" ]; then
  exit 0
fi

invalid_found=false

for file in $changed_agents; do
  if [ -f "$file" ]; then
    # invalid permissionMode 검사
    if grep -qE '^permissionMode:\s*(auto|ask)\s*$' "$file"; then
      echo "❌ Invalid permissionMode in: $file"
      invalid_found=true
    fi
  fi
done

if [ "$invalid_found" = true ]; then
  echo ""
  echo "⚠️  Invalid permissionMode values detected!"
  echo "Valid options: acceptEdits, bypassPermissions, default, dontAsk, plan"
  echo ""
  echo "Run fix script:"
  echo "  uv run .moai/scripts/fix-agent-permissions.py"
  echo ""
  exit 1
fi

echo "✅ All agent permissionMode values are valid"
exit 0
