#!/bin/bash
# Hook: status-transition-ownership
# Purpose: Verify Write/Edit invoker matches Status Transition Ownership Matrix
# Trigger: PostToolUse event when tool ∈ {Write, Edit, MultiEdit} on SPEC artifact files
# Origin: SPEC-V3R6-AGENT-TEAM-REBUILD-001 M4 (2026-05-25)
# REQs: REQ-ATR-009 (hook architecture); Status Transition Ownership Matrix per
#       .claude/rules/moai/development/spec-frontmatter-schema.md

set -e

# Opt-out flag
if [ "$1" = "--skip-hook" ]; then
    echo "{\"skipped\": true, \"reason\": \"--skip-hook flag\"}" >&2
    mkdir -p "${CLAUDE_PROJECT_DIR:-$PWD}/.moai/logs"
    echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) [status-transition-ownership] skipped via --skip-hook" \
        >> "${CLAUDE_PROJECT_DIR:-$PWD}/.moai/logs/hook-skip.log"
    exit 0
fi

# Graceful degradation: jq is required for JSON parsing
if ! command -v jq >/dev/null 2>&1; then
    echo "{\"hook\":\"status-transition-ownership\",\"decision\":\"allow\",\"warning\":\"jq absent — hook degraded to no-op\"}"
    exit 0
fi

# Read stdin JSON from Claude Code
INPUT=$(cat)
TOOL_NAME=$(echo "$INPUT" | jq -r '.tool_name // ""')
FILE_PATH=$(echo "$INPUT" | jq -r '.tool_input.file_path // .tool_input.path // ""')

# Only inspect SPEC artifact files
case "$FILE_PATH" in
    *.moai/specs/SPEC-*/spec.md|*.moai/specs/SPEC-*/plan.md|*.moai/specs/SPEC-*/acceptance.md|*.moai/specs/SPEC-*/design.md|*.moai/specs/SPEC-*/research.md)
        ;;
    *)
        # Not a SPEC artifact — allow without inspection
        echo "{\"hook\":\"status-transition-ownership\",\"decision\":\"allow\",\"reason\":\"not a SPEC artifact\"}"
        exit 0
        ;;
esac

# Verify the tool is a write operation
case "$TOOL_NAME" in
    Write|Edit|MultiEdit) ;;
    *)
        echo "{\"hook\":\"status-transition-ownership\",\"decision\":\"allow\",\"reason\":\"non-write tool\"}"
        exit 0
        ;;
esac

# Extract status: from new content (Write) or post-edit state
# Status Transition Ownership Matrix (canonical reference):
#   * → draft       : manager-spec
#   draft → in-progress : manager-develop (first M-commit only; frontmatter status+updated only)
#   in-progress → implemented : manager-docs (frontmatter status+updated only)
#   implemented → completed : manager-docs OR orchestrator (Mx chore)
#   * → superseded  : manager-spec (when authoring superseding SPEC)
#   * → archived    : manager-docs
#   * → rejected    : orchestrator (recorded by manager-docs)

# Read current status from file on disk (post-edit state for Edit/MultiEdit; pre-write state for Write)
if [ -f "$FILE_PATH" ]; then
    CURRENT_STATUS=$(grep -E '^status:\s*' "$FILE_PATH" 2>/dev/null | head -1 | sed 's/^status:\s*//;s/\s*$//')
else
    CURRENT_STATUS="<file absent — Write creating new>"
fi

# Diagnostic output (advisory; never blocks)
# Note: agent name attribution via tool_input is not directly available in Claude Code hook payload;
# this hook is currently advisory and logs the transition site. Future enhancement: integrate with
# SubagentStop hook for agent-name correlation.
cat <<EOF
{
  "hook": "status-transition-ownership",
  "decision": "allow",
  "file_path": "$FILE_PATH",
  "tool_name": "$TOOL_NAME",
  "current_status": "$CURRENT_STATUS",
  "advisory": "Verify the modifying agent matches Status Transition Ownership Matrix at .claude/rules/moai/development/spec-frontmatter-schema.md § Status Transition Ownership Matrix"
}
EOF

# Log for audit trail
mkdir -p "${CLAUDE_PROJECT_DIR:-$PWD}/.moai/logs"
echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) [status-transition-ownership] $TOOL_NAME $FILE_PATH status=$CURRENT_STATUS" \
    >> "${CLAUDE_PROJECT_DIR:-$PWD}/.moai/logs/status-transition-audit.log"

exit 0
