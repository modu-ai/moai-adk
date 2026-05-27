#!/bin/bash
# @MX:NOTE: pre-commit hook for SPEC-V3R6-LIFECYCLE-SYNC-GATE-001 M3.
# Verifies cross-file invariant between spec.md frontmatter `status:` field and
# progress.md §E.3 `status:` field. When a commit subject matches the canonical
# 4-phase close pattern, also enforces spec.md status: completed in the diff.
# On mismatch: exit 2 + JSON output (REQ-LSG-011 HARD — NEVER invokes AskUserQuestion).
# Cross-reference: .moai/specs/SPEC-V3R6-LIFECYCLE-SYNC-GATE-001/design.md §B.3
#                  .claude/rules/moai/core/agent-common-protocol.md § Hook Invocation Surface

set -u

# -----------------------------------------------------------------------------
# Input — Claude Code hook protocol stdin JSON
# -----------------------------------------------------------------------------
# Expected fields (future-extensible; missing fields tolerated via defaults):
#   staged_files: array of {path, before:{status}, after:{status}}
#   commit_subject: string (commit message first line)
#   spec_id: string (optional — derived from staged paths if absent)
#   project_root: string (optional — defaults to $CLAUDE_PROJECT_DIR or PWD)
# -----------------------------------------------------------------------------

STDIN_JSON=$(cat)

# Resolve project root for filesystem lookups
PROJECT_ROOT="${CLAUDE_PROJECT_DIR:-$(pwd)}"
if command -v jq >/dev/null 2>&1; then
  ROOT_OVERRIDE=$(printf '%s' "$STDIN_JSON" | jq -r '.project_root // empty' 2>/dev/null)
  if [ -n "$ROOT_OVERRIDE" ] && [ "$ROOT_OVERRIDE" != "null" ]; then
    PROJECT_ROOT="$ROOT_OVERRIDE"
  fi
fi

# Helper — extract field from JSON via jq; returns empty string when absent
json_get() {
  if command -v jq >/dev/null 2>&1; then
    printf '%s' "$STDIN_JSON" | jq -r "$1 // empty" 2>/dev/null
  else
    printf ''
  fi
}

# Helper — extract frontmatter status field from a markdown file
# Usage: read_status <file-path>
# Returns: status value (lowercase) or empty if not found
read_status() {
  local file="$1"
  if [ ! -f "$file" ]; then
    printf ''
    return
  fi
  # Match `status:` line in YAML frontmatter (within first 50 lines).
  # Strip surrounding quotes and trailing whitespace.
  awk '
    /^---[[:space:]]*$/ { fm++; next }
    fm == 1 && /^status:/ {
      sub(/^status:[[:space:]]*/, "")
      gsub(/^["'\'']|["'\'']$/, "")
      gsub(/[[:space:]]+$/, "")
      print
      exit
    }
    fm >= 2 { exit }
  ' "$file"
}

# Helper — emit blocking JSON to stdout and exit 2
# Usage: emit_block <stop_reason> <spec_id> <spec_md_status> <progress_md_status> <resolution_command>
emit_block() {
  local reason="$1"
  local spec_id="$2"
  local spec_status="$3"
  local prog_status="$4"
  local resolution="$5"
  # JSON construction without external dependency (deterministic ordering)
  printf '{"continue":false,"stopReason":"%s","details":{"spec_id":"%s","spec_md_status":"%s","progress_md_status":"%s","resolution_command":"%s"}}\n' \
    "$reason" "$spec_id" "$spec_status" "$prog_status" "$resolution"
  exit 2
}

# -----------------------------------------------------------------------------
# Stage 1 — derive SPEC ID and discover affected SPEC files
# -----------------------------------------------------------------------------
SPEC_ID=$(json_get '.spec_id')
COMMIT_SUBJECT=$(json_get '.commit_subject')

