#!/bin/bash
# gateguard-fact-force.sh — PreToolUse hook (first-edit investigation advisory)
#
# @MX:ANCHOR High fan-in advisory — every edit passes through this hook
#
# Emits a one-time ADVISORY notice on the FIRST Edit/Write/MultiEdit on each file
# path per session, recommending investigation (importers, data schemas, user
# instruction). The edit is ALLOWED to proceed (exit 0) — this hook never blocks.
# Subsequent edits to the same path in the same session produce no advisory. A
# prior Read on the same path pre-populates the fact state, so the first post-Read
# Edit skips the advisory (Read-as-investigation). Advisory opt-out via
# MOAI_FACT_FORCE=off. Shell-only, O(1), fail-open, self-terminates < 5s, jq-free.
#
# State: ${CLAUDE_PROJECT_DIR:-$PWD}/.moai/state/fact-force/<hash>
# keyed by SHA-1(session_id + absolute_file_path), 0o600, single JSON line.
# The hook MUST NOT invoke any user-prompting mechanism (subagent boundary, C-HRA-008).
#
# fact-force-skip.log (.moai/logs/): intentionally WRITE-ONLY — a local audit
# trail of MOAI_FACT_FORCE=off opt-outs (timestamp, session, path), with no
# mechanical consumer; pruned at the operator's discretion
# (SPEC-OBSERVE-HYGIENE-001 REQ-OBH-003).

# Fail-open wrapper: any unexpected error → exit 0 (allow). This hook NEVER
# exits 2 — it is advisory-only (exit 0 on every path).
set +e

# --- 1. Read payload (cap 1MB to avoid truncation mid-JSON) ---
payload=$(head -c 1048576)

# --- 2. Extract tool_name; self-loop prevention (REQ-FF-010) ---
tool_name=$(printf '%s' "$payload" \
    | grep -oE '"tool_name"[[:space:]]*:[[:space:]]*"[^"]*"' \
    | head -n1 \
    | sed -E 's/^"tool_name"[[:space:]]*:[[:space:]]*"//; s/"$//')

case "$tool_name" in
    Read) read_mode=1 ;;                  # Read → pre-populate fact state, allow (no block)
    Edit|Write|MultiEdit) read_mode=0 ;;  # gated tools — proceed to block logic
    *) exit 0 ;;                          # non-gated tool (Bash, or absent) → allow
esac

# --- 3. Advisory opt-out (REQ-FF-004) ---
if [ "${MOAI_FACT_FORCE:-}" = "off" ]; then
    project_dir="${CLAUDE_PROJECT_DIR:-$PWD}"
    skip_log="$project_dir/.moai/logs/fact-force-skip.log"
    ts=$(date -u '+%Y-%m-%dT%H:%M:%SZ' 2>/dev/null || echo '')
    sess=$(printf '%s' "$payload" \
        | grep -oE '"session_id"[[:space:]]*:[[:space:]]*"[^"]*"' \
        | head -n1 \
        | sed -E 's/^"session_id"[[:space:]]*:[[:space:]]*"//; s/"$//')
    fpath=$(printf '%s' "$payload" \
        | grep -oE '"file_path"[[:space:]]*:[[:space:]]*"[^"]*"' \
        | head -n1 \
        | sed -E 's/^"file_path"[[:space:]]*:[[:space:]]*"//; s/"$//')
    mkdir -p "$(dirname "$skip_log")" 2>/dev/null
    printf '%s\t%s\t%s\n' "$ts" "${sess:-none}" "${fpath:-none}" >> "$skip_log" 2>/dev/null
    exit 0
fi

# --- 4. Extract session_id; fail-open if absent (edge E4) ---
session_id=$(printf '%s' "$payload" \
    | grep -oE '"session_id"[[:space:]]*:[[:space:]]*"[^"]*"' \
    | head -n1 \
    | sed -E 's/^"session_id"[[:space:]]*:[[:space:]]*"//; s/"$//')
[ -z "$session_id" ] && exit 0

