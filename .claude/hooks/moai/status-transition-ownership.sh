#!/bin/bash
# Hook: status-transition-ownership
# Purpose: Verify Write/Edit invoker matches Status Transition Ownership Matrix
# Trigger: PostToolUse event when tool ∈ {Write, Edit, MultiEdit} on SPEC artifact files
# Cross-reference: .claude/rules/moai/development/spec-frontmatter-schema.md (Status Transition Ownership Matrix)

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
    # jq absent — hook degrades to no-op. stdout intentionally empty (PostToolUse schema).
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
        # Not a SPEC artifact — allow without inspection. stdout intentionally empty (PostToolUse schema).
        exit 0
        ;;
esac

# Verify the tool is a write operation
case "$TOOL_NAME" in
    Write|Edit|MultiEdit) ;;
    *)
        # Non-write tool — allow. stdout intentionally empty (PostToolUse schema).
        exit 0
        ;;
esac

# Extract status: from new content (Write) or post-edit state
# Status Transition Ownership Matrix (canonical reference, 3-phase close — plan/run/sync):
#   * → draft       : manager-spec
#   draft → in-progress : manager-develop (first run-phase commit; frontmatter status+updated only)
#   in-progress → implemented → completed : manager-docs (single sync commit carries the completed
#                       transition — there is no separate "Mx chore commit"; MX Tag validation is a
#                       sync sub-step, NOT a separate phase)
#   * → superseded  : manager-spec (when authoring superseding SPEC)
#   * → archived    : manager-docs
#   * → rejected    : orchestrator (recorded by manager-docs)

# Read current status from file on disk (post-edit state for Edit/MultiEdit; pre-write state for Write).
# [[:space:]] (not \s): BSD sed/grep on macOS do not support the \s escape —
# with \s the pattern silently fails to strip, leaving stray whitespace in the log.
if [ -f "$FILE_PATH" ]; then
    CURRENT_STATUS=$(grep -E '^status:[[:space:]]*' "$FILE_PATH" 2>/dev/null | head -1 | sed 's/^status:[[:space:]]*//;s/[[:space:]]*$//')
else
    CURRENT_STATUS="<file absent — Write creating new>"
fi

# Advisory hook (never blocks; exit 2 reserved for future ownership-mismatch enforcement).
# stdout intentionally empty: PostToolUse accepts only empty / {} / {"systemMessage": "..."} /
# {"hookSpecificOutput": {...}} on stdout. A custom {"hook":...,"decision":"allow",...} object
# failed Claude Code JSON-schema validation on every Write|Edit (validation error noise with no
# functional effect — the file write still completed, and the audit log below captured the
# transition site). Agent-name attribution via tool_input is not directly available in the
# Claude Code hook payload; future enhancement: integrate with SubagentStop for correlation.

# Log for audit trail
mkdir -p "${CLAUDE_PROJECT_DIR:-$PWD}/.moai/logs"
echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) [status-transition-ownership] $TOOL_NAME $FILE_PATH status=$CURRENT_STATUS" \
    >> "${CLAUDE_PROJECT_DIR:-$PWD}/.moai/logs/status-transition-audit.log"

exit 0
