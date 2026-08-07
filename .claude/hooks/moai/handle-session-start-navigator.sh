#!/usr/bin/env bash
# handle-session-start-navigator.sh — SessionStart ambient auto-brief hook.
#
# Role: emit a bounded additionalContext (≤500 tokens) that surfaces the
# Navigator entry brief (current frontier + next task + link to the full file)
# into a new session's context, so a returning session re-orients to project
# state in a single read instead of re-deriving from many scattered files.
#
# Fail-open + time-boxed (per Advisory-Check Discipline):
#   * missing navigator.md      → exit 0, no output
#   * unreadable / empty file   → exit 0, no output
#   * staleness computation fail→ exit 0, no output (advisory dropped)
#   * this script itself missing→ Claude Code handles missing hooks gracefully
# This hook NEVER blocks session start (exit 0 always; advisory only).
#
# Self-contained bash: depends only on `git` + standard coreutils. Does NOT
# depend on the `moai` binary or `jq`, so it does NOT share mode A of the
# wrapper-layer shared-failure catalogue (hook-independence.md §3).
#
# Staleness signal: when the Navigator's last-regen-commit is more
# than 3 commits behind HEAD (hard-coded default N=3; overridable via
# `navigator.staleness_cycles` in .moai/config/sections/navigator.yaml), the
# additionalContext carries a staleness advisory naming the gap.
set -u

ROOT="${CLAUDE_PROJECT_DIR:-$PWD}"
NAV_FILE="$ROOT/.moai/project/navigator/navigator.md"
LAST_REGEN="$ROOT/.moai/state/navigator/last-regen-commit.txt"
NAV_CONFIG="$ROOT/.moai/config/sections/navigator.yaml"

# Fail-open: if the Navigator file is absent or unreadable, emit nothing.
if [ ! -r "$NAV_FILE" ]; then
    exit 0
fi

# --- read bounded content (≤500-token ceiling ≈ 2000 chars) -----------------
# Extract: the "## Current frontier" block + the "## Next task" block + a link.
# Cap at ~2000 chars to stay under the token ceiling.
brief_raw="$(awk '
    /^## Current frontier/      { inf=1; print; next }
    /^## Next task/             { inf=1; print; next }
    /^## /                      { inf=0; next }
    inf                         { print }
' "$NAV_FILE" 2>/dev/null | head -c 1800)"

if [ -z "$brief_raw" ]; then
    # Nothing parseable — fail open.
    exit 0
fi

# --- staleness signal (best-effort, non-fatal) -------------------------------
staleness_line=""
if [ -r "$LAST_REGEN" ]; then
    regen_sha="$(head -1 "$LAST_REGEN" 2>/dev/null | tr -dc '0-9a-f')"
    head_sha="$(git -C "$ROOT" rev-parse HEAD 2>/dev/null | tr -dc '0-9a-f')"
    if [ -n "$regen_sha" ] && [ -n "$head_sha" ] && [ "$regen_sha" != "$head_sha" ]; then
        # Count commits between regen_sha and HEAD (inclusive of HEAD, exclusive of regen).
        behind="$(git -C "$ROOT" rev-list --count "${regen_sha}..${head_sha}" 2>/dev/null || echo 0)"
        behind="${behind:-0}"
        # Threshold: default N=3, overridable via navigator.staleness_cycles.
        threshold=3
        if [ -r "$NAV_CONFIG" ]; then
            tv="$(grep -E '^[[:space:]]*staleness_cycles:' "$NAV_CONFIG" 2>/dev/null | head -1 | sed -E 's/^[^:]*:[[:space:]]*//; s/[^0-9].*$//')"
            case "$tv" in
                ''|*[!0-9]*) ;;  # keep default
                *) threshold="$tv" ;;
            esac
        fi
        if [ "$behind" -gt "$threshold" ]; then
            staleness_line="Stale: Navigator is ${behind} commits behind HEAD (threshold ${threshold}) — recent work may be missing."
        fi
    fi
fi

# --- compose + emit (≤500 tokens) -------------------------------------------
# Truncate the brief body to keep the total under the ceiling.
max_body=1600
if [ ${#brief_raw} -gt "$max_body" ]; then
    brief_raw="${brief_raw:0:$max_body} …"
fi

# Build the additionalContext string. Newlines are preserved inside the JSON
# string via manual escaping.
ac="Project Navigator brief (auto, SessionStart):
${brief_raw}

Full file: .moai/project/navigator/navigator.md"
if [ -n "$staleness_line" ]; then
    ac="${ac}

⚠️ ${staleness_line}"
fi

# JSON-escape the additionalContext (newline → \n, quote → \", backslash → \\).
ac_escaped="$(printf '%s' "$ac" | awk '
    BEGIN { ORS="" }
    { gsub(/\\/, "\\\\"); gsub(/"/, "\\\""); print; if (NR < 9999) printf "\\n" }
')"

# Emit the Claude Code SessionStart hookSpecificOutput contract.
printf '{"hookSpecificOutput":{"hookEventName":"SessionStart","additionalContext":"%s"}}\n' "$ac_escaped"

exit 0