# --- 5. Extract file_path; fail-open if empty (edge E1) ---
file_path=$(printf '%s' "$payload" \
    | grep -oE '"file_path"[[:space:]]*:[[:space:]]*"[^"]*"' \
    | head -n1 \
    | sed -E 's/^"file_path"[[:space:]]*:[[:space:]]*"//; s/"$//')
[ -z "$file_path" ] && exit 0

# --- 6. Resolve relative path to absolute (edge E5) ---
case "$file_path" in
    /*) ;;                            # already absolute
    *) file_path="$PWD/$file_path" ;; # relative → prefix cwd
esac

# --- 7. Compute composite hash key (portable shasum/sha1sum, REQ-FF-003) ---
if command -v shasum >/dev/null 2>&1; then
    key=$(printf '%s|%s' "$session_id" "$file_path" | shasum | awk '{print $1}')
elif command -v sha1sum >/dev/null 2>&1; then
    key=$(printf '%s|%s' "$session_id" "$file_path" | sha1sum | awk '{print $1}')
else
    # Fallback: sanitized path (no hashing) — longer filename but no external dep.
    key=$(printf '%s|%s' "$session_id" "$file_path" | sed 's|/|_|g; s|:|-|g; s| |_|g')
fi
[ -z "$key" ] && exit 0

# --- 8. Locate state directory + state file (REQ-FF-003) ---
project_dir="${CLAUDE_PROJECT_DIR:-$PWD}"
state_dir="$project_dir/.moai/state/fact-force"
state_file="$state_dir/$key"

# --- 9. Second-or-subsequent edit → allow (REQ-FF-002) ---
if [ -f "$state_file" ]; then
    exit 0
fi

# --- 10. First edit: write state file (0o600, key=value lines — no JSON interpolation) ---
mkdir -p "$state_dir" 2>/dev/null || exit 0
ts=$(date -u '+%Y-%m-%dT%H:%M:%SZ' 2>/dev/null || echo '')
# umask 077 → created file is 0o600 regardless of touch/cat path (REQ-FF-012).
# `via` records whether the state was seeded by a Read (pre-populate) or a
# blocked Edit (the original first-edit gate path).
# State file uses plain key=value lines (NOT JSON) so a `"` or `\` in
# session_id / path / tool_name cannot break the record (the gate only checks
# file existence via [ -f "$state_file" ], never parses the content).
( umask 077; printf 'session_id=%s\npath=%s\nfirst_seen=%s\nvia=%s\n' "$session_id" "$file_path" "$ts" "$tool_name" > "$state_file" ) 2>/dev/null || exit 0

# Read mode: fact state pre-populated; allow the read with no block. The next
# Edit/Write/MultiEdit on this path will skip the gate because the state file
# now exists (REQ-FF-002 second-or-subsequent-edit allow + Read-as-investigation).
if [ "$read_mode" = "1" ]; then
    exit 0
fi

# --- 11. Emit advisory systemMessage + allow (REQ-FA-001) ---
# stdout JSON systemMessage: Claude Code renders this as informational context
# (NOT a red error box). exit 0 = allow. This hook NEVER blocks.
guidance=$(cat <<GUIDANCE
First-edit advisory on $file_path.

Before proceeding, investigate:
  1. IMPORTERS — who imports / depends on this file?
       grep -rn "<file-basename>" --include='*.go' --include='*.ts' --include='*.py' .  (adapt to language)
  2. DATA SCHEMAS — what data structures / contracts / API types does this file touch?
       Read the struct / interface / type definitions and their consumers.
  3. USER INSTRUCTION — what user instruction justifies this edit?
       Re-read the SPEC acceptance criteria or the explicit user request.

This is a one-time advisory per (session, file path). Your NEXT edit to this path
will not produce this notice. To disable for the session: MOAI_FACT_FORCE=off
GUIDANCE
)

# JSON-escape via awk (jq-free, per §C.5 NFR): backslash to double-backslash,
# double-quote to backslash-quote, inter-line newline to literal backslash-n.
escaped=$(printf '%s' "$guidance" | awk '
BEGIN { sep = "" }
{ gsub(/\\/, "\\\\"); gsub(/"/, "\\\""); printf "%s%s", sep, $0; sep = "\\n" }
')
printf '{"systemMessage":"%s"}\n' "$escaped"
exit 0