# Collect staged paths (one per line) — if jq unavailable or empty, fall back to git
STAGED_PATHS=$(printf '%s' "$STDIN_JSON" | jq -r '.staged_files[]?.path // empty' 2>/dev/null || true)
if [ -z "$STAGED_PATHS" ]; then
  STAGED_PATHS=$(cd "$PROJECT_ROOT" && git diff --cached --name-only 2>/dev/null || true)
fi

# Filter to spec.md / progress.md under .moai/specs/SPEC-*/
SPEC_STAGED=$(printf '%s\n' "$STAGED_PATHS" | grep -E '\.moai/specs/SPEC-[A-Z0-9-]+/(spec|progress)\.md$' || true)

# Infer SPEC_ID from staged paths if not provided
if [ -z "$SPEC_ID" ] && [ -n "$SPEC_STAGED" ]; then
  SPEC_ID=$(printf '%s\n' "$SPEC_STAGED" | head -1 | sed -E 's|.*\.moai/specs/(SPEC-[A-Z0-9-]+)/.*|\1|')
fi

# -----------------------------------------------------------------------------
# Stage 2 — fast path: no SPEC files affected → continue
# -----------------------------------------------------------------------------
if [ -z "$SPEC_STAGED" ] && [ -z "$SPEC_ID" ]; then
  printf '{"continue":true}\n'
  exit 0
fi

# -----------------------------------------------------------------------------
# Stage 3 — read current status from spec.md and progress.md on disk
# -----------------------------------------------------------------------------
SPEC_DIR="$PROJECT_ROOT/.moai/specs/$SPEC_ID"
SPEC_MD="$SPEC_DIR/spec.md"
PROGRESS_MD="$SPEC_DIR/progress.md"

SPEC_STATUS=$(read_status "$SPEC_MD")
PROGRESS_STATUS=$(read_status "$PROGRESS_MD")

# If both files unavailable or both statuses empty, skip silently (e.g. SPEC deleted)
if [ -z "$SPEC_STATUS" ] && [ -z "$PROGRESS_STATUS" ]; then
  printf '{"continue":true}\n'
  exit 0
fi

# -----------------------------------------------------------------------------
# Stage 4 — Canonical 4-phase close subject enforcement (AC-LSG-008)
# -----------------------------------------------------------------------------
# Pattern: chore(SPEC-XXX): 4-phase close — atomic
CLOSE_SUBJECT_RE='^chore\(SPEC-[A-Z0-9-]+\): 4-phase close [—-] atomic$'
if printf '%s' "$COMMIT_SUBJECT" | grep -Eq "$CLOSE_SUBJECT_RE"; then
  # Must include spec.md status: completed in the diff
  if [ "$SPEC_STATUS" != "completed" ]; then
    emit_block \
      "canonical 4-phase close subject requires spec.md status: completed in diff" \
      "$SPEC_ID" \
      "$SPEC_STATUS" \
      "$PROGRESS_STATUS" \
      "moai spec close $SPEC_ID"
  fi
fi

# -----------------------------------------------------------------------------
# Stage 5 — Cross-file status invariant (AC-LSG-003, AC-LSG-015)
# -----------------------------------------------------------------------------
# When both files declare a status, the values MUST match.
# This catches the most common drift: spec.md frontmatter status:completed
# committed without corresponding progress.md §E.3 status:completed update
# (and vice versa).
if [ -n "$SPEC_STATUS" ] && [ -n "$PROGRESS_STATUS" ] && [ "$SPEC_STATUS" != "$PROGRESS_STATUS" ]; then
  emit_block \
    "spec.md/progress.md status field mismatch" \
    "$SPEC_ID" \
    "$SPEC_STATUS" \
    "$PROGRESS_STATUS" \
    "moai spec close $SPEC_ID --backfill-only"
fi

# -----------------------------------------------------------------------------
# All checks passed
# -----------------------------------------------------------------------------
printf '{"continue":true}\n'
exit 0
