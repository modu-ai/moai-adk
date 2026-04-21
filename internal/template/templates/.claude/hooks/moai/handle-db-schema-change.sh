#!/bin/bash
# MoAI DB Schema Change Hook — PostToolUse handler (SPEC-DB-SYNC-001)
# Detects migration file changes and triggers db-schema-sync processing.
INPUT=$(cat)
FILE_PATH=$(echo "$INPUT" | jq -r '.tool_input.file_path // empty')
[ -z "$FILE_PATH" ] && exit 0
if command -v moai &> /dev/null; then
  moai hook db-schema-sync --file "$FILE_PATH" 2>>/tmp/moai-db-sync.log
  exit 0
fi
if [ -f "$HOME/go/bin/moai" ]; then
  "$HOME/go/bin/moai" hook db-schema-sync --file "$FILE_PATH" 2>>/tmp/moai-db-sync.log
  exit 0
fi
if [ -f "$HOME/.local/bin/moai" ]; then
  "$HOME/.local/bin/moai" hook db-schema-sync --file "$FILE_PATH" 2>>/tmp/moai-db-sync.log
  exit 0
fi
exit 0
